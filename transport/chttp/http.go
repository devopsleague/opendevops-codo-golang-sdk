package chttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	nethttp "net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/opendevops-cn/codo-golang-sdk/cerr"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type options struct {
	propagator propagation.TextMapPropagator
}

var optionsDefault = options{
	propagator: propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{}),
}

type Resp struct {
	// 业务 code
	Code cerr.ErrCode `json:"code"`
	// 用户看
	Msg string `json:"msg"`
	// 开发看
	Reason string `json:"reason"`
	// 服务器毫秒时间戳
	Timestamp string `json:"timestamp"`
	// 结构化数据
	Result json.RawMessage `json:"result"`
	// TraceID
	TraceID string `json:"trace_id"`
}

var (
	// marshalOptions is a configurable JSON format marshaller.
	marshalOptions = protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   true,
	}

	// unmarshalOptions is a configurable JSON format parser.
	unmarshalOptions = protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
)

func ResponseEncoder(writer nethttp.ResponseWriter, request *nethttp.Request, i interface{}) error {
	bs, err := marshalJSON(i)
	if err != nil {
		return err
	}

	ctx := optionsDefault.propagator.Extract(request.Context(), propagation.HeaderCarrier(request.Header))
	sp := trace.SpanContextFromContext(ctx)
	milliSecondsStr := strconv.Itoa(int(time.Now().UnixMilli()))

	// 写入
	writer.WriteHeader(nethttp.StatusOK)
	return json.NewEncoder(writer).Encode(&Resp{
		Code:      cerr.SCode,
		Msg:       "success",
		Reason:    "success",
		Timestamp: milliSecondsStr,
		Result:    bs,
		TraceID:   sp.TraceID().String(),
	})
}

func RequestBodyDecoder(r *nethttp.Request, i interface{}) error {
	const megaBytes4 = 4 << 20
	if r.ContentLength == 0 || r.ContentLength > megaBytes4 {
		return nil
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		return cerr.New(cerr.EParamUnparsedCode, err)
	}

	if len(data) == 0 {
		return nil
	}

	// reset body.
	r.Body = io.NopCloser(bytes.NewBuffer(data))

	err = unmarshalJSON(data, i)
	if err != nil {
		return cerr.New(cerr.EParamUnparsedCode, err)
	}

	return nil
}

func marshalJSON(v interface{}) ([]byte, error) {
	switch m := v.(type) {
	case json.Marshaler:
		return m.MarshalJSON()
	case proto.Message:
		return marshalOptions.Marshal(m)
	default:
		return json.Marshal(m)
	}
}

func unmarshalJSON(data []byte, v interface{}) error {
	switch m := v.(type) {
	case json.Unmarshaler:
		return m.UnmarshalJSON(data)
	case proto.Message:
		return unmarshalOptions.Unmarshal(data, m)
	default:
		rv := reflect.ValueOf(v)
		for rv := rv; rv.Kind() == reflect.Ptr; {
			if rv.IsNil() {
				rv.Set(reflect.New(rv.Type().Elem()))
			}
			rv = rv.Elem()
		}
		if m, ok := reflect.Indirect(rv).Interface().(proto.Message); ok {
			return unmarshalOptions.Unmarshal(data, m)
		}
		return json.Unmarshal(data, m)
	}
}

func ErrorEncoder(writer http.ResponseWriter, request *http.Request, err error) {
	// 错误转化
	if impl, ok := err.(interface{ ErrorName() string }); ok &&
		strings.HasSuffix(impl.ErrorName(), "ValidationError") {
		err = cerr.New(cerr.EInvalidParamCode, err)
	}

	// 错误返回
	codeError := cerr.From(err)
	errCode := codeError.Code
	statusCode := codeError.Code.AsHTTPCode()
	msg := codeError.Code.String()

	ctx := optionsDefault.propagator.Extract(request.Context(), propagation.HeaderCarrier(request.Header))
	sp := trace.SpanContextFromContext(ctx)
	milliSecondsStr := strconv.Itoa(int(time.Now().UnixMilli()))

	// 写入
	writer.WriteHeader(statusCode)
	_ = json.NewEncoder(writer).Encode(&Resp{
		Code:      errCode,
		Msg:       msg,
		Reason:    err.Error(),
		Timestamp: milliSecondsStr,
		TraceID:   sp.TraceID().String(),
	})
}

func FromError(resp *Resp) (*cerr.CodeError, bool) {
	if resp.Code == cerr.SCode {
		return nil, false
	}
	return cerr.New(resp.Code, errors.New(resp.Reason)), true
}
