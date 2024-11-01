package websocket

import (
	"context"
	"encoding/json"
	http "net/http"
	"strings"

	"github.com/ccheers/xpkg/sync/errgroup"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

type WebSocket[T any, R any] struct {
	replier IReplier[T, R]
	options WSOptions
}

type Encoding struct {
	responseEncodeFunc func(w http.ResponseWriter, resp interface{})
	errorEncodeFunc    func(w http.ResponseWriter, err error)
	requestDecodeFunc  func(r *http.Request, req interface{}) error

	replyEncodeFunc      func(ws *websocket.Conn, resp interface{})
	replyErrorEncodeFunc func(ws *websocket.Conn, err error)
}

type WSPreHandleFunc func(context.Context, *http.Request) error

type WSMiddlewareFunc func(WSPreHandleFunc) WSPreHandleFunc

type WSOptions struct {
	encoding Encoding

	middlewares []WSMiddlewareFunc
}

func defaultWSOptions() WSOptions {
	return WSOptions{
		encoding: Encoding{
			responseEncodeFunc: func(w http.ResponseWriter, resp interface{}) {
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"code": 200,
					"msg":  "ok",
					"data": resp,
				})
			},
			errorEncodeFunc: func(w http.ResponseWriter, err error) {
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"code": 500,
					"msg":  err.Error(),
				})
			},
			requestDecodeFunc: func(r *http.Request, req interface{}) error {
				return json.NewDecoder(r.Body).Decode(req)
			},
			replyEncodeFunc: func(ws *websocket.Conn, resp interface{}) {
				_ = ws.WriteJSON(map[string]interface{}{
					"code": 200,
					"msg":  "ok",
					"data": resp,
				})
			},
			replyErrorEncodeFunc: func(ws *websocket.Conn, err error) {
				_ = ws.WriteJSON(map[string]interface{}{
					"code": 500,
					"msg":  err.Error(),
				})
			},
		},
	}
}

type IWSOption interface {
	apply(*WSOptions)
}

type WSOptionFunc func(*WSOptions)

func (f WSOptionFunc) apply(options *WSOptions) {
	f(options)
}

func WithResponseEncodeFunc(fn func(w http.ResponseWriter, resp interface{})) WSOptionFunc {
	return func(options *WSOptions) {
		options.encoding.responseEncodeFunc = fn
	}
}

func WithErrorEncodeFunc(fn func(w http.ResponseWriter, err error)) WSOptionFunc {
	return func(options *WSOptions) {
		options.encoding.errorEncodeFunc = fn
	}
}

func WithRequestDecodeFunc(fn func(r *http.Request, req interface{}) error) WSOptionFunc {
	return func(options *WSOptions) {
		options.encoding.requestDecodeFunc = fn
	}
}

func WithReplyEncodeFunc(fn func(ws *websocket.Conn, resp interface{})) WSOptionFunc {
	return func(options *WSOptions) {
		options.encoding.replyEncodeFunc = fn
	}
}

func WithReplyErrorEncodeFunc(fn func(ws *websocket.Conn, err error)) WSOptionFunc {
	return func(options *WSOptions) {
		options.encoding.replyErrorEncodeFunc = fn
	}
}
func WithWSMiddlewareFunc(fns ...WSMiddlewareFunc) WSOptionFunc {
	return func(options *WSOptions) {
		options.middlewares = fns
	}
}

func NewWebSocket[T any, R any](replier IReplier[T, R], opts ...IWSOption) *WebSocket[T, R] {
	options := defaultWSOptions()
	for _, opt := range opts {
		opt.apply(&options)
	}
	return &WebSocket[T, R]{
		replier: replier,
		options: options,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (x *WebSocket[T, R]) handleConnections(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	handleFn := WSPreHandleFunc(func(ctx context.Context, r *http.Request) error {
		// Upgrade initial GET request to a websocket
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return err
		}
		// Close the connection when the function returns
		defer ws.Close()
		defer x.replier.Close(ctx)

		ws.SetCloseHandler(func(code int, text string) error {
			// Close the connection when the function returns
			defer ws.Close()
			defer x.replier.Close(ctx)
			cancel()
			return nil
		})
		ws.SetPingHandler(func(appData string) error {
			return nil
		})

		eg := errgroup.WithCancel(ctx)
		// write loop
		eg.Go(func(ctx context.Context) error {
			x.writeLoop(ctx, ws)
			return nil
		})
		// read loop
		eg.Go(func(ctx context.Context) error {
			x.readLoop(ctx, ws)
			return nil
		})
		return eg.Wait()
	})

	for i := len(x.options.middlewares) - 1; i >= 0; i-- {
		handleFn = x.options.middlewares[i](handleFn)
	}

	err := handleFn(ctx, r)
	if err != nil {
		x.options.encoding.errorEncodeFunc(w, err)
	}
	return
}

func (x *WebSocket[T, R]) readLoop(ctx context.Context, ws *websocket.Conn) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		var dst R
		err := ws.ReadJSON(&dst)
		if err != nil {
			if strings.Contains(err.Error(), "connection reset by peer") {
				return
			}
			if _, isCloseErr := err.(*websocket.CloseError); isCloseErr {
				return
			}
			continue
		}

		if validate, ok := (interface{})(&dst).(interface{ Validate() error }); ok {
			err := validate.Validate()
			if err != nil {
				x.options.encoding.replyErrorEncodeFunc(ws, err)
				return
			}
		}

		err = x.replier.Apply(ctx, dst)
		if errors.Is(err, context.Canceled) {
			continue
		}
		if err != nil {
			x.options.encoding.replyErrorEncodeFunc(ws, err)
			continue
		}
	}
}

func (x *WebSocket[T, R]) writeLoop(ctx context.Context, ws *websocket.Conn) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		resp, err := x.replier.Reply(ctx)
		if errors.Is(err, context.Canceled) {
			continue
		}
		x.options.encoding.replyEncodeFunc(ws, resp)
	}
}
