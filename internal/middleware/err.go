package middleware

import (
	httperr "all-in-one/pkg/ekit/http_err"
)

var (
	errResultNotFound = httperr.NewIError(500, 1001, "result not set")
	errContextToken   = httperr.NewIError(403, 1003, "非法请求: token没过检验")

	errContextLocationNOFound = httperr.NewIError(403, 1003, "非法请求: location 没找到")
)

func errContextLoc(info string) httperr.IError {
	return httperr.NewIError(403, 1003, "非法请求: location 没过检验: "+info)
}
