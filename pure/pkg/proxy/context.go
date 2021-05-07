package proxy

import "context"

type Logger interface {
	Printf(format string, v ...interface{})
}

type Context struct {
	Addr    string
	Logger  Logger
	Traffic *Traffic
	context.Context
}

type Traffic struct {
	Records map[string]TrafficRecord
}

type TrafficRecord struct {
	Req Request
	Res Response
}

func NewContext(ctx context.Context, addr string, logger Logger) *Context {
	return &Context{
		Addr:    addr,
		Logger:  logger,
		Traffic: &Traffic{Records: make(map[string]TrafficRecord)},
		Context: ctx,
	}
}
