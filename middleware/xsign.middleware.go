package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/opendevops-cn/codo-golang-sdk/logger"
	xsign "github.com/opendevops-cn/codo-golang-sdk/tools/xsgin"
)

var (
	ErrXSignExpired = fmt.Errorf("接口签名已过期")
	ErrXSignInvalid = fmt.Errorf("接口签名无效")
)

const (
	MegaByte16 = 16 * 1024 * 1024
)

const (
	// 签名参数名
	xSignKey = "x-sign"
	// 时间戳参数名
	xTimestampKey = "x-ts"
)

const (
	// 签名过期容忍时间
	xSignExpiredTolerant = time.Minute
)

type XSignMiddleware struct {
	options XSignMiddlewareOptions
}

type XSignMiddlewareOptions struct {
	signKey string
	enabled bool
	logger  *logger.Helper
}

func defaultXSignMiddlewareOptions() XSignMiddlewareOptions {
	return XSignMiddlewareOptions{
		signKey: "123456",
		enabled: true,
		logger:  logger.NewHelper(logger.DefaultLogger),
	}
}

type IXSignMiddlewareOption interface {
	apply(*XSignMiddlewareOptions)
}

type XSignMiddlewareOptionFunc func(*XSignMiddlewareOptions)

func (f XSignMiddlewareOptionFunc) apply(options *XSignMiddlewareOptions) {
	f(options)
}

func WithSignKey(signKey string) XSignMiddlewareOptionFunc {
	return func(options *XSignMiddlewareOptions) {
		options.signKey = signKey
	}
}

func WithEnabled(enabled bool) XSignMiddlewareOptionFunc {
	return func(options *XSignMiddlewareOptions) {
		options.enabled = enabled
	}
}

func WithLogger(log logger.Logger) XSignMiddlewareOptionFunc {
	return func(options *XSignMiddlewareOptions) {
		options.logger = logger.NewHelper(log)
	}
}

func NewXSignMiddleware(opts ...IXSignMiddlewareOption) *XSignMiddleware {
	options := defaultXSignMiddlewareOptions()
	for _, opt := range opts {
		opt.apply(&options)
	}
	return &XSignMiddleware{options: options}
}

func (x *XSignMiddleware) ServerHTTP(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		// 如果未启用，则直接跳过
		if !x.options.enabled {
			return
		}

		r := request
		ctx := r.Context()
		if r.Method == http.MethodOptions {
			return
		}

		query := r.URL.Query()
		xSign := query.Get(xSignKey)
		i64, _ := strconv.Atoi(query.Get(xTimestampKey))
		xTs := time.Unix(int64(i64), 0)

		// 检查请求是否过期
		if time.Now().Sub(xTs) > xSignExpiredTolerant {
			x.options.logger.Errorf(ctx, "接口签名已过期: %s", r.URL.String())
			writer.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(writer).Encode(map[string]interface{}{
				"msg": ErrXSignExpired.Error(),
			})
			return
		}

		query.Del(xSignKey)
		content := query.Encode()

		if r.ContentLength < MegaByte16 {
			var bodyBytes []byte
			if r.Body != nil {
				bodyBytes, _ = io.ReadAll(r.Body)
				// 重置 body
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
			content += string(bodyBytes)
		}

		signer := xsign.NewSignV3(x.options.signKey)
		_, err := signer.Write([]byte(content))
		if err != nil {
			x.options.logger.Errorf(ctx, "签名失败: %s, url=%s", err.Error(), r.URL.String())
			writer.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(writer).Encode(map[string]interface{}{
				"msg": ErrXSignInvalid.Error(),
			})
			return
		}
		if xSign != signer.CheckSum() {
			x.options.logger.Errorf(ctx, "接口签名无效: %s", r.URL.String())
			writer.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(writer).Encode(map[string]interface{}{
				"msg": ErrXSignInvalid.Error(),
			})
			return
		}

		// 验签通过
		next(writer, request)
		return
	}
}

func (x *XSignMiddleware) ClientHTTP(ctx context.Context, r *http.Request) error {
	query := r.URL.Query()
	query.Del(xSignKey)
	query.Set(xTimestampKey, strconv.Itoa(int(time.Now().Unix())))

	content := query.Encode()

	if r.ContentLength < MegaByte16 {
		var bodyBytes []byte
		if r.Body != nil {
			bodyBytes, _ = io.ReadAll(r.Body)
			// 重置 body
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
		content += string(bodyBytes)
	}

	signer := xsign.NewSignV3(x.options.signKey)
	_, err := signer.Write([]byte(content))
	if err != nil {
		x.options.logger.Errorf(ctx, "签名失败: %s, url=%s", err.Error(), r.URL.String())
		return err
	}
	query.Set(xSignKey, signer.CheckSum())
	r.URL.RawQuery = query.Encode()
	return nil
}
