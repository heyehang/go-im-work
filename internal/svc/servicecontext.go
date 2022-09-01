package svc

import (
	"github.com/heyehang/go-im-grpc/im_server"
	"github.com/zeromicro/go-zero/zrpc"
	"go-im-work/internal/config"
)

type ServiceContext struct {
	Config   config.Config
	IMSrvCli im_server.IMServerClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:   c,
		IMSrvCli: im_server.NewIMServerClient(zrpc.MustNewClient(c.IMServer).Conn()),
	}
}
