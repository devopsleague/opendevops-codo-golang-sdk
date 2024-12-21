package chttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	nethttp "net/http"
	"reflect"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/opendevops-cn/codo-golang-sdk/cerr"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type Resp struct {
	// 业务 code
	Code cerr.ErrCode `json:"code"`
	// 开发看
	Msg string `json:"msg"`
	// 用户看
	Reason string `json:"reason"`
	// 服务器时间戳
	Timestamp uint32 `json:"timestamp"`
	// 结构化数据
	Result json.RawMessage `json:"result"`
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
	return json.NewEncoder(writer).Encode(&Resp{
		Code:      cerr.SCode,
		Msg:       "success",
		Timestamp: uint32(time.Now().Unix()),
		Result:    bs,
	})
}

func RequestBodyDecoder(r *nethttp.Request, i interface{}) error {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return cerr.New(cerr.EParamUnparsedCode, err)
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

	// 写入
	writer.WriteHeader(statusCode)
	_ = json.NewEncoder(writer).Encode(&Resp{
		Code:      errCode,
		Msg:       msg,
		Reason:    err.Error(),
		Timestamp: uint32(time.Now().Unix()),
	})
}

func FromError(resp *Resp) (*cerr.CodeError, bool) {
	if resp.Code == cerr.SCode {
		return nil, false
	}
	return cerr.New(resp.Code, errors.New(resp.Reason)), true
}
