package worker

import (
	"context"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/bytedance/sonic"
	"github.com/heyehang/go-im-grpc/im_server"
	"github.com/heyehang/go-im-pkg/pulsarsdk"
	"github.com/panjf2000/ants/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"go-im-work/internal/config"
	"time"
)

type Worker struct {
	IMSrvCli im_server.IMServerClient
	pool     *ants.Pool
	conf     config.Config
}

func NewWorker(conf config.Config) *Worker {
	w := new(Worker)
	w.IMSrvCli = im_server.NewIMServerClient(zrpc.MustNewClient(conf.IMServer).Conn())
	pool, err := ants.NewPool(conf.WorkPoolSize, ants.WithNonblocking(true))
	if err != nil {
		panic(err)
		return nil
	}
	w.pool = pool
	w.conf = conf
	return w
}

type MsgReq struct {
	DeviceTokens []string `json:"device_tokens,omitempty"`
	// 消息体
	Msg []byte `protobuf:"json:"msg,omitempty"`
}

func (w *Worker) Start(ctx context.Context) {
	pulsarsdk.SubscribeMsg(ctx, w.conf.Pulsar.Topic, w.conf.Pulsar.SubscriptionName, func(message pulsar.Message, err error) {
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
			ctx1, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			_, err = w.IMSrvCli.SnedMsg(ctx1, in)
			if err != nil {
				logx.Errorf("SubscribeMsg_SnedMsg_err :%+v", err)
				return
			}
		})
	})
}
