package middleware

import (
	"all-in-one/internal/base"
	"all-in-one/internal/base/constant"
	"all-in-one/internal/entity"
	"all-in-one/internal/local"
	httperr "all-in-one/pkg/ekit/http_err"
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"

	consumerouter "all-in-one/internal/consume_router"
	"all-in-one/pkg/consumer"
)

const (
	HasNOCaptchaToken       = "HasNOCaptchaToken"
	CaptchaTypeIsNotAllowed = "CaptchaTypeIsNotAllowed"
	CaptchaNotFound         = "CaptchaNotFound"

	hasNoUserID                = "hasNoUserID"
	attemptToValidateContextID = "attemptToSuperUpdateUser"
)

// 跨域问题
func Cors(frontEndAddress string) gin.HandlerFunc {
	return func(c *gin.Context) {

		if strings.HasPrefix(c.Request.URL.Path, "/test/") {
			// 允许所有来源
			c.Header("Access-Control-Allow-Origin", "*")
		} else {
			// 其他路径限制为指定域名
			c.Header("Access-Control-Allow-Origin", frontEndAddress)
			// 注意：此时不能同时使用 "*" 和具体域名，浏览器会报错
		}

		// 允许携带凭证
		c.Header("Access-Control-Allow-Credentials", "true")
		// 允许的请求方法
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		// 允许的请求头
		c.Header("Access-Control-Allow-Headers", "Content-Type, pre-login-token, user-token, id")
		// 暴露的响应头
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
		// 预检请求缓存时间
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
	}
}

// 上下文id跟上下文token
// id := c.GetHeader(constant.ContextID) 不用管检查,网关发网关检查
func SetContextIDAndToken(c *gin.Context) {
	id := c.GetHeader(constant.ContextID)
	rawToken := c.GetHeader(constant.ContextToken)
	contextToken, err := entity.ToContextToken(id, rawToken)
	c.Set(constant.ContextID, id)
	if err != nil {
		base.AddResultIErrAndArgsPure(c, base.NewSlogIErrAndArgs(httperr.ErrForbidden(err.Error()), base.NewSlogArgs(2).SlogPath(c)))
		c.Abort()
		return
	}
	// 得看看是否有userID跟Code
	if contextToken.UserID == 0 {
		c.Set(constant.ContextToken, entity.ContextToken{ID: id})
		return
	}
	ierr := local.TokenStorage.CheckAndRestore(contextToken.UserID, contextToken.UserCode)
	if !ierr.IsZero() {
		base.AddResultIErrAndArgs(c, ierr)
		c.Abort()
		return
	}
	// 没过
	c.Set(constant.ContextToken, contextToken)
}

// 检查是否是注册user
func CheckContextUser(c *gin.Context) {
	contextToken := c.MustGet(constant.ContextToken).(entity.ContextToken)
	if contextToken.UserID == 0 {
		base.AddResultIErrAndArgs(c, httperr.ErrForbidden(hasNoUserID))
		c.Abort()
		return
	}
}

func CheckSuperContextID(c *gin.Context) {
	contextToken := c.MustGet(constant.ContextToken).(entity.ContextToken)
	if contextToken.ID != base.SuperID {
		base.AddResultIErrAndArgs(c, httperr.ErrForbidden(attemptToValidateContextID))
		c.Abort()
		return
	}
}

// location 进行记录生成结构体 entity.ContextLocation
func SetLocation(c *gin.Context) {
	rawLoc := c.GetHeader(constant.ContextLocation)
	contextLocation, err := entity.ToContextLocation(rawLoc)
	if err != nil {
		base.AddResultIErrAndArgs(c, errContextLoc(err.Error()))
		c.Abort()
		return
	}
	c.Set(constant.ContextLocation, contextLocation)
}

// 最终写入, 只支持string跟map[string]interface{}
// 之后不主动abort,根据 ierrs 的 httpcode 来直接json或者stirng
// abort必须主动调, 这里不需要进行abort检查,因为都是会到这里,而且必须主动调用abort才能防止继续运行
func Result(c *gin.Context) {
	c.Next()
	slogIErrAndArgs, okS := c.Get(constant.IErrAndArgs)
	result, okR := c.Get(constant.Result)
	if !okR { // 如果没有result
		errArgs, ok := slogIErrAndArgs.([]*base.IErrAndArgs)
		if !okS || !ok {
			slog.Error(errResultNotFound.Msg)
			c.JSON(errResultNotFound.HttpCode, errResultNotFound)
			return
		}
		// 这里得根据err的httpcode来判定最终返回code
		maxIndex := base.ForSwitchSlogIErrArgs(errArgs)
		if maxIndex == -1 { // 空
			slog.Error(errResultNotFound.Msg)
			c.JSON(errResultNotFound.HttpCode, errResultNotFound)
			return
		}
		// 这里对于没有result但是err有,而且err只是200的或者更低的直接string 200 ok
		if errArgs[maxIndex].IErr.HttpCode <= 200 {
			c.String(200, "ok")
			return
		}
		// 其他正常err
		c.JSON(errArgs[maxIndex].IErr.HttpCode, errArgs[maxIndex].IErr)
		return
	}
	// 这里说明有result
	// resul, okR := result.(map[string]interface{})
	errArgs, ok := slogIErrAndArgs.([]*base.IErrAndArgs)
	if !okS || !ok { // 没找到err,直接返回200完事
		switchResult(c, result)
		return
	}
	// 这里如果有err,检查err是否有大于200的,如果有直接否定正常的返回
	maxIndex := base.ForSwitchSlogIErrArgs(errArgs)
	if maxIndex == -1 { // err为空
		switchResult(c, result)
		return
	}
	if errArgs[maxIndex].IErr.HttpCode > 200 {
		c.JSON(errArgs[maxIndex].IErr.HttpCode, errArgs[maxIndex].IErr)
		return
	}
	switchResult(c, result) // 正常有err但 正常的err
}

func switchResult(c *gin.Context, result interface{}) {
	switch result := result.(type) {
	case string:
		c.String(200, result)
	case map[string]interface{}:
		c.JSON(200, result)
	}
}

// constant.Captcha, constant.CaptchaValPlain 必要
func CheckCaptcha(p *consumer.ConsumerPool, mustRequired bool, allowCaptchas ...string) func(c *gin.Context) {
	return func(c *gin.Context) {
		rawCode, okCode := c.GetQuery(constant.Captcha)
		captchaVal, ok := c.GetQuery(constant.CaptchaValPlain)
		if !okCode || !ok {
			if !mustRequired {
				c.Set(constant.IsCaptchaChecked, false)
				c.Set(constant.Captcha, &entity.Captcha{})
				c.Next()
				return
			}
			// 没有code但必要直接err
			base.AddResultIErrAndArgs(c, httperr.ErrBadRequest(CaptchaNotFound))
			c.Abort()
			return
		}
		// 先转型
		captcha, err := entity.ToCaptchaPtr(rawCode, captchaVal)
		if err != nil {
			base.AddResultIErrAndArgs(c, httperr.ErrForbidden(err.Error()))
			c.Abort()
			return
		}
		// 这里进行type检查
		if !base.Contain(allowCaptchas, captcha.Type) {
			base.AddResultIErrAndArgs(c, httperr.ErrBadRequest(base.Sprintf(CaptchaTypeIsNotAllowed, captcha.Type)))
			c.Abort()
			return
		}
		// redisStore CaptchaB64s:randKey -> answer
		// return CaptchaX:urlBase64(json(entity.Captcha{Code:randKey}))
		// or
		// redisStore CaptchaEmail:CipherDB(email) -> answer
		// return CaptchaX:urlBase64(json(entity.Captcha{Code:CipherDB(email)}))

		// 可以搁这里直接check
		er := p.BlockSubmit(consumerouter.RedisStringValueCheck, consumer.NewConsumeDataNilCallBack(entity.KeyValue{
			Key:   base.Sprintf(captcha.Type, captcha.Code),
			Value: captcha.Val,
		}))
		if !er.IsZero() {
			base.AddResultIErrAndArgs(c, httperr.ErrForbidden(er.Error()))
			c.Abort()
			return
		}
		captcha.IsChecked = true
		c.Set(constant.IsCaptchaChecked, true)
		c.Set(constant.Captcha, captcha)
	}
}
