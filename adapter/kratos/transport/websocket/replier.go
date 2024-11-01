package websocket

import (
	"context"
	"net/http"
)

type IReplier[T any, R any] interface {
	Apply(ctx context.Context, req R) error
	Reply(ctx context.Context) (T, error)
	Close(ctx context.Context) error
}

type Handler interface {
	handleConnections(ctx context.Context, w http.ResponseWriter, r *http.Request)
}

type IWSHandlerBuilder interface {
	Build(ctx context.Context) (Handler, error)
	Path() string
}
