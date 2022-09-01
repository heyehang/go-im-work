package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/heyehang/go-im-pkg/pulsarsdk"
	"go-im-work/internal/worker"
	"go-im-work/pulsar"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go-im-work/internal/config"
	"go-im-work/internal/handler"
	"go-im-work/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/work.yaml", "the config file")

func main() {
	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c)
	pulsar.Init(c)
	defer pulsarsdk.Closed()
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()
	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)
	wg := new(sync.WaitGroup)
	go func() {
		wg.Add(1)
		defer wg.Done()
		server.Start()
	}()
	w := worker.NewWorker(c)
	go func() {
		wg.Add(1)
		defer wg.Done()
		w.Start(context.Background())
	}()
	sig := make(chan os.Signal, 1)
	//syscall.SIGINT 线上记得加上这个信号 ctrl + c
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM)
	for {
		s := <-sig
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			wg.Wait()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)

}
