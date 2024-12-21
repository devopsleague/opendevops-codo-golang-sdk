package cerr

import (
	"errors"
	"fmt"
	nethttp "net/http"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrCode int32

// 定义返回信息
const (
	// SCode success
	SCode = ErrCode(0)
	// EUnknownCode 未知错误!
	EUnknownCode = ErrCode(1)
	// EParamUnparsedCode 参数解析错误!
	EParamUnparsedCode = ErrCode(101)
	// EDBErrorCode 数据库查询或操作错误!
	EDBErrorCode = ErrCode(102)
	// ENoPermCode 没有任何查看权限!
	ENoPermCode = ErrCode(103)
	// EReqExpiredCode 时间范围不能超过24小时!
	EReqExpiredCode = ErrCode(104)
	// ERequiredFieldsCode 必填参数不能为空!
	ERequiredFieldsCode = ErrCode(105)
	// ECallApiCode 调用api接口失败或没有返回数据!
	ECallApiCode = ErrCode(106)
	// EK8sCheckCode k8s检查异常!
	EK8sCheckCode = ErrCode(107)
	// EHealthCheckCode API健康检查异常!
	EHealthCheckCode = ErrCode(108)
	// EInvalidConfigCode 配置定义错误!
	EInvalidConfigCode = ErrCode(109)
	// EOpTimeExceedCode 已经过了操作时效!
	EOpTimeExceedCode = ErrCode(110)
	// EInvalidParamCode 参数不符合规范!
	EInvalidParamCode = ErrCode(111)
	// EOpK8sCode k8s查询或操作异常!
	EOpK8sCode = ErrCode(112)
	// ECronCode cron操作异常!
	ECronCode = ErrCode(113)
	// EDataExistsCode 数据已存在!
	EDataExistsCode = ErrCode(114)
	// EDataNotFoundCode 数据不存在!
	EDataNotFoundCode = ErrCode(115)
	// EReadFileCode 文件读取错误!
	EReadFileCode = ErrCode(116)
	// EAddrNotMatchedCode 地址类型和地址格式不匹配!
	EAddrNotMatchedCode = ErrCode(117)
	// EInspectionCode 巡检异常!
	EInspectionCode = ErrCode(118)
	// EInvalidTokenCode token验证失败!
	EInvalidTokenCode = ErrCode(119)
	// ENoEnoughPermissionCode 权限不足!
	ENoEnoughPermissionCode = ErrCode(120)
	// EDataFormatCode 数据格式不正确!
	EDataFormatCode = ErrCode(121)
	// EUnAuthCode 未登录
	EUnAuthCode = ErrCode(201)
)

var transMap = map[ErrCode]string{
	SCode:                   "success",
	EUnknownCode:            "未知错误!",
	EParamUnparsedCode:      "参数解析错误!",
	EDBErrorCode:            "数据库查询或操作错误!",
	ENoPermCode:             "没有任何查看权限!",
	EReqExpiredCode:         "时间范围不能超过24小时!",
	ERequiredFieldsCode:     "必填参数不能为空!",
	ECallApiCode:            "调用api接口失败或没有返回数据!",
	EK8sCheckCode:           "k8s检查异常!",
	EHealthCheckCode:        "API健康检查异常!",
	EInvalidConfigCode:      "配置定义错误!",
	EOpTimeExceedCode:       "已经过了操作时效!",
	EInvalidParamCode:       "参数不符合规范!",
	EOpK8sCode:              "k8s查询或操作异常!",
	ECronCode:               "cron操作异常!",
	EDataExistsCode:         "数据已存在!",
	EDataNotFoundCode:       "数据不存在!",
	EReadFileCode:           "文件读取错误!",
	EAddrNotMatchedCode:     "地址类型和地址格式不匹配!",
	EInspectionCode:         "巡检异常!",
	EInvalidTokenCode:       "token验证失败!",
	ENoEnoughPermissionCode: "权限不足!",
	EDataFormatCode:         "数据格式不正确!",
	EUnAuthCode:             "请登陆后再操作!",
}

var httpStatusCodeMap = map[ErrCode]int{
	SCode:                   nethttp.StatusOK,
	EUnknownCode:            nethttp.StatusInternalServerError,
	EParamUnparsedCode:      nethttp.StatusBadRequest,
	EInvalidParamCode:       nethttp.StatusBadRequest,
	EDataExistsCode:         nethttp.StatusConflict,
	EDataNotFoundCode:       nethttp.StatusNotFound,
	ENoEnoughPermissionCode: nethttp.StatusForbidden,
	EUnAuthCode:             nethttp.StatusUnauthorized,
}

type registerCodeOptions struct {
	httpCode int
}

func defaultRegisterCodeOptions() registerCodeOptions {
	return registerCodeOptions{
		httpCode: nethttp.StatusInternalServerError,
	}
}

type RegisterCodeOption interface {
	apply(*registerCodeOptions)
}

type RegisterCodeFunc func(*registerCodeOptions)

func (f RegisterCodeFunc) apply(o *registerCodeOptions) {
	f(o)
}

func WithRegisterCodeOptionHTTPCode(code int) RegisterCodeOption {
	return RegisterCodeFunc(func(o *registerCodeOptions) {
		o.httpCode = code
	})
}

var registerCodeMu sync.Mutex

func RegisterCode(code ErrCode, transMsg string, opts ...RegisterCodeOption) error {
	registerCodeMu.Lock()
	defer registerCodeMu.Unlock()

	options := defaultRegisterCodeOptions()
	for _, o := range opts {
		o.apply(&options)
	}

	if _, ok := transMap[code]; ok {
		return fmt.Errorf("code %d already exists", code)
	}
	transMap[code] = transMsg
	httpStatusCodeMap[code] = options.httpCode
	return nil
}

func (code ErrCode) String() string {
	if v, ok := transMap[code]; ok {
		return v
	}
	return "未知错误"
}

func (code ErrCode) AsHTTPCode() int {
	if v, ok := httpStatusCodeMap[code]; ok {
		return v
	}
	return nethttp.StatusInternalServerError
}

type CodeError struct {
	Code   ErrCode
	Src    error
	ErrMsg string
}

func (x *CodeError) Error() string {
	// 之所以引用 ErrMsg , 是为了防止 Error 嵌套循环引用
	return fmt.Sprintf("err_code: %d, err_msg: %s", x.Code, x.ErrMsg)
}

func (x *CodeError) Unwrap() error {
	return x.Src
}

func (x *CodeError) AsGrpcError() *status.Status {
	// 之所以引用 ErrMsg , 是为了防止 Error 嵌套循环引用
	return status.New(codes.Code(x.Code), x.ErrMsg)
}

// New 创建一个新的错误, 并且将 Src 错误包装进去
// 注意: 只能 传输层调用!!!
func New(code ErrCode, src error) *CodeError {
	return &CodeError{
		Code:   code,
		Src:    src,
		ErrMsg: src.Error(),
	}
}

func From(err error) *CodeError {
	var e *CodeError
	if errors.As(err, &e) { // 自定义错误类型
		return e
	}

	return New(EUnknownCode, err)
}
