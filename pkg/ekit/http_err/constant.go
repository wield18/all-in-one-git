package httperr

var CodeErrMap = map[int]string{
	400: "Bad Request | 请求参数错误",
	401: "Unauthorized | 未认证",
	403: "Forbidden | 无权限",
	404: "Not Found | 资源不存在",
	405: "Method Not Allowed | 方法不允许",
	409: "Conflict | 资源冲突",
	422: "Unprocessable Entity | 语义错误",
	429: "Too Many Requests | 请求过频",
	500: "Internal Server Error | 服务器内部错误",
	502: "Bad Gateway | 网关错误",
	503: "Service Unavailable | 服务不可用",
	504: "Gateway Timeout | 网关超时",
}

func ErrBadRequest(info string) IError {
	return NewIError(400, 1001, info)
}

func ErrUnauthorized(info string) IError {
	return NewIError(401, 1001, info)
}

func ErrForbidden(info string) IError {
	return NewIError(403, 1001, info)
}

func ErrNotFound(info string) IError {
	return NewIError(404, 1001, info)
}

func ErrMethodNotAllowed(info string) IError {
	return NewIError(405, 1001, info)
}

func ErrConflict(info string) IError {
	return NewIError(409, 1001, info)
}

func ErrUnprocessableEntity(info string) IError {
	return NewIError(422, 1001, info)
}

func ErrTooManyRequests(info string) IError {
	return NewIError(429, 1001, info)
}

func ErrInternalServerError(info string) IError {
	return NewIError(500, 1001, info)
}

func ErrBadGateway(info string) IError {
	return NewIError(502, 1001, info)
}

func ErrServiceUnavailable(info string) IError {
	return NewIError(503, 1001, info)
}

func ErrGatewayTimeout(info string) IError {
	return NewIError(504, 1001, info)
}
