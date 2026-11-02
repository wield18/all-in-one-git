package service

import (
	"all-in-one/config"
	"all-in-one/internal/base"
	"all-in-one/internal/base/constant"
	"all-in-one/internal/entity"
	httperr "all-in-one/pkg/ekit/http_err"

	consumerouter "all-in-one/internal/consume_router"
	"all-in-one/pkg/consumer"
	"all-in-one/pkg/ekit/randx"
	"all-in-one/pkg/emailqq"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/mojocn/base64Captcha"
)

const (
	HasNOEmail         = "has no email"
	EmailValidateWrong = "email validate wrong"
)

func GetLocation(c *gin.Context) {
	gao, ok := c.GetQuery("Gaojian")
	if !ok || gao != config.Conf.TheOther.SuperID {
		base.AddResultIErrAndArgs(c, ErrForbidden("nope"))
	}
	data := c.MustGet(constant.QueryData).(*entity.LocationInner)
	val := entity.ToUrlSafeBase64String(data)
	c.SetCookie(constant.ContextLocation, val, constant.ContextLocationMaxAge, "", "", false, true)
	c.Set(constant.Result, "ok")
}

func GetCaptchaEmail(qqDialer *emailqq.QQDialer, p *consumer.ConsumerPool) func(c *gin.Context) {
	return func(c *gin.Context) {
		email, ok := c.GetQuery(constant.EmailPlain)
		if !ok {
			base.AddResultIErrAndArgs(c, ErrBadRequest(HasNOEmail))
			c.Abort()
			return
		}
		// 是不是email
		if !base.IsEmail(email) {
			base.AddResultIErrAndArgs(c, ErrBadRequest(EmailValidateWrong))
			c.Abort()
			return
		}
		code, _ := randx.RandCode(constant.CaptchaLen, randx.TypeDigit)
		sendQQ := func() httperr.IError {
			err := qqDialer.SendCaptcha(email, code)
			if err != nil {
				return httperr.ErrInternalServerError(err.Error())
			}
			var zero httperr.IError
			return zero
		}
		// 加密email生成键
		aesEmail := base.CipherDB.Encrypt(email)
		sendRedis := func() httperr.IError {
			return p.BlockSubmit(consumerouter.RedisKeyValueSet, consumer.NewConsumeDataNilCallBack(entity.KeyValueRedis{
				Key:     base.Sprintf(constant.CaptchaEmailPlain, aesEmail),
				Value:   code,
				Timeout: time.Minute * 5,
			}))
		}
		errs, hasErr := AsyncPreceed(sendQQ, sendRedis)
		if hasErr {
			c.Abort()
			for _, v := range errs {
				if !v.IsZero() {
					base.AddResultIErrAndArgs(c, v)
				}
			}
			return
		}
		by, _ := json.Marshal(entity.Captcha{
			Type: constant.CaptchaEmailPlain,
			Code: aesEmail,
		})
		c.Set(constant.Result, map[string]interface{}{
			constant.Captcha: base64.RawURLEncoding.EncodeToString(by),
		})
		c.Next()
	}
}

func GetCaptchaEmailTemp(qqDialer *emailqq.QQDialer, p *consumer.ConsumerPool) func(c *gin.Context) {
	return func(c *gin.Context) {
		email, ok := c.GetQuery(constant.EmailPlain)
		if !ok {
			base.AddResultIErrAndArgs(c, ErrBadRequest(HasNOEmail))
			c.Abort()
			return
		}
		// 是不是email
		if !base.IsEmail(email) {
			base.AddResultIErrAndArgs(c, ErrBadRequest(EmailValidateWrong))
			c.Abort()
			return
		}
		code, _ := randx.RandCode(constant.CaptchaLen, randx.TypeDigit)

		// 加密email生成键
		aesEmail := base.CipherDB.Encrypt(email)
		ierr := p.BlockSubmit(consumerouter.RedisKeyValueSet, consumer.NewConsumeDataNilCallBack(entity.KeyValueRedis{
			Key:     base.Sprintf(constant.CaptchaEmailPlain, aesEmail),
			Value:   code,
			Timeout: time.Minute * 5,
		}))
		if !ierr.IsZero() {
			base.AddResultIErrAndArgs(c, ierr)
			c.Abort()
			return
		}

		by, _ := json.Marshal(entity.Captcha{
			Type: constant.CaptchaEmailPlain,
			Code: aesEmail,
		})
		ginData(c, map[string]interface{}{
			constant.Captcha:         base64.RawURLEncoding.EncodeToString(by),
			constant.CaptchaValPlain: code,
		})
		c.Next()
	}
}

// constant.Captcha:      b64s,
// constant.CaptchaToken: code,
func GetB64sCaptcha(captchaGenerator *base64Captcha.Captcha, p *consumer.ConsumerPool) func(c *gin.Context) {
	return func(c *gin.Context) {

		// 生成基本验证码图片
		_, b64s, answer, _ := captchaGenerator.Generate()
		randKey, _ := randx.RandCode(constant.RandKeyLen, randx.TypeDigit|randx.TypeLowerCase)

		// 因为这是阻塞发送
		err := p.BlockSubmit(consumerouter.RedisKeyValueSet, consumer.NewConsumeDataNilCallBack(entity.KeyValueRedis{
			Key:     base.Sprintf(constant.CaptchaB64sPlain, randKey),
			Value:   answer,
			Timeout: 5 * time.Minute,
		}))
		// 执行错误
		if !err.IsZero() {
			base.AddResultIErrAndArgs(c, err)
			c.Abort()
			return
		}
		// 这里返回的key携带类型json下携带类型
		// aMap := map[string]interface{}{
		// 	constant.CaptchaKey: randKey,
		// 	constant.CaptchaType: constant.CaptchaB64s,
		// }
		by, _ := json.Marshal(entity.Captcha{
			Type: constant.CaptchaB64sPlain,
			Code: randKey,
		})
		ginData(c, map[string]interface{}{
			constant.CaptchaImage:    b64s,
			constant.Captcha:         base64.RawURLEncoding.EncodeToString(by),
			constant.CaptchaValPlain: randKey,
		})
		c.Next()
	}
}
