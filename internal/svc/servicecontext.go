package svc

import (
	"github.com/heyehang/go-im-grpc/im_server"
	"github.com/zeromicro/go-zero/zrpc"
	"go-im-work/internal/config"
)

type ServiceContext struct {
	Config   config.Config
	IMSrvCli im_server.ImClient
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	ctx := new(ServiceContext)
	ctx.Config = c
	cli, err := zrpc.NewClient(c.IMServer)
	if err != nil {
		return nil, err
	}
	ctx.IMSrvCli = im_server.NewImClient(cli.Conn())
	return ctx, nil
}
