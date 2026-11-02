package tokenstorage

import (
	httperr "all-in-one/pkg/ekit/http_err"
)

var (
	errTokenEmpty   = httperr.NewIError(403, 1001, "没有token")
	errTokenUnknow  = httperr.NewIError(403, 1002, "未知token")
	errTokenExpired = httperr.NewIError(403, 1002, "token过期")
)

func errRedis(err error) httperr.IError {
	return httperr.NewIError(500, 1003, err.Error())
}
