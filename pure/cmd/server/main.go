package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/kiselev-nikolay/ha-http-proxy/pure/pkg/proxy"
)

func main() {
	proxyCtx := &proxy.Context{
		Addr:           ":8080",
		LoggingEnabled: true,
	}
	ctxWithTimeout, cancel := context.WithCancel(context.Background())
	proxyCtx.Context = ctxWithTimeout
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err := proxy.Run(proxyCtx)
		if err != nil {
			fmt.Println(err)
		}
		wg.Done()
	}()
	<-stop
	cancel()
	wg.Wait()
	fmt.Printf("%+v", proxyCtx.Traffic)
}
