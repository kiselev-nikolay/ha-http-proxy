package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"

	"github.com/gin-gonic/gin"
	"github.com/kiselev-nikolay/ha-http-proxy/light/pkg/trace"
)

type request struct {
	Method  string            `json:"method" binding:"required"`
	RawURL  string            `json:"url" binding:"required"`
	Headers map[string]string `json:"headers" binding:"required"`
}

type response struct {
	ID      string            `json:"id"`
	Status  int               `json:"status"`
	Length  int               `json:"length"`
	Headers map[string]string `json:"headers"`
}

type traffic struct {
	Req *request
	Res *response
}

func main() {
	r := gin.Default()
	trafficData := make(map[string]traffic)
	r.POST("/proxy", func(g *gin.Context) {
		req := &request{}
		err := g.BindJSON(req)
		if err != nil {
			g.JSON(http.StatusBadRequest, gin.H{"errors": []string{err.Error()}})
			return
		}
		url, err := url.Parse(req.RawURL)
		if err != nil {
			g.JSON(http.StatusBadRequest, gin.H{"errors": []string{err.Error()}})
		}
		reqHeaders := make(http.Header)
		for k, v := range req.Headers {
			reqHeaders.Add(k, v)
		}
		traceID := trace.GetID()
		reqHeaders.Add("X-Hhp-Trace-Id", traceID)
		proxyRes, err := http.DefaultClient.Do(&http.Request{
			Method: req.Method,
			URL:    url,
			Header: reqHeaders,
		})
		if err != nil {
			g.JSON(http.StatusBadGateway, gin.H{"errors": []string{err.Error()}})
			return
		}
		resHeaders := make(map[string]string)
		for k, v := range proxyRes.Header {
			resHeaders[k] = v[0]
		}
		length := 0
		responseBody, err := ioutil.ReadAll(proxyRes.Body)
		if err == nil {
			length = len(responseBody)
		}
		res := &response{
			ID:      traceID,
			Status:  proxyRes.StatusCode,
			Headers: resHeaders,
			Length:  length,
		}
		g.JSON(200, res)
		trafficData[traceID] = traffic{
			Req: req,
			Res: res,
		}
	})
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	go r.Run()
	<-stop
	fmt.Printf("%+v", trafficData)
}
