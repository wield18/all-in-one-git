package middleware

import (
	"all-in-one/internal/base"
	"all-in-one/internal/base/constant"
	"errors"

	"github.com/gin-gonic/gin"
)

// 可以进行decode或validate
type FuncStruct interface {
	Fun() error
}

var (
	errBindIsNotStrucPtr    = errors.New("不是指向结构体的指针")
	errBindIsNotStrucAndPtr = errors.New("不是指向结构体跟指针")
)

func JsonBindAndValidateWithFunc(c *gin.Context, data FuncStruct) {
	if validateJsonWithCustomValidate(c, data, data.Fun) {
		c.Set(constant.JsonData, data)
	}
}

func QueryBindAndValidateWithFunc(c *gin.Context, data FuncStruct) {
	if validateQueryWithCustomValidate(c, data, data.Fun) {
		c.Set(constant.QueryData, data)
	}
}

func FormBindAndValidateWithFunc(c *gin.Context, data FuncStruct) {
	if validateFormWithCustomValidate(c, data, data.Fun) {
		c.Set(constant.FormData, data)
	}
}

// 存储指针类型
func JsonBindAndValidate[v any](c *gin.Context, dataPtr *v) {
	if validateJson(c, dataPtr) {
		c.Set(constant.JsonData, dataPtr)
	}
}

// 存储指针类型
func QueryBindAndValidate[v any](c *gin.Context, dataPtr *v) {
	if validateQuery(c, dataPtr) {
		c.Set(constant.QueryData, dataPtr)
	}
}

// 存储指针类型
func FormBindAndValidate[v any](c *gin.Context, dataPtr *v) {
	if validateForm(c, dataPtr) {
		c.Set(constant.FormData, dataPtr)
	}
}

func validateFormWithCustomValidate[v any](c *gin.Context, dataPtr v, fun func() error) bool {
	var err error
	err = c.ShouldBind(dataPtr)
	if err == nil {
		err = fun()
	}
	if err == nil {
		err = base.V.Struct(dataPtr)
	}
	if err != nil {
		base.AddResultIErrAndArgs(c, base.IErrValidateFromErr(err))
		c.Abort()
		return false
	}
	return true
}

func validateJsonWithCustomValidate[v any](c *gin.Context, dataPtr v, fun func() error) bool {
	var err error
	err = c.ShouldBindJSON(dataPtr)
	if err == nil {
		err = fun()
	}
	if err == nil {
		err = base.V.Struct(dataPtr)
	}
	if err != nil {
		base.AddResultIErrAndArgs(c, base.IErrValidateFromErr(err))
		c.Abort()
		return false
	}
	return true
}

func validateQueryWithCustomValidate[v any](c *gin.Context, dataPtr v, fun func() error) bool {
	var err error
	err = c.ShouldBindQuery(dataPtr)
	if err == nil {
		err = fun()
	}
	if err == nil {
		err = base.V.Struct(dataPtr)
	}
	if err != nil {
		base.AddResultIErrAndArgs(c, base.IErrValidateFromErr(err))
		c.Abort()
		return false
	}
	return true
}

func validateForm[v any](c *gin.Context, dataPtr v) bool {
	var err error
	err = c.ShouldBind(dataPtr)
	if err == nil {
		err = base.V.Struct(dataPtr)
	}
	if err != nil {
		base.AddResultIErrAndArgs(c, base.IErrValidateFromErr(err))
		c.Abort()
		return false
	}
	return true
}

func validateJson[v any](c *gin.Context, dataPtr v) bool {
	var err error
	err = c.ShouldBindJSON(dataPtr)
	if err == nil {
		err = base.V.Struct(dataPtr)
	}
	if err != nil {
		base.AddResultIErrAndArgs(c, base.IErrValidateFromErr(err))
		c.Abort()
		return false
	}
	return true
}

func validateQuery[v any](c *gin.Context, dataPtr v) bool {
	var err error
	err = c.ShouldBindQuery(dataPtr)
	if err == nil {
		err = base.V.Struct(dataPtr)
	}
	if err != nil {
		base.AddResultIErrAndArgs(c, base.IErrValidateFromErr(err))
		c.Abort()
		return false
	}
	return true
}
