package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"miaosha/common"
	"miaosha/rabbitmq"
	"miaosha/redislock"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const script = `
local limit=tonumber(ARGV[2])
redis.call('SET',KEYS[1],0,'NX','PX',ARGV[1])
local cur=tonumber(redis.call('GET',KEYS[1]) or 0)
if cur>=limit then
	return false
else
	cur=cur+1
	redis.call('SET',KEYS[1],cur)
	return true
end
`

var (
	configFile string
	upgrade    bool
	ln         net.Listener
	server     *http.Server
)

func init() {
	flag.BoolVar(&upgrade, "upgrade", false, "user can't use this")
	flag.StringVar(&configFile, "config_file", "", "set config argv for redis and rabbitMQ")
}

func main() {
	flag.Parse()
	fmt.Println("start-up at ", time.Now(), upgrade)
	var (
		config *common.Config
		rmq    *rabbitmq.RabbitMQ
		limit  *redislock.Limit
		dl     *redislock.DisLock
	)
	//初始化配置

	config = common.InitConfig(configFile)

	//初始化redis客户端连接并设置redis required keys if not exist
	redc := redis.NewClient(&redis.Options{
		Addr:     config.RedCfg.Address + ":" + config.RedCfg.Port,
		Password: config.RedCfg.Password,
	})
	redislock.SetKeys(redc, redislock.Keys{Key: config.LimitCfg.LimitKey, Value: strconv.FormatInt(config.LimitCfg.Burst, 10)}, redislock.Keys{Key: config.LimitCfg.LimitKey + ":last", Value: strconv.FormatInt(time.Now().Unix(), 10)},
		redislock.Keys{Key: "inventory", Value: strconv.FormatInt(config.LimitCfg.Inventory, 10)})

	//初始化限流器
	limit = redislock.NewLimit(redc, config.LimitCfg.LimitKey, config.LimitCfg.Rate, config.LimitCfg.Burst)
	//初始化分布式锁
	dl = redislock.NewDisLock(redc)

	//初始化RabbitMQ连接
	rmq = rabbitmq.NewSimpleRabbitMQ(config.MqCfg.QueName, config.MqCfg.MqUrl)

	http.HandleFunc("/getone", phrase(limit, dl, rmq))
	server = &http.Server{Addr: ":8000"}

	var err error
	if upgrade {
		fd := os.NewFile(3, "")
		ln, err = net.FileListener(fd)
		if err != nil {
			fmt.Printf("fileListener fail, error: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("graceful-reborn  %v %v  %#v \n", fd.Fd(), fd.Name(), ln)
		fd.Close()
	} else {
		ln, err = net.Listen("tcp", server.Addr)
		if err != nil {
			fmt.Printf("listen %s fail, error: %s\n", server.Addr, err)
			os.Exit(1)
		}
		tcp, _ := ln.(*net.TCPListener)
		fd, _ := tcp.File()
		fmt.Printf("first-boot  %v %v %#v \n ", fd.Fd(), fd.Name(), ln)
	}
	go func() {
		//http服务
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Fatalln(err)
		}
	}()
	setupSignal()
	fmt.Println("over")

}
func phrase(limit *redislock.Limit, dl *redislock.DisLock, rmq *rabbitmq.RabbitMQ) http.HandlerFunc {
	//Ip限流

	return func(w http.ResponseWriter, r *http.Request) {
		m, n := 10, 2
		ip := r.Header.Get("X-Real-IP")
		ipBlock := redis.NewScript(script)
		if ok, _ := ipBlock.Run(limit.RedisClient, []string{ip}, m, n).Bool(); !ok {
			rsp, _ := json.Marshal(common.Error{Code: 500, Msg: "请勿频繁访问，小心加入黑名单"})
			_, _ = w.Write(rsp)
			return
		}
		//拿到令牌才能被服务
		if limit.Allow() {
			log.Printf("%v    %s get access", time.Now(), r.RemoteAddr)
			if dl.GetOne() {
				uid := r.FormValue("user_id")
				pid := r.FormValue("product_id")
				msg, _ := json.Marshal(rabbitmq.Message{UserId: uid, ProdId: pid})

				rmq.Publish(msg)
				rsp, _ := json.Marshal(common.Error{Code: 200, Msg: "抢购成功"})
				_, _ = w.Write(rsp)
				return
			} else {
				rsp, _ := json.Marshal(common.Error{Code: 200, Msg: "库存不足"})
				_, _ = w.Write(rsp)
				return
			}
		}
		rsp, _ := json.Marshal(common.Error{Code: 500, Msg: "网络繁忙"})
		_, _ = w.Write(rsp)
	}

}
func setupSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGUSR2, syscall.SIGINT, syscall.SIGTERM)
	sig := <-ch
	switch sig {
	case syscall.SIGUSR2:
		log.Println("signal cause fork")
		err := forkProcess()
		if err != nil {
			fmt.Printf("fork process error: %s\n", err)
		}
		err = server.Shutdown(context.Background())
		if err != nil {
			fmt.Printf("shutdown after forking process error: %s\n", err)
		}
	case syscall.SIGINT, syscall.SIGTERM:
		log.Println("signal cause stop")
		signal.Stop(ch)
		close(ch)
		err := server.Shutdown(context.Background())
		if err != nil {
			fmt.Printf("shutdown error: %s\n", err)
		}
	}
}

func forkProcess() error {
	flags := []string{"-upgrade", "-config_file=" + configFile}
	cmd := exec.Command(os.Args[0], flags...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	l, _ := ln.(*net.TCPListener)
	lfd, err := l.File()
	if err != nil {
		return err
	}
	cmd.ExtraFiles = []*os.File{lfd}
	return cmd.Start()
}
