package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/heyehang/go-im-pkg/pulsarsdk"
	"github.com/heyehang/go-im-pkg/tlog"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"go-im-work/internal/worker"
	"go-im-work/pulsar"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/heyehang/go-im-pkg/etcdtool"
	"github.com/pyroscope-io/client/pyroscope"
	"github.com/zeromicro/go-zero/rest"
	"go-im-work/internal/config"
	"go-im-work/internal/handler"
	"go-im-work/internal/svc"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	profile    *pyroscope.Profiler
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
	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)
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

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if profile != nil {
				_ = profile.Stop()
			}
		}()
		startPyroscope()
	}()

	logx.Info("listen on http port ", fmt.Sprintf("addr: %s:%d", c.Host, c.Port))
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

func startPyroscope() {
	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)
	var err error
	profile, err = pyroscope.Start(pyroscope.Config{
		ApplicationName: "go-im-server",
		// replace this with the address of pyroscope server
		ServerAddress: "http://172.16.0.15:4040",
		// you can disable logging by setting this to nil
		Logger: pyroscope.StandardLogger,
		// optionally, if authentication is enabled, specify the API key:
		// AuthToken: os.Getenv("PYROSCOPE_AUTH_TOKEN"),
		ProfileTypes: []pyroscope.ProfileType{
			// these profile types are enabled by default:
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
			// these profile types are optional:
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
		SampleRate: 200,
	})
	if err != nil {
		panic(err)
		return
	}
}
