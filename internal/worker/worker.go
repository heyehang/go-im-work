package worker

import (
	"context"
	"go-im-work/internal/config"
	"sync"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/bytedance/sonic"

	"github.com/heyehang/go-im-grpc/im_server"
	"github.com/heyehang/go-im-pkg/pulsarsdk"
	"github.com/panjf2000/ants/v2"
	"github.com/zeromicro/go-zero/core/logx"
)

type Worker struct {
	pool     *ants.Pool
	conf     config.Config
	IMSrvMap map[string]im_server.ImClient
	RWMutex  sync.RWMutex
}

func NewWorker(conf config.Config) *Worker {
	w := new(Worker)
	pool, err := ants.NewPool(conf.WorkPoolSize, ants.WithNonblocking(true))
	if err != nil {
		panic(err)
		return nil
	}
	w.pool = pool
	w.conf = conf
	w.IMSrvMap = make(map[string]im_server.ImClient, 10)
	err = w.watchIMServer()
	if err != nil {
		panic(err)
	}
	return w
}

type MsgReq struct {
	ServerKey    string   `json:"server_key"`
	DeviceTokens []string `json:"device_tokens,omitempty"`
	// 消息体
	Msg []byte `protobuf:"json:"msg,omitempty"`
}

func (w *Worker) Start(ctx context.Context) {
	pulsarsdk.SubscribeMsg(ctx, w.conf.Pulsar.Topic, func(message pulsar.Message, err error) {
		if err != nil {
			logx.Errorf("SubscribeMsg_err :%+v", err)
			return
		}
		// 消息体
		data := message.Payload()
		msg := MsgReq{}
		err = sonic.Unmarshal(data, &msg)
		if err != nil {
			logx.Errorf("SubscribeMsg_Unmarshal_err :%+v", err)
			return
		}
		err = w.pool.Submit(func() {
			in := new(im_server.SnedMsgReq)
			in.Msg = msg.Msg
			in.DeviceTokens = msg.DeviceTokens
			// 查找到节点了，定向推
			cli, ok := w.IMSrvMap[msg.ServerKey]
			if ok && cli != nil {
				_, err = cli.SnedMsg(context.Background(), in)
				if err != nil {
					logx.Errorf("SubscribeMsg_SnedMsg_err :%+v", err)
					return
				}
				return
			}
			// 没有查找到，广播所有节点
			for _, imsrv := range w.IMSrvMap {
				_, err = imsrv.SnedMsg(context.Background(), in)
				if err != nil {
					logx.Errorf("SubscribeMsg_SnedMsg_err :%+v", err)
					return
				}
			}
		})
	})
}
