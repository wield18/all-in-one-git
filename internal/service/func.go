package service

import (
	"all-in-one/internal/base"
	"all-in-one/internal/base/constant"
	httperr "all-in-one/pkg/ekit/http_err"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	noFound = "noFound"
	noFuncs = "noFuncs"
)

// 基本异步方法
func AsyncPreceed(funs ...func() httperr.IError) ([]httperr.IError, bool) {
	if len(funs) == 0 {
		panic(noFuncs)
	}
	ch := make(chan struct{}, 1)
	l := len(funs)
	ierrs := make([]httperr.IError, l)
	var hasErr bool
	for i := 0; i < l; i++ {
		go func() {
			ierr := funs[i]()
			if !ierr.IsZero() {
				hasErr = true
			}
			ierrs[i] = ierr
			ch <- struct{}{}
		}()
	}
	// 这里已经死等了
	// 但errs部分可能已经被赋值了
	for i := 0; i < l; i++ {
		<-ch
	}
	return ierrs, hasErr
}

// 超时阻塞等待一堆
func AsyncPreceedWithTimeout(timeout time.Duration, funs ...func() httperr.IError) ([]httperr.IError, bool) {
	ch := make(chan struct{}, 1)
	defer close(ch)
	if timeout <= 0 { // 如果为零或者小于零直接死等
		return AsyncPreceed(funs...)
	}
	// 有timeout的时候
	t := time.NewTimer(timeout)
	var errs []httperr.IError
	var hasErr bool
	go func() {
		errs, hasErr = AsyncPreceed(funs...)
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		return errs, hasErr
	case <-t.C:
		for i, err := range errs {
			if err.IsZero() {
				errs[i] = httperr.ErrServiceUnavailable("async timeout")
				hasErr = true
			}
		}
		return errs, hasErr
	}
}

// 写入返回
func ginData(c *gin.Context, data interface{}) {
	c.Set(constant.Result, map[string]interface{}{constant.Data: data})
}

// 内部不用主动abort,只需要返回 ierr, 只要不为nil的ierr就abort
func ginAbortIerr(c *gin.Context, ierr httperr.IError) {
	if !ierr.IsZero() {
		base.AddResultIErrAndArgs(c, ierr)
		c.Abort()
		return
	}
}
