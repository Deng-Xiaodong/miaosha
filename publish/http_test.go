package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"testing"
	"time"
)

type Server struct {
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/getone":
		log.Println("执行业务中")
		time.Sleep(5 * time.Second)
		w.Write([]byte("just test"))
	}

}
func TestName(t *testing.T) {

	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		//ip := strings.Split(r.Header.Get("X-Real-IP"), ":")[0]
		//log.Printf("ip:%s", ip)
		log.Printf("URI:%v\nPATH:%v\n", r.RequestURI, r.URL.Path)
		w.Write([]byte(r.URL.Path))

	})
	lis, _ := net.Listen("tcp", ":81")
	t1 := lis.(*net.TCPListener)
	t1.File()
	//go func() {
	//	time.Sleep(1 * time.Minute)
	//	lis.Close()
	//}()
	log.Fatalf("http服务错误：%v\n", http.Serve(lis, nil))
}
func TestShutdown(t *testing.T) {

	s := &http.Server{
		Handler: &Server{},
		Addr:    ":82",
	}
	idleConnClosed := make(chan struct{})
	workAgainChan := make(chan struct{})
	slight := make(chan os.Signal, 1)

	//负责安全关闭
	go func() {
		for {

			signal.Notify(slight, os.Interrupt)
			select {
			case <-slight:
				if err := s.Shutdown(context.Background()); err != nil {
					log.Printf("HTTP server Shutdown: %v\n", err)
				}
				idleConnClosed <- struct{}{}
			}

		}

	}()

	//监控内部资源
	go func() {

		//time.Sleep(30 * time.Second)
		tk := time.NewTicker(30 * time.Second)
		for {
			select {
			case <-tk.C:
				if monitor() {
					log.Println("资源不足，通知先停止服务")
					slight <- os.Interrupt //通知shutdown协程停止服务

					<-idleConnClosed //接收服务已正常停止完毕信号
					log.Println("已停止完成")
					log.Println("通知可以继续开启服务")
					workAgainChan <- struct{}{} //通知主协程继续开放http服务

				}
			}
		}

	}()
	//pauseChan := make(chan struct{})
	cnt := 1
	for {
		log.Printf("第%d次服务\n", cnt)
		listen, _ := net.Listen("tcp", ":83")
		if err := s.Serve(listen); err != http.ErrServerClosed {
			break
		}
		_ = listen.Close()
		<-workAgainChan
		cnt++

	}

}
func monitor() bool {
	//percent, _ := cpu.Percent(time.Second, false)
	//return percent[0]>0.7
	return true
}
