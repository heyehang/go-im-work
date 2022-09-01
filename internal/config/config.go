package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	IMServer     zrpc.RpcClientConf
	WorkPoolSize int `json:",optional"`
	Pulsar       *Pulsar
}

type Pulsar struct {
	Url string `json:",optional"`
	// 单位秒
	ConnectionTimeout int `json:",optional"`
	// 单位秒
	OperationTimeout int `json:",optional"`
	// 最大连接数
	MaxConnectionsPerBroker int `json:",optional"`
	// 订阅topic
	Topic            string `json:",optional"`
	SubscriptionName string `json:",optional"`
}
