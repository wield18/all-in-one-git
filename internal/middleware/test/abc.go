package test

import (
	"all-in-one/internal/base"
	"all-in-one/internal/base/constant"
	httperr "all-in-one/pkg/ekit/http_err"

	"github.com/gin-gonic/gin"
)

// 俩都没有
func TestResult1(c *gin.Context) {}

// ierrs是空
func TestResult2(c *gin.Context) {
	// base.AddResultIErrAndArgsPure(c, nil)
	c.Set(constant.IErrAndArgs, []*base.IErrAndArgs{})
}

// ierrs有多个但都小于等于200
func TestResult3(c *gin.Context) {
	base.AddResultIErrAndArgs(c, base.IErrDebug("IErrDebug:dwdw"))
	base.AddResultIErrAndArgs(c, base.IErrInfo("IErrInfo:dwdw"))
	base.AddResultIErrAndArgs(c, base.IErrInfo("IErrInfo:dwdw123"))
	base.AddResultIErrAndArgs(c, base.IErrInfo("IErrInfo:dwdw456"))
}

// ierrs一个
func TestResult4(c *gin.Context) {
	base.AddResultIErrAndArgs(c, httperr.ErrInternalServerError("ErrInternalServerError:dwdw"))
}

// 有result但没有ierr
func TestResult5(c *gin.Context) {
	c.Set(constant.Result, map[string]interface{}{
		constant.Data: "gaojian",
	})
}

// 有result也有ierr但为空
func TestResult6(c *gin.Context) {
	c.Set(constant.Result, map[string]interface{}{
		constant.Data: "gaojian",
	})
	c.Set(constant.IErrAndArgs, []*base.IErrAndArgs{})
}

// 有result也有ierr多个200
func TestResult7(c *gin.Context) {
	base.AddResultIErrAndArgs(c, base.IErrInfo("IErrInfo:dwdw"))
	base.AddResultIErrAndArgs(c, base.IErrInfo("IErrInfo:dwdw123"))
	base.AddResultIErrAndArgs(c, base.IErrInfo("IErrInfo:dwdw456"))
	c.Set(constant.Result, "ok")
}

// 有result也有ierr多个400,500
func TestResult8(c *gin.Context) {
	base.AddResultIErrAndArgs(c, httperr.ErrInternalServerError("ErrInternalServerError:dwdw"))
	base.AddResultIErrAndArgs(c, httperr.ErrForbidden("ErrForbidden:dwdwwdw"))
	c.Set(constant.Result, map[string]interface{}{
		constant.Data: "gaojian",
	})
}

func TestContextToken(c *gin.Context) {
	c.Set(constant.Result, "ok")
}
