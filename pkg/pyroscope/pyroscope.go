package pyroscope

import (
	"context"
	"github.com/heyehang/go-im-pkg/pyroscopesdk"
	"github.com/pyroscope-io/client/pyroscope"
	"sync"
)

func Start(ctx context.Context, wg *sync.WaitGroup, appName, serverAddr string, logger pyroscope.Logger, openBlockProfile bool) {
	pyroscopesdk.Start(ctx, wg, appName, serverAddr, logger, true)
}

func Closed() {
	pyroscopesdk.Closed()
}
