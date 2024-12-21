package cgrpc

import (
	"github.com/opendevops-cn/codo-golang-sdk/cerr"
	"google.golang.org/grpc/status"
)

func FromError(status *status.Status) (*cerr.CodeError, bool) {
	if cerr.ErrCode(status.Code()) == cerr.SCode {
		return nil, false
	}
	return cerr.New(cerr.ErrCode(status.Code()), status.Err()), true
}
