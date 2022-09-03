package worker

import (
	"context"

	"github.com/heyehang/go-im-grpc/im_server"
	"github.com/heyehang/go-im-pkg/etcdtool"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

func (w *Worker) watchIMServer() error {
	err := etcdtool.GetEtcdTool().WatchPrefix(w.conf.IMServer.Etcd.Key, context.Background(), func(status, key, value string) {
		logx.Slowf("watchIMServer status:%s, key:%s,val:%s ", status, key, value)
		w.RWMutex.RLock()
		defer w.RWMutex.RUnlock()
		switch status {
		case etcdtool.EtcdKeyCreate, etcdtool.EtcdKeyModify:
			w.IMSrvMap[key] = im_server.NewImClient(zrpc.MustNewClient(zrpc.RpcClientConf{
				Endpoints: []string{value},
			}).Conn())
		case etcdtool.EtcdKeyDelete:
			delete(w.IMSrvMap, key)
		default:
			logx.Error("watchIMServer not found case, status:%s, key:%s,val:%s ", status, key, value)
		}
		logx.Slowf("watchIMServer status:%s, key:%s,val:%s  success", status, key, value)
	})
	return err

}
