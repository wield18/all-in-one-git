package httperr

import (
	"fmt"
)

func NewAssertWrong(wanted interface{}, gain interface{}) error {
	return fmt.Errorf("need %#v, but gain %#v", wanted, gain)
}

// IError可能作为返回体使用
// json只暴露httpCode跟对应err
type IError struct {
	HttpCode  int    `json:"code"`  // http code
	HttpErr   string `json:"error"` // 导致的http错误
	InnerCode int    `json:"-"`     // 内部code
	Msg       string `json:"-"`     // 真正的原因, 得记录
}

// 返回 msg 个性的原因
func (i IError) Error() string {
	return i.Msg
}

func (i IError) IsZero() bool {
	return i.HttpCode == 0
}

func NewIError(httpCode, innerCode int, mesg string) IError {
	return IError{HttpCode: httpCode, HttpErr: CodeErrMap[httpCode], InnerCode: innerCode, Msg: mesg}
}
