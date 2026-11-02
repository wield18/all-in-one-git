package service

import httperr "all-in-one/pkg/ekit/http_err"

func ErrBadRequest(info string) httperr.IError {
	return httperr.NewIError(400, 1001, info)
}

func ErrUnauthorized(info string) httperr.IError {
	return httperr.NewIError(401, 1001, info)
}

func ErrForbidden(info string) httperr.IError {
	return httperr.NewIError(403, 1001, info)
}

func ErrNotFound(info string) httperr.IError {
	return httperr.NewIError(404, 1001, info)
}

func ErrMethodNotAllowed(info string) httperr.IError {
	return httperr.NewIError(405, 1001, info)
}

func ErrConflict(info string) httperr.IError {
	return httperr.NewIError(409, 1001, info)
}

func ErrUnprocessableEntity(info string) httperr.IError {
	return httperr.NewIError(422, 1001, info)
}

func ErrTooManyRequests(info string) httperr.IError {
	return httperr.NewIError(429, 1001, info)
}

func ErrInternalServerError(info string) httperr.IError {
	return httperr.NewIError(500, 1001, info)
}

func ErrBadGateway(info string) httperr.IError {
	return httperr.NewIError(502, 1001, info)
}

func ErrServiceUnavailable(info string) httperr.IError {
	return httperr.NewIError(503, 1001, info)
}

func ErrGatewayTimeout(info string) httperr.IError {
	return httperr.NewIError(504, 1001, info)
}
