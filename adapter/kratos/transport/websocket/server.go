package websocket

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
)

type ServerOptions struct {
	handlerBuilders []IWSHandlerBuilder

	errorEncodeFunc func(w http.ResponseWriter, err error)
}

func defaultServerOptions() ServerOptions {
	return ServerOptions{
		errorEncodeFunc: func(w http.ResponseWriter, err error) {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 500,
				"msg":  err.Error(),
			})
		},
	}
}

type IServerOption interface {
	apply(*ServerOptions)
}

type ServerOptionFunc func(*ServerOptions)

func WithHandlerBuilders(handlerBuilders ...IWSHandlerBuilder) IServerOption {
	return ServerOptionFunc(func(options *ServerOptions) {
		options.handlerBuilders = append(options.handlerBuilders, handlerBuilders...)
	})
}

func WithServerOptionErrorEncodeFunc(fn func(w http.ResponseWriter, err error)) ServerOptionFunc {
	return func(options *ServerOptions) {
		options.errorEncodeFunc = fn
	}
}

func (f ServerOptionFunc) apply(options *ServerOptions) {
	f(options)
}

type Server struct {
	listener net.Listener
	mux      *http.ServeMux

	options ServerOptions
}

func NewServer(addr string, opts ...IServerOption) (*Server, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	options := defaultServerOptions()
	for _, opt := range opts {
		opt.apply(&options)
	}

	mux := http.NewServeMux()
	for _, builder := range options.handlerBuilders {
		mux.HandleFunc(builder.Path(), func(writer http.ResponseWriter, request *http.Request) {
			handler, err := builder.Build(request.Context())
			if err != nil {
				options.errorEncodeFunc(writer, err)
				return
			}
			handler.handleConnections(request.Context(), writer, request)
		})
	}
	return &Server{
		listener: listener,
		mux:      mux,
		options:  options,
	}, nil
}

func (x *Server) Start(ctx context.Context) error {
	return http.Serve(x.listener, x.mux)
}

func (x *Server) Stop(ctx context.Context) error {
	return x.listener.Close()
}
