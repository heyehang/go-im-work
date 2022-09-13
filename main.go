package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/heyehang/go-im-pkg/etcdtool"
	"github.com/heyehang/go-im-pkg/pulsarsdk"
	"github.com/heyehang/go-im-pkg/tlog"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"go-im-work/internal/config"
	"go-im-work/internal/handler"
	"go-im-work/internal/svc"
	"go-im-work/internal/worker"
	"go-im-work/pkg/pulsar"
	"go-im-work/pkg/pyroscope"
	clientv3 "go.etcd.io/etcd/client/v3"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	configFile = flag.String("f", "etc/work.yaml", "the config file")
)

func main() {
	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c)
	logx.MustSetup(c.Log)
	fileWriter := logx.Reset()
	writer, err := tlog.NewMultiWriter(fileWriter)
	logx.Must(err)
	logx.SetWriter(writer)
	pulsar.Init(c)
	defer pulsarsdk.Closed()
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()
	serverCtx, err := svc.NewServiceContext(c)
	if err != nil {
		panic(err)
	}
	handler.RegisterHandlers(server, serverCtx)
	etcdcli, err := clientv3.New(clientv3.Config{
		Endpoints: c.IMServer.Etcd.Hosts,
	})
	if err != nil {
		panic(err)
	}
	etcdtool.InitEtcd(etcdcli)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.Start()
	}()
	w := worker.NewWorker(c)
	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx1, cancel := context.WithCancel(context.Background())
		defer cancel()
		w.Start(ctx1)
	}()
	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pyroscope.Start(cancelCtx, wg, c.Name, c.PyroscopeAddr, nil, true)
	logx.Info("listen on http port ", fmt.Sprintf("addr: %s:%d", c.Host, c.Port))
	sig := make(chan os.Signal, 1)
	//syscall.SIGINT 线上记得加上这个信号 ctrl + c
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM)
	for {
		s := <-sig
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			pyroscope.Closed()
			wg.Wait()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
}
