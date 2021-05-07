package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/kiselev-nikolay/ha-http-proxy/pure/pkg/proxy"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	proxyCtx := &proxy.Context{
		Addr:    ":8080",
		Logger:  log.Default(),
		Context: ctx,
	}

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		err := proxy.Run(proxyCtx)
		if err != nil {
			fmt.Println(err)
		}
		wg.Done()
	}()

	<-proxyCtx.Done()
	wg.Wait()

	fmt.Printf("%+v", proxyCtx.Traffic)
}
