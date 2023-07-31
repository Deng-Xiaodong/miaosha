package main

import (
	"encoding/json"
	"log"
	"miaosha/common"
	"miaosha/rabbitmq"
	"miaosha/redislock"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {

	//初始化配置
	var configFile string
	if len(os.Args) > 1 {
		configFile = os.Args[1]
		log.Printf("配置文件为：%s\n", configFile)
	}
	config := common.InitConfig(configFile)

	//初始化redis客户端连接
	redislock.InitRedisClient(config.RedCfg)

	//设置redis required keys if not exist
	redislock.SetKeys(redislock.Keys{Key: config.LimitCfg.LimitKey, Value: strconv.FormatInt(config.LimitCfg.Burst, 10)}, redislock.Keys{Key: config.LimitCfg.LimitKey + ":last", Value: strconv.FormatInt(time.Now().Unix(), 10)},
		redislock.Keys{Key: "inventory", Value: strconv.FormatInt(config.LimitCfg.Inventory, 10)})

	//初始化RabbitMQ连接
	rabbitmq.SetURL(config.MqCfg.MqUrl)
	mq := rabbitmq.NewSimpleRabbitMQ(config.MqCfg.QueName)

	//初始化限流器
	limit := redislock.NewLimit(config.LimitCfg.LimitKey, config.LimitCfg.Rate, config.LimitCfg.Burst)

	//http服务
	http.HandleFunc("/getone", func(w http.ResponseWriter, r *http.Request) {
		//Ip限流
		redislock.RedisClient.SetNX(r.RemoteAddr, 0, 10*time.Second)
		redislock.RedisClient.Incr(r.RemoteAddr)
		cnt, _ := redislock.RedisClient.Get(r.RemoteAddr).Int()
		if cnt > 5 {
			rsp, _ := json.Marshal(common.Error{Code: 500, Msg: "请求过于频繁，网络拥堵"})
			_, _ = w.Write(rsp)
			return
		}
		//拿到令牌才能被服务
		if limit.Allow() {
			log.Printf("%v    %s get access", time.Now(), r.RemoteAddr)
			if redislock.GetOne() {
				uid := r.FormValue("user_id")
				pid := r.FormValue("product_id")
				msg, _ := json.Marshal(rabbitmq.Message{UserId: uid, ProdId: pid})

				mq.Publish(msg)
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
	})
	log.Fatalln(http.ListenAndServe(":"+config.HttpCfg.Port, nil))
}
