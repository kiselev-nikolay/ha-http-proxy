package main

import (
	"log"

	"github.com/kiselev-nikolay/ha-http-proxy/light/pkg/proxy"
)

func main() {
	log.Fatal(proxy.Run(":8080", true))
}
