// This file starts the net-flux compatible TCP collection server.

package collection

import (
	"context"
	"net"

	"github.com/dellinger2023/net-flux/pkg/network"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability/cachecap"
)

// tcpServerRunner abstracts the concrete net-flux server for startup tests.
type tcpServerRunner interface {
	Run(ctx context.Context) error
}

// newTCPServer creates the net-flux TCP server. Tests may replace it with a fake runner.
var newTCPServer = func(addr string, handler network.EventHandler) tcpServerRunner {
	return network.NewTcpServer(addr, handler, nil)
}

// runServer starts one TCP collection server in a background goroutine.
func runServer(ctx context.Context, cfg *Config, cacheSvc cachecap.Service) error {
	if cfg == nil {
		return gerror.New("media collection server config cannot be nil")
	}
	if cacheSvc == nil {
		return gerror.New("media collection server requires host cache service")
	}

	var listenConfig net.ListenConfig
	listener, err := listenConfig.Listen(ctx, "tcp", cfg.Addr)
	if err != nil {
		return gerror.Wrap(err, "listen media collection server failed")
	}
	if err := listener.Close(); err != nil {
		return gerror.Wrap(err, "close media collection preflight listener failed")
	}

	discovery := newDiscoveryRuntime(cfg.Discovery)
	server := newTCPServer(cfg.Addr, newEventHandler(ctx, discovery, newReportRuntime(cacheSvc)))
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Run(ctx)
	}()

	go func() {
		<-ctx.Done()
		discovery.Close()
	}()

	go func() {
		if err := <-errCh; err != nil && ctx.Err() == nil {
			logger.Errorf(ctx, "media collection server stopped unexpectedly addr=%s err=%v", cfg.Addr, err)
		}
	}()
	return nil
}
