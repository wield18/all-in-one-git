package base

import (
	"all-in-one/internal/base/constant"
	httperr "all-in-one/pkg/ekit/http_err"
	"log/slog"
	"runtime"

	"github.com/gin-gonic/gin"
)

// 这是用来构建一个err的slog使用
// 一般来说 IError的 Httpcode
// 为1时使用debug
// 为100时使用info
// 200使用warn
// 其他都使用err
type IErrAndArgs struct {
	IErr httperr.IError
	Args SlogArgs
}

// 最终返回最大的httpcode的errIndex
// 内部进行长度检查
// 0 -1
// 1 0
// 其他正常
func ForSwitchSlogIErrArgs(errs []*IErrAndArgs) int {
	l := len(errs)
	if l == 0 { // 长度为0时-1
		return -1
	}
	if len(errs) == 1 { // 这里不考虑空slice的可能
		SwitchSlogIErrArgs(errs[0])
		return 0
	}
	code := 0
	codeIndex := -1
	for i, errArg := range errs {
		SwitchSlogIErrArgs(errArg)
		if errArg.IErr.HttpCode > code {
			code = errArg.IErr.HttpCode
			codeIndex = i
		}
	}

	if code == 0 {
		// 如果还是0,那就说明不太正常,因为只有零值时才行
		// 不应该为零值
		panic(code)
	}
	return codeIndex
}

// 这里真正进行了写入
func SwitchSlogIErrArgs(ierrArg *IErrAndArgs) {
	switch ierrArg.IErr.HttpCode {
	case -1:
		slog.Debug(ierrArg.IErr.Msg, ierrArg.Args...)
	case 100:
		slog.Info(ierrArg.IErr.Msg, ierrArg.Args...)
	case 200:
		slog.Warn(ierrArg.IErr.Msg, ierrArg.Args...)
	default:
		slog.Error(ierrArg.IErr.Msg, ierrArg.Args...)
	}
}

func NewSlogIErrAndArgs(ierr httperr.IError, args SlogArgs) *IErrAndArgs {
	return &IErrAndArgs{
		IErr: ierr,
		Args: args,
	}
}

var initCap = 10

func NewSlogArgs(cap ...int) SlogArgs {
	if len(cap) != 0 {
		return SlogArgs(make([]any, 0, cap[0]))
	}
	return SlogArgs(make([]any, 0, initCap))
}

// func OptionCap() {

// }

type SlogArgs []any

// 这里必然获得
func (s SlogArgs) SlogContextID(c *gin.Context) SlogArgs {
	s = append(s, constant.ContextID, c.MustGet(constant.ContextID))
	return s
}

func (s SlogArgs) SlogPath(c *gin.Context) SlogArgs {
	s = append(s, constant.RawPathPlain, c.Request.URL.Path)
	return s
}

func (s SlogArgs) SlogContextIDAndPath(c *gin.Context) SlogArgs {
	s = append(s, constant.ContextID, c.MustGet(constant.ContextID), constant.RawPathPlain, c.Request.URL.Path)
	return s
}

func (s SlogArgs) SlogJsonDate(data interface{}) SlogArgs {
	s = append(s, constant.JsonData, data)
	return s
}

func (s SlogArgs) SlogSource() SlogArgs {
	s = append(s, constant.SourcePlain, slogSourceData())
	return s
}

// 获得并构建slog.Source的json
func getCallerInfo(skip int) (function, file string, line int) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "", "", 0
	}
	// 从 pc 还原函数名
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "", file, line
	}
	return fn.Name(), file, line
}

// 获得并构建slog.Source的json
func slogSourceData() slog.Source {
	function, file, line := getCallerInfo(0)
	return slog.Source{
		Function: function,
		File:     file,
		Line:     line,
	}
}
