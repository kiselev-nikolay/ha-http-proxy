package proxy

import (
	"net/http"
	"net/url"
	"time"
)

type Request struct {
	Method  string            `json:"method"`
	RawURL  string            `json:"url"`
	URL     *url.URL          `json:"-"`
	Headers map[string]string `json:"headers"`
}

type Response struct {
	ID      string            `json:"id"`
	Status  int               `json:"status"`
	Length  int               `json:"length"`
	Headers map[string]string `json:"headers"`
}

type ErrResponse struct {
	Errors []string `json:"errors"`
}

type Server struct {
	server *http.Server
	client *http.Client
}

func NewServer(ctx *Context) *Server {
	s := &Server{}
	s.client = &http.Client{
		Timeout: 10 * time.Second,
	}
	handler := NewHandler(ctx, s.client)
	s.server = &http.Server{
		Addr:           ctx.Addr,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	return s
}

func (s *Server) Run(ctx *Context) error {
	go s.server.ListenAndServe()
	<-ctx.Done()
	err := s.server.Shutdown(ctx)
	return err
}
