package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"time"
)

var (
	server   *http.Server
	listener net.Listener
	graceful = flag.Bool("graceful", false, "listen on fd open 3 (internal use only)")
	ch       = make(chan os.Signal, 1)
)

func handler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(20 * time.Second)
	w.Write([]byte("hello world233333!!!!"))
}

func main() {
	flag.Parse()

	http.HandleFunc("/hello", handler)
	server = &http.Server{Addr: ":9999"}

	var err error
	if *graceful {
		log.Print("main: Listening to existing file descriptor 3.")
		// cmd.ExtraFiles: If non-nil, entry i becomes file descriptor 3+i.
		// when we put socket FD at the first entry, it will always be 3(0+3)
		//为什么是3呢，而不是1 0 或者其他数字？这是因为父进程里给了个fd给子进程了 而子进程里0，1，2是预留给 标准输入、输出和错误的，所以父进程给的第一个fd在子进程里顺序排就是从3开始了；如果fork的时候cmd.ExtraFiles给了两个文件句柄，那么子进程里还可以用4开始，就看你开了几个子进程自增就行。因为我这里就开一个子进程所以把3写死了。l, err = net.FileListener(f)这一步只是把 fd描述符包装进TCPListener这个结构体。

		f := os.NewFile(3, "")
		//先复制fd到新的fd, 然后设置子进程exec时自动关闭父进程的fd,即“F_DUPFD_CLOEXEC”
		listener, err = net.FileListener(f)
	} else {
		log.Print("main: Listening on a new file descriptor.")
		listener, err = net.Listen("tcp", server.Addr)
	}

	if err != nil {
		log.Fatalf("listener error: %v", err)
	}

	go func() {
		// server.Shutdown() stops Serve() immediately, thus server.Serve() should not be in main goroutine
		err = server.Serve(listener)
		log.Printf("server.Serve err: %v\n", err)
	}()
	go func() {
		time.Sleep(1 * time.Minute)
		ch <- os.Interrupt
	}()
	signalHandler()
	log.Printf("signal end")
}

func reload() error {
	tl, ok := listener.(*net.TCPListener)
	if !ok {
		return errors.New("listener is not tcp listener")
	}

	f, err := tl.File()
	if err != nil {
		return err
	}

	args := []string{"-graceful"}
	cmd := exec.Command(os.Args[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// put socket FD at the first entry
	cmd.ExtraFiles = []*os.File{f}
	return cmd.Start()
}

func signalHandler() {

	signal.Notify(ch, os.Interrupt)
	for {
		sig := <-ch
		log.Printf("signal: %v", sig)

		// timeout context for shutdown
		ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
		switch sig {
		//case syscall.SIGINT, syscall.SIGTERM:
		//	// stop
		//	log.Printf("stop")
		//	signal.Stop(ch)
		//	server.Shutdown(ctx)
		//	log.Printf("graceful shutdown")
		//	return
		case os.Interrupt:
			// reload
			log.Printf("reload")
			err := reload()
			if err != nil {
				log.Fatalf("graceful restart error: %v", err)
			}
			server.Shutdown(ctx)
			log.Printf("graceful reload")
			return
		}
	}
}
