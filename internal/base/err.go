package base

import (
	httperr "all-in-one/pkg/ekit/http_err"
	"fmt"
)

func IErrValidateFromErr(err error) httperr.IError {
	return httperr.NewIError(400, 1001, fmt.Sprintf("validate 没通过: %s", err.Error()))
}

// 一般来说 IError的 Httpcode
// 为1时使用debug
// 为100时使用info
// 200使用warn
// 其他都使用err

func IErrDebug(msg string) httperr.IError {
	return httperr.NewIError(1, 1001, msg)
}

func IErrInfo(msg string) httperr.IError {
	return httperr.NewIError(100, 1001, msg)
}

func IErrWarn(msg string) httperr.IError {
	return httperr.NewIError(200, 1001, msg)
}
