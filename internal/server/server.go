package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/henvic/httpretty"
	"github.com/mbndr/logo"
	"ktbs.dev/mubeng/common"
)

// Run proxy server with a user defined listener.
//
// An active log have 2 receivers, especially stdout and into file if opt.Output isn't empty.
// Then close the proxy server if it receives a signal that interrupts the program.
func Run(opt *common.Options) {
	cli := logo.NewReceiver(os.Stderr, "")
	cli.Color = true
	cli.Level = logo.DEBUG

	file, _ := logo.Open(opt.Output)
	out := logo.NewReceiver(file, "")
	out.Format = "%s: %s"

	dump = &httpretty.Logger{
		RequestHeader:  true,
		ResponseHeader: true,
		Colors:         true,
	}

	handler := &Proxy{}
	handler.Options = opt
	handler.HTTPProxy = goproxy.NewProxyHttpServer()
	handler.HTTPProxy.OnRequest().DoFunc(handler.onRequest)
	handler.HTTPProxy.OnRequest().HandleConnectFunc(handler.onConnect)
	handler.HTTPProxy.OnResponse().DoFunc(handler.onResponse)
	handler.HTTPProxy.NonproxyHandler = http.HandlerFunc(nonProxy)

	server = &http.Server{
		Addr:    opt.Address,
		Handler: handler.HTTPProxy,
	}

	log = logo.NewLogger(cli, out)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	go func() {
		<-stop

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		Stop(ctx)
	}()

	log.Infof("[PID: %d] Starting proxy server on %s", os.Getpid(), opt.Address)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// Stop will terminate proxy server
func Stop(ctx context.Context) {
	_ = server.Shutdown(ctx)
}
