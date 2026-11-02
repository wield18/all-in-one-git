package ramserver

import (
	"all-in-one/config"
	"all-in-one/internal/base"
	"all-in-one/internal/base/constant"
	"all-in-one/internal/entity"
	"all-in-one/internal/local"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"
)

func TestMain(t *testing.T) {
	r := NewRouter()

	req := httptest.NewRequest(get, "/GetLocation", nil)
	// req.Header.Set("Content-Type", "application/json")
	req.Header.Set("ID", "qwq")
	ServerReq(r, req)
	base.SaveSlogState()
}

func TestResult(t *testing.T) {
	r := NewRouter()
	testCases := []struct {
		name string
		code int
		req  func() *http.Request
	}{
		{name: "俩都没有", code: 500, req: func() *http.Request {
			return httptest.NewRequest(get, "/TestResult1", nil)
		}},
		{name: "ierrs是空", code: 500, req: func() *http.Request {
			return httptest.NewRequest(get, "/TestResult2", nil)
		}},
		{name: "ierrs有多个但都小于等于200", code: 200, req: func() *http.Request {
			return httptest.NewRequest(get, "/TestResult3", nil)
		}},
		{name: "ierrs一个", code: 500, req: func() *http.Request {
			return httptest.NewRequest(get, "/TestResult4", nil)
		}},
		{name: "有result但没有ierr", code: 200, req: func() *http.Request {
			return httptest.NewRequest(get, "/TestResult5", nil)
		}},
		{name: "有result也有ierr但为空", code: 200, req: func() *http.Request {
			return httptest.NewRequest(get, "/TestResult6", nil)
		}},
		{name: "有result也有ierr多个200", code: 200, req: func() *http.Request {
			return httptest.NewRequest(get, "/TestResult7", nil)
		}},
		{name: "有result也有ierr多个400,500", code: 500, req: func() *http.Request {
			return httptest.NewRequest(get, "/TestResult8", nil)
		}},
	}
	// req.Header.Set("Content-Type", "application/json")
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.req()
			req.Header.Add("ContextID", "qwq")
			code, _ := ServerReq(r, req)
			if code != tc.code {
				fmt.Println("nooooooooooooooooooooo")
				fmt.Println(tc.name)
				fmt.Println("nooooooooooooooooooooo")
			}
		})
	}

	base.SaveSlogState()
}

func TestContextToken(t *testing.T) {
	r := NewRouter()
	testCases := []struct {
		name string
		code int
		fun  func() int
		req  func() *http.Request
	}{
		// {name: "// id没有", code: 403, fun: func() int {
		// 	code, _ := ServerReq(r, httptest.NewRequest(get, "/TestContextToken", nil))
		// 	return code
		// }},
		// {name: "// id有但没有token", code: 200, fun: func() int {
		// 	req := httptest.NewRequest(get, "/TestContextToken", nil)
		// 	req.Header.Add("ContextID", "qwq")
		// 	code, _ := ServerReq(r, req)
		// 	return code
		// }},
		// {name: "// id有 本地查不到假的token", code: 403, fun: func() int {
		// 	req := httptest.NewRequest(get, "/TestContextToken", nil)
		// 	req.Header.Add("ContextID", "qwq")
		// 	req.Header.Add(constant.ContextToken, "dwdw")
		// 	code, _ := ServerReq(r, req)
		// 	return code
		// }},
		// {name: "// id有 真的token但本地过期了 得看看删掉没", code: 403, fun: func() int {
		// 	req := httptest.NewRequest(get, "/TestContextToken", nil)
		// 	req.Header.Add("ContextID", "qwq")
		// 	contextToken := entity.ContextToken{UserID: 123, UserCode: "dwdw"}
		// 	// 这个token只是传输是得形成一个字段
		// 	req.Header.Add(constant.ContextToken, entity.ToJsonAesString(contextToken))
		// 	local.TokenStorage.Store(contextToken.UserCode, time.Now().Add(-time.Second))
		// 	code, _ := ServerReq(r, req)
		// 	_, ok := local.TokenStorage.Get(contextToken.UserCode)
		// 	if ok {
		// 		fmt.Println("没删掉")
		// 		fmt.Println("// id有 真的token但本地过期了 得看看删掉没")
		// 		fmt.Println("没删掉")
		// 	}
		// 	return code
		// }},
		// {name: "// token本地查到没过期", code: 200, fun: func() int {
		// 	req := httptest.NewRequest(get, "/TestContextToken", nil)
		// 	req.Header.Add("ContextID", "qwq")
		// 	contextToken := entity.ContextToken{UserID: 123, UserCode: "dwdw"}
		// 	// 这个token只是传输是得形成一个字段
		// 	req.Header.Add(constant.ContextToken, entity.ToJsonAesString(contextToken))
		// 	local.TokenStorage.Store(contextToken.UserCode, time.Now().Add(time.Second))
		// 	code, _ := ServerReq(r, req)
		// 	_, ok := local.TokenStorage.Get(contextToken.UserCode)
		// 	if ok {
		// 		fmt.Println("还在")
		// 		fmt.Println("token本地查到没过期")
		// 		fmt.Println("还在")
		// 	}
		// 	return code
		// }},
		// {name: "// token本地没查到去redis也没查到", code: 403, fun: func() int {
		// 	req := httptest.NewRequest(get, "/TestContextToken", nil)
		// 	req.Header.Add("ContextID", "qwq")
		// 	// 这里得过验证
		// 	userCode, _ := base.HashCryptoCode(strconv.FormatInt(123, 10))
		// 	aToken := entity.ContextToken{
		// 		UserID:   123,
		// 		UserCode: userCode,
		// 	}
		// 	req.Header.Add(constant.ContextToken, entity.ToJsonAesString(aToken))
		// 	code, _ := ServerReq(r, req)
		// 	return code
		// }},
		{name: "// token本地没查到去redis查到, 这还得看看本地是否再次存上,还有redis是否刷新", code: 200, fun: func() int {
			req := httptest.NewRequest(get, "/TestContextToken", nil)
			req.Header.Add("ContextID", "qwq")
			// 这里得过验证
			userCode, _ := base.HashCryptoCode(strconv.FormatInt(123, 10))
			aToken := entity.ContextToken{
				UserID:   123,
				UserCode: userCode,
			}
			base.Redis.Set(context.TODO(), base.Sprintf(constant.UserToken, aToken.UserCode), nil, time.Second).Result()
			// local.TokenStorage.Store()
			req.Header.Add(constant.ContextToken, entity.ToUrlSafeBase64String(aToken))
			code, _ := ServerReq(r, req)
			_, ok := local.TokenStorage.Get(aToken.UserCode)
			if !ok {
				fmt.Println("没查到")
				fmt.Println("这还得看看本地是否再次存上")
				fmt.Println("没查到")
			}
			d, _ := base.Redis.TTL(context.TODO(), base.Sprintf(constant.UserToken, aToken.UserCode)).Result()
			fmt.Println("时长")
			fmt.Println(d)
			fmt.Println("时长")
			return code
		}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Println(tc.name)
			code := tc.fun()
			if code != tc.code {
				fmt.Println("nooooooooooooooooooooo")
				fmt.Println(tc.name)
				fmt.Println("nooooooooooooooooooooo")
			}
		})
	}
}

func TestGetLocation(t *testing.T) {
	r := NewRouter()
	// /GetLocation
	testCases := []struct {
		name string
		data entity.LocationInner
		code int
	}{
		{
			name: "no ip4", data: entity.LocationInner{
				IP:        "dwd",
				UserAgant: "dwd",
				Platform:  "dwd",
			}, code: 400,
		},
		{
			name: "no UserAgant", data: entity.LocationInner{
				IP:       "8.8.8.8",
				Platform: "dwd",
			}, code: 400,
		},
		{
			name: "pass", data: entity.LocationInner{
				IP:        "8.8.8.8",
				UserAgant: "dwd",
				Platform:  "dwd",
			}, code: 200,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
			req := httptest.NewRequest(get, fmt.Sprintf("/GetLocation?IP=%s&UserAgant=%s&Platform=%s",
				tc.data.IP, tc.data.UserAgant, tc.data.Platform,
			), nil)
			req.Header.Add("ContextID", "qwq")
			code, _ := ServerReq(r, req)
			if code != tc.code {
				fmt.Println("not Good")
				fmt.Println(tc.name)
				fmt.Println("not Good")
			}
		})
	}
}

func TestSetLocation(t *testing.T) {
	r := NewRouter()
	// /GetLocation
	testCases := []struct {
		name    string
		rawCode string
		code    int
	}{
		// {name: "false loc", rawCode: "dwdw", code: 403},
		// {name: "empty ip", rawCode: entity.ToJsonAesString(entity.LocationPlain{
		// 	UserAgant: "dwd",
		// 	Platform:  "dwd",
		// }), code: 403},
		{name: "pass", rawCode: entity.ToUrlSafeBase64String(entity.LocationInner{
			IP:        "8.8.8.8",
			UserAgant: "dwd",
			Platform:  "dwd",
		}), code: 200},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
			req := httptest.NewRequest(get, "/TestSetLocation", nil)
			req.Header.Add("ContextID", "qwq")
			req.Header.Add(constant.ContextLocation, tc.rawCode)
			code, _ := ServerReq(r, req)
			if code != tc.code {
				fmt.Println("not Good")
				fmt.Println(tc.name)
				fmt.Println("not Good")
			}
		})
	}
}

func TestCheckCaptcha(t *testing.T) {
	r := NewRouter()
	type Captcha struct {
		Type string `json:"Type" validate:"required"`
		Code string `json:"Code" validate:"required"`
		Val1 string `json:"Val1" validate:"required"`
		Val  string `json:"-" validate:"required"`
	}
	testCases := []struct {
		name         string
		encodedQuery func() string // 这是获得对应的query
		code         int
	}{

		// {name: "// 没有captcha时", encodedQuery: func() string {
		// 	formData := url.Values{}
		// 	formData.Set(CaptchaValPlain, "123456")
		// 	return formData.Encode()
		// }, code: 400},
		{name: "// 非法的captcha,这里使用正常的aes但不是对应的结构", encodedQuery: func() string {
			formData := url.Values{}
			by, _ := json.Marshal(Captcha{Type: CaptchaEmailPlain, Code: "123", Val1: "123"})
			formData.Set(constant.Captcha, string(by))
			formData.Set(CaptchaValPlain, "123456")

			// 2. 编码成字符串并转成 Reader
			return formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
		}, code: 403},
		{name: "// 使用b64s的合法captcha来试试", encodedQuery: func() string {
			formData := url.Values{}
			by, _ := json.Marshal(entity.Captcha{Type: CaptchaB64sPlain, Code: "123"})
			formData.Set(constant.Captcha, string(by))
			formData.Set(CaptchaValPlain, "123456")

			// 2. 编码成字符串并转成 Reader
			return formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
		}, code: 400},
		{name: "// 正常的captcha但验证码不对", encodedQuery: func() string {
			formData := url.Values{}
			code := "123"
			by, _ := json.Marshal(entity.Captcha{Type: CaptchaEmailPlain, Code: code})
			base.Redis.Set(context.TODO(), base.Sprintf(CaptchaEmailPlain, code), "654321", time.Minute)
			formData.Set(constant.Captcha, string(by))
			formData.Set(CaptchaValPlain, "123456")

			// 2. 编码成字符串并转成 Reader
			return formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
		}, code: 403},
		{name: "// 所有都正常", encodedQuery: func() string {
			formData := url.Values{}
			code := "123"
			by, _ := json.Marshal(entity.Captcha{Type: CaptchaEmailPlain, Code: code})
			base.Redis.Set(context.TODO(), base.Sprintf(CaptchaEmailPlain, code), "654321", time.Minute)
			formData.Set(constant.Captcha, string(by))
			formData.Set(CaptchaValPlain, "654321")

			// 2. 编码成字符串并转成 Reader
			return formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
		}, code: 200},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(get, "/TestCheckCaptcha?"+tc.encodedQuery(), nil)
			req.Header.Add("ContextID", "qwq")
			code, _ := ServerReq(r, req)
			if code != tc.code {
				fmt.Println("not Good")
				fmt.Println(tc.name)
				fmt.Println("not Good")
			}
		})
	}

}

// router := internal.InitRouter()
// 	formData := url.Values{}
// 	formData.Set("abc", "zhangsan")

// 	// 2. 编码成字符串并转成 Reader
// 	encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
// 	fmt.Println(encoded)
// 	body := strings.NewReader(encoded)

// 	path := internal.HandleFormAndQuery.Path + "?gaojian=gaojianQuery"

// 	// 发送get
// 	req := httptest.NewRequest(internal.HandleFormAndQuery.Method, path, body)
// 	req.Header.Set("gaojian1", "abc")
// 	req.Header.Set("ID", "application/json")
// 	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
// 	internal.ServerReq(router, req)

func TestRegister(t *testing.T) {
	r := NewRouter()
	// /GetLocation
	testCases := []struct {
		name  string
		query func() string

		formData func() *strings.Reader
		code     int
	}{
		// {name: "has no email and phone", code: 400, query: func() string {
		// 	formData := url.Values{}
		// 	code := "123"
		// 	by, _ := json.Marshal(entity.Captcha{Type: CaptchaEmailPlain, Code: code})
		// 	base.Redis.Set(context.TODO(), base.Sprintf(CaptchaEmailPlain, code), "654321", time.Minute)
		// 	formData.Set(constant.Captcha, string(by))
		// 	formData.Set(CaptchaValPlain, "654321")
		// 	// 2. 编码成字符串并转成 Reader
		// 	return formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
		// }, formData: func() *strings.Reader {
		// 	// entity.UserV1Register{}
		// 	formData := url.Values{}
		// 	formData.Set(NamePlain, "zhangsan")
		// 	formData.Set(PasswordPlain, ("123456")
		// 	// formData.Set(EmailPlain, ("a@.com"))
		// 	// 2. 编码成字符串并转成 Reader
		// 	encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
		// 	return strings.NewReader(encoded)
		// }},
		// {name: "有email redis也能找着", code: 400, query: func() string {
		// 	formData := url.Values{}
		// 	code := "123"
		// 	by, _ := json.Marshal(entity.Captcha{Type: CaptchaEmailPlain, Code: code})
		// 	base.Redis.Set(context.TODO(), base.Sprintf(CaptchaEmailPlain, code), "654321", time.Minute)
		// 	formData.Set(constant.Captcha, string(by))
		// 	formData.Set(CaptchaValPlain, "654321")
		// 	// 2. 编码成字符串并转成 Reader
		// 	return formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
		// }, formData: func() *strings.Reader {
		// 	// entity.UserV1Register{}
		// 	formData := url.Values{}
		// 	formData.Set(NamePlain, "zhangsan")
		// 	formData.Set(PasswordPlain, ("123456")
		// 	// 2. 编码成字符串并转成 Reader
		// 	encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
		// 	return strings.NewReader(encoded)
		// }},
		{name: "good", code: 200, query: func() string {
			formData := url.Values{}
			code := ("a123@.com")
			by, _ := json.Marshal(entity.Captcha{Type: CaptchaEmailPlain, Code: code})
			base.Redis.Set(context.TODO(), base.Sprintf(CaptchaEmailPlain, code), "654321", time.Minute)
			formData.Set(constant.Captcha, string(by))
			formData.Set(CaptchaValPlain, "654321")
			formData.Set(constant.ContextLocation, entity.ToUrlSafeBase64String(entity.LocationInner{IP: "8.8.8.8", Platform: "gaojian", UserAgant: "user1"}))
			// 2. 编码成字符串并转成 Reader
			return formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
		}, formData: func() *strings.Reader {
			_ = entity.UserV1{}

			formData := url.Values{}
			formData.Set(NamePlain, "guyuchun")
			formData.Set(PasswordPlain, "123456")
			// 2. 编码成字符串并转成 Reader
			encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
			return strings.NewReader(encoded)
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
			req := httptest.NewRequest(post, "/Register?"+tc.query(), tc.formData())
			req.Header.Set("ContextID", "qwq")
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set(constant.ContextLocation, entity.ToUrlSafeBase64String(entity.LocationInner{IP: "8.8.8.8", Platform: "gaojian", UserAgant: "user1"}))

			code, _ := ServerReq(r, req)
			if code != tc.code {
				fmt.Println("not Good")
				fmt.Println(tc.name)
				fmt.Println("not Good")
			}
		})
	}
}

func TestLogin(t *testing.T) {
	r := NewRouter()

	getQueryWithCaptcha := func() string {
		formData := url.Values{}
		code := ("a@.com")
		val := "654321"
		by, _ := json.Marshal(entity.Captcha{Type: CaptchaEmailPlain, Code: code})
		base.Redis.Set(context.TODO(), base.Sprintf(CaptchaEmailPlain, code), val, time.Minute)
		formData.Set(constant.Captcha, string(by))
		formData.Set(CaptchaValPlain, val)
		formData.Set(constant.ContextLocation, entity.ToUrlSafeBase64String(entity.LocationInner{IP: "8.8.8.8", Platform: "gaojian", UserAgant: "user1"}))
		return formData.Encode()
	}

	getReq := func(query string, body io.Reader) *http.Request {
		return httptest.NewRequest(post, "/Login?"+query, body)
	}

	// /GetLocation
	testCases := []struct {
		name  string
		datas func() (req *http.Request)
		code  int
	}{
		{name: "// 只有密码", datas: func() (req *http.Request) {
			_ = entity.UserV1{}
			formData := url.Values{}
			formData.Set(PasswordPlain, "123456")
			// 2. 编码成字符串并转成 Reader
			encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
			body := strings.NewReader(encoded)
			req = getReq("", body)
			req.Header.Set("ContextID", "qwq")
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set(constant.ContextLocation, entity.ToUrlSafeBase64String(entity.LocationInner{IP: "8.8.8.8", Platform: "gaojian", UserAgant: "user1"}))
			return
		}, code: 400},
		// "$2a$10$HJ2yTYrCutBQnnQ3/8gRHOzo45r3bAc9IkTVrLJSo37n9b/.BSuLe"
		// ""
		{name: "// 没有captcha但其他正常", datas: func() (req *http.Request) {
			formData := url.Values{}
			formData.Set(PasswordPlain, "123456")
			formData.Set(NamePlain, "zhangsan2")
			// 2. 编码成字符串并转成 Reader
			encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
			body := strings.NewReader(encoded)
			req = getReq("", body)
			req.Header.Set("ContextID", "qwq")
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set(constant.ContextLocation, entity.ToUrlSafeBase64String(entity.LocationInner{IP: "8.8.8.8", Platform: "gaojian", UserAgant: "user1"}))
			return
		}, code: 200},
		{name: "// 没有captcha但密码不对", datas: func() (req *http.Request) {
			formData := url.Values{}
			formData.Set(PasswordPlain, "654321")
			formData.Set(NamePlain, "zhangsan2")
			// 2. 编码成字符串并转成 Reader
			encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
			body := strings.NewReader(encoded)
			req = getReq("", body)
			req.Header.Set("ContextID", "qwq")
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set(constant.ContextLocation, entity.ToUrlSafeBase64String(entity.LocationInner{IP: "8.8.8.8", Platform: "gaojian", UserAgant: "user1"}))
			return
		}, code: 400},
		{name: "// 没有captcha但location不对", datas: func() (req *http.Request) {
			formData := url.Values{}
			formData.Set(PasswordPlain, "123456")
			formData.Set(NamePlain, "zhangsan2")
			// 2. 编码成字符串并转成 Reader
			encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
			body := strings.NewReader(encoded)

			req = getReq("", body)
			req.Header.Set("ContextID", "qwq")
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set(constant.ContextLocation, entity.ToUrlSafeBase64String(entity.LocationInner{IP: "16.16.16.16", Platform: "gaojian", UserAgant: "user1"}))

			return
		}, code: 200},
		{name: "// 有captcha新地点", datas: func() (req *http.Request) {
			query := getQueryWithCaptcha()
			formData := url.Values{}
			formData.Set(PasswordPlain, "123456")
			formData.Set(NamePlain, "zhangsan2")
			// 2. 编码成字符串并转成 Reader
			encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
			body := strings.NewReader(encoded)

			req = getReq(query, body)
			req.Header.Set("ContextID", "qwq")
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set(constant.ContextLocation, entity.ToUrlSafeBase64String(entity.LocationInner{IP: "16.16.16.16", Platform: "gaojian", UserAgant: "user1"}))

			return
		}, code: 200},
		{name: "// 有captcha密码不对", datas: func() (req *http.Request) {
			query := getQueryWithCaptcha()
			formData := url.Values{}
			formData.Set(PasswordPlain, "6543213")
			formData.Set(NamePlain, "zhangsan2")
			// 2. 编码成字符串并转成 Reader
			encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
			body := strings.NewReader(encoded)

			req = getReq(query, body)
			req.Header.Set("ContextID", "qwq")
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set(constant.ContextLocation, entity.ToUrlSafeBase64String(entity.LocationInner{IP: "8.8.8.8", Platform: "gaojian", UserAgant: "user1"}))

			return
		}, code: 400},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"

			req := tc.datas()

			code, _ := ServerReq(r, req)
			if code != tc.code {
				fmt.Println("not Good")
				fmt.Println(tc.name)
				fmt.Println("not Good")
			}
		})
	}
}

func TestKeys(*testing.T) {
	base.InitAll()
	str := "123456"
	// code, _ := base.HashCryptoCode(str)
	code := "$2a$10$7imfL49d79N4N97Wp01GX.TBEE2v3DWZqpykczzjs2nvThg77XQfu"
	if base.CheckCryptoCode(code, str) {
		fmt.Println("ok")
	}

}

func TestUpdateUserSuper(t *testing.T) {
	initReq := func(query string, body io.Reader, id string) *http.Request {
		req := httptest.NewRequest(post, "/UpdateUserSuper?"+query, body)
		req.Header.Set("ContextID", id)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return req
	}

	r := NewRouter()
	// /GetLocation
	testCases := []struct {
		name   string
		getReq func() (req *http.Request)
		code   int
	}{
		{name: "// 不是super时", getReq: func() (req *http.Request) {
			formData := url.Values{}
			formData.Set(IDPlain, "7")
			formData.Set(PasswordPlain, "123456")
			formData.Set(EmailPlain, ("123@.com"))

			// 2. 编码成字符串并转成 Reader
			encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
			body := strings.NewReader(encoded)
			req = initReq("", body, "dwdw")
			return
		}, code: 403},
		{name: "// 是super简单改一下", getReq: func() (req *http.Request) {
			formData := url.Values{}
			formData.Set(IDPlain, "7")
			formData.Set(PasswordPlain, "123456")
			formData.Set(EmailPlain, ("123@.com"))
			// 2. 编码成字符串并转成 Reader
			encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
			body := strings.NewReader(encoded)
			req = initReq("", body, config.Conf.TheOther.SuperID)
			return
		}, code: 200},
		{name: "// 删一下", getReq: func() (req *http.Request) {
			formData := url.Values{}
			formData.Set(DeletePlain, "true")
			formData.Set(IDPlain, "7")
			// 2. 编码成字符串并转成 Reader
			encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
			body := strings.NewReader(encoded)
			req = initReq("", body, config.Conf.TheOther.SuperID)
			return
		}, code: 200},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.getReq()
			code, _ := ServerReq(r, req)
			if code != tc.code {
				fmt.Println("not Good")
				fmt.Println(tc.name)
				fmt.Println("not Good")
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	r := NewRouter()

	// getQueryWithCaptcha := func(T string, codePlain string, val string) string {
	// 	formData := url.Values{}
	// 	code := codePlain)
	// 	by, _ := json.Marshal(entity.Captcha{Type: T, Code: code})
	// 	base.Redis.Set(context.TODO(), base.Sprintf(T, code), val, time.Minute)
	// 	formData.Set(constant.Captcha, string(by))
	// 	formData.Set(CaptchaValPlain, val)
	// 	formData.Set(constant.ContextLocation, entity.ToJsonAesString(entity.LocationInner{IP: "8.8.8.8", Platform: "gaojian", UserAgant: "user1"}))
	// 	return formData.Encode()
	// }

	getReq := func(query string, body io.Reader) *http.Request {
		return httptest.NewRequest(post, "/UpdateUser?"+query, body)
	}

	processReq := func(req *http.Request, needUserToken bool) {

		req.Header.Set("ContextID", "qwq")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if needUserToken {
			var id int64 = 9
			code, _ := base.HashCryptoCode(strconv.FormatInt(id, 10))
			local.TokenStorage.Store(code, time.Now().Add(constant.UserTokenDuration))
			req.Header.Set(constant.ContextToken, entity.ToUrlSafeBase64String(entity.ContextToken{
				UserID:   id,
				UserCode: code}))
		}
		req.Header.Set(constant.ContextLocation, entity.ToUrlSafeBase64String(entity.LocationInner{IP: "8.8.8.8", Platform: "gaojian", UserAgant: "user1"}))
	}

	// /GetLocation
	testCases := []struct {
		name  string
		datas func() (req *http.Request)
		code  int
	}{
		{name: "// bytes", datas: func() (req *http.Request) {
			_ = entity.UserV1{}
			formData := url.Values{}
			by, _ := json.Marshal(entity.Bytes{Video: entity.VideoBytes{
				IncludedTags:    []entity.VideoBytesTag{{Tag: "zxc", Ratio: 1}},
				ExcludedUserIDs: []int64{1, 2, 3, 11, 12, 13},
			}})
			fmt.Println(string(by))
			formData.Set(BytesPlain, string(by))
			// 2. 编码成字符串并转成 Reader
			encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
			body := strings.NewReader(encoded)
			req = getReq("", body)
			processReq(req, true)
			return
		}, code: 200},

		// {name: "// 不携带userToken", datas: func() (req *http.Request) {
		// 	_ = entity.UserV1{}
		// 	formData := url.Values{}
		// 	formData.Set(PasswordPlain, ("123456")

		// 	// 2. 编码成字符串并转成 Reader
		// 	encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
		// 	body := strings.NewReader(encoded)
		// 	req = getReq("", body)
		// 	processReq(req, false)
		// 	return
		// }, code: 403},
		// {name: "// 更新字段多余1", datas: func() (req *http.Request) {
		// 	_ = entity.UserV1{}
		// 	formData := url.Values{}
		// 	formData.Set(PasswordPlain, ("123456")
		// 	formData.Set(EmailPlain, ("wdw@.com"))
		// 	// 2. 编码成字符串并转成 Reader
		// 	encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
		// 	body := strings.NewReader(encoded)
		// 	req = getReq("", body)
		// 	processReq(req, true)
		// 	return
		// }, code: 400},
		// {name: "// 改名字", datas: func() (req *http.Request) {
		// 	_ = entity.UserV1{}
		// 	formData := url.Values{}
		// 	formData.Set(NamePlain, "guyuchun")
		// 	// 2. 编码成字符串并转成 Reader
		// 	encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
		// 	body := strings.NewReader(encoded)
		// 	req = getReq("", body)
		// 	processReq(req, true)
		// 	return
		// }, code: 200},
		// {name: "// 状态", datas: func() (req *http.Request) {
		// 	_ = entity.UserV1{}
		// 	formData := url.Values{}
		// 	formData.Set(StatusPlain, "sad")
		// 	// 2. 编码成字符串并转成 Reader
		// 	encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
		// 	body := strings.NewReader(encoded)
		// 	req = getReq("", body)
		// 	processReq(req, true)
		// 	return
		// }, code: 200},

		// {name: "// Password 旧密码不正确", datas: func() (req *http.Request) {
		// 	_ = entity.UserV1{}
		// 	formData := url.Values{}
		// 	formData.Set(PasswordPlain, ("654321")
		// 	formData.Set(OldPasswordPlain, ("789")
		// 	// 2. 编码成字符串并转成 Reader
		// 	encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
		// 	body := strings.NewReader(encoded)
		// 	query := getQueryWithCaptcha(CaptchaB64sPlain, "wdw", "123456")
		// 	req = getReq(query, body)
		// 	processReq(req, true)
		// 	return
		// }, code: 400},
		// {name: "// Password captcha没检查", datas: func() (req *http.Request) {
		// 	_ = entity.UserV1{}
		// 	formData := url.Values{}
		// 	formData.Set(PasswordPlain, ("654321")
		// 	formData.Set(OldPasswordPlain, ("789")
		// 	// 2. 编码成字符串并转成 Reader
		// 	encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
		// 	body := strings.NewReader(encoded)
		// 	// query := getQueryWithCaptcha(CaptchaB64sPlain, "wdw", "123456")
		// 	req = getReq("", body)
		// 	processReq(req, true)
		// 	return
		// }, code: 400},
		// {name: "// Password 正常", datas: func() (req *http.Request) {
		// 	_ = entity.UserV1{}
		// 	formData := url.Values{}
		// 	formData.Set(PasswordPlain, ("654321")
		// 	formData.Set(OldPasswordPlain, ("123456")
		// 	// 2. 编码成字符串并转成 Reader
		// 	encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
		// 	body := strings.NewReader(encoded)
		// 	query := getQueryWithCaptcha(CaptchaB64sPlain, "wdw", "123456")
		// 	req = getReq(query, body)
		// 	processReq(req, true)
		// 	return
		// }, code: 200},
		// {name: "// phone captcha没检查", datas: func() (req *http.Request) {
		// 	_ = entity.UserV1{}
		// 	formData := url.Values{}

		// 	formData.Set(PhonePlain, ("+86123456789123")
		// 	// 2. 编码成字符串并转成 Reader
		// 	encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
		// 	body := strings.NewReader(encoded)
		// 	// query := getQueryWithCaptcha(CaptchaB64sPlain, "wdw", "123456")
		// 	req = getReq("", body)
		// 	processReq(req, true)
		// 	return
		// }, code: 400},

		// {name: "// phone 正常", datas: func() (req *http.Request) {
		// 	_ = entity.UserV1{}
		// 	formData := url.Values{}

		// 	formData.Set(PhonePlain, ("+86123456789123")
		// 	// 2. 编码成字符串并转成 Reader
		// 	encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
		// 	body := strings.NewReader(encoded)
		// 	query := getQueryWithCaptcha(CaptchaPhonePlain, "wdw", "123456")
		// 	req = getReq(query, body)
		// 	processReq(req, true)
		// 	return
		// }, code: 200},
		// {name: "// email captcha没检查", datas: func() (req *http.Request) {
		// 	_ = entity.UserV1{}
		// 	formData := url.Values{}

		// 	formData.Set(EmailPlain, ("a@.com"))
		// 	// 2. 编码成字符串并转成 Reader
		// 	encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
		// 	body := strings.NewReader(encoded)
		// 	// query := getQueryWithCaptcha(CaptchaB64sPlain, "wdw", "123456")
		// 	req = getReq("", body)
		// 	processReq(req, true)
		// 	return
		// }, code: 400},
		// {name: "// email 正常", datas: func() (req *http.Request) {
		// 	_ = entity.UserV1{}
		// 	formData := url.Values{}

		// 	formData.Set(EmailPlain, ("a@.com"))
		// 	// 2. 编码成字符串并转成 Reader
		// 	encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
		// 	body := strings.NewReader(encoded)
		// 	query := getQueryWithCaptcha(CaptchaEmailPlain, "wdw", "123456")
		// 	req = getReq(query, body)
		// 	processReq(req, true)
		// 	return
		// }, code: 200},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.datas()
			code, _ := ServerReq(r, req)
			if code != tc.code {
				fmt.Println("not Good")
				fmt.Println(tc.name)
				fmt.Println("not Good")
			}
		})
	}
}

func TestGetCaptchaEmail(t *testing.T) {
	initReq := func(query string, body io.Reader, id string) *http.Request {
		req := httptest.NewRequest(get, "/GetCaptchaEmail?"+query, body)
		req.Header.Set("ContextID", id)
		req.Header.Set(constant.ContextLocation, entity.ToUrlSafeBase64String(entity.LocationInner{IP: "8.8.8.8", Platform: "gaojian", UserAgant: "user1"}))
		return req
	}

	r := NewRouter()
	// /GetLocation
	testCases := []struct {
		name   string
		getReq func() (req *http.Request)
		code   int
	}{

		{name: "正常", getReq: func() (req *http.Request) {
			req = initReq("Email="+"363099642@qq.com", nil, "dwd")
			return
		}, code: 200},
		{name: "不正常", getReq: func() (req *http.Request) {
			req = initReq("Email="+"13456", nil, "dwd")
			return
		}, code: 400},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.getReq()
			code, _ := ServerReq(r, req)
			if code != tc.code {
				fmt.Println("not Good")
				fmt.Println(tc.name)
				fmt.Println("not Good")
			}
		})
	}
}
func TestGetB64sCaptcha(t *testing.T) {
	initReq := func(query string, body io.Reader, id string) *http.Request {
		req := httptest.NewRequest(get, "/GetB64sCaptcha?"+query, body)
		req.Header.Set("ContextID", id)
		req.Header.Set(constant.ContextLocation, entity.ToUrlSafeBase64String(entity.LocationInner{IP: "8.8.8.8", Platform: "gaojian", UserAgant: "user1"}))
		return req
	}

	r := NewRouter()
	// /GetLocation
	testCases := []struct {
		name   string
		getReq func() (req *http.Request)
		code   int
	}{

		{name: "正常", getReq: func() (req *http.Request) {
			req = initReq("", nil, "dwd")
			return
		}, code: 200},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.getReq()
			code, _ := ServerReq(r, req)
			if code != tc.code {
				fmt.Println("not Good")
				fmt.Println(tc.name)
				fmt.Println("not Good")
			}
		})
	}
}

// user.GET("/SearchUserSecondTable",
// user.GET("/SearchUserMany",

func TestSearchUserOne(t *testing.T) {
	initReq := func(query string, body io.Reader, id string) *http.Request {
		local.TokenStorage.Store("asd", time.Now().Add(time.Minute))
		req := httptest.NewRequest(get, "/SearchUserOne?"+query, body)
		req.Header.Set("ContextID", id)
		req.Header.Set(constant.ContextToken, entity.ToUrlSafeBase64String(entity.ContextToken{UserID: 123, UserCode: "asd"}))
		req.Header.Set(constant.ContextLocation, entity.ToUrlSafeBase64String(entity.LocationInner{IP: "8.8.8.8", Platform: "gaojian", UserAgant: "user1"}))
		return req
	}

	r := NewRouter()
	// /GetLocation
	testCases := []struct {
		name   string
		getReq func() (req *http.Request)
		code   int
	}{

		{name: "// id", getReq: func() (req *http.Request) {
			data := url.Values{}
			data.Set(IDPlain, "1")
			encoded := data.Encode()
			// body := strings.NewReader(encoded)
			req = initReq(encoded, nil, "dwd")
			return
		}, code: 200},
		{name: "// name", getReq: func() (req *http.Request) {
			data := url.Values{}
			data.Set(NamePlain, "gaojian")
			encoded := data.Encode()
			// body := strings.NewReader(encoded)
			req = initReq(encoded, nil, "dwd")
			return
		}, code: 200},
		{name: "// email", getReq: func() (req *http.Request) {
			data := url.Values{}
			data.Set(EmailPlain, ("a@.com"))
			encoded := data.Encode()
			// body := strings.NewReader(encoded)
			req = initReq(encoded, nil, "dwd")
			return
		}, code: 200},
		{name: "// phone", getReq: func() (req *http.Request) {
			data := url.Values{}
			data.Set(PhonePlain, "+86123456789123")
			encoded := data.Encode()
			// body := strings.NewReader(encoded)
			req = initReq(encoded, nil, "dwd")
			return
		}, code: 200},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.getReq()
			code, _ := ServerReq(r, req)
			if code != tc.code {
				fmt.Println("not Good")
				fmt.Println(tc.name)
				fmt.Println("not Good")
			}
		})
	}

}

func TestSearchUserMany(t *testing.T) {

	initReq := func(query string, body io.Reader, id string) *http.Request {
		local.TokenStorage.Store("asd", time.Now().Add(time.Minute))
		req := httptest.NewRequest(get, "/SearchUserMany?"+query, body)
		req.Header.Set("ContextID", id)
		req.Header.Set(constant.ContextToken, entity.ToUrlSafeBase64String(entity.ContextToken{UserID: 123, UserCode: "asd"}))
		req.Header.Set(constant.ContextLocation, entity.ToUrlSafeBase64String(entity.LocationInner{IP: "8.8.8.8", Platform: "gaojian", UserAgant: "user1"}))
		return req
	}

	r := NewRouter()
	// /GetLocation
	testCases := []struct {
		name   string
		getReq func() (req *http.Request)
		code   int
	}{

		{name: "// id", getReq: func() (req *http.Request) {
			data := url.Values{}
			by, _ := json.Marshal([]int64{1, 2, 3, 9})
			data.Set(IDsPlain, string(by))
			encoded := data.Encode()
			req = initReq(encoded, nil, "dwd")
			return
		}, code: 200},
		{name: "// name", getReq: func() (req *http.Request) {
			data := url.Values{}
			data.Set(NamePlain, "gaojian")
			encoded := data.Encode()
			req = initReq(encoded, nil, "dwd")
			return
		}, code: 200},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.getReq()
			code, _ := ServerReq(r, req)

			if code != tc.code {
				fmt.Println("not Good")
				fmt.Println(tc.name)
				fmt.Println("not Good")
			}
		})
	}
	// ids
	// name

}

func TestUploadVideoFile(t *testing.T) {
	target := "UploadVideoFile"
	initReq := func(method string, query string, body io.Reader) *http.Request {
		local.TokenStorage.Store("asd", time.Now().Add(time.Minute))
		req := httptest.NewRequest(method, fmt.Sprintf("/%s?", target)+query, body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("ContextID", "id")
		req.Header.Set(constant.ContextToken, entity.ToUrlSafeBase64String(entity.ContextToken{UserID: 123, UserCode: "asd"}))
		req.Header.Set(constant.ContextLocation, entity.ToUrlSafeBase64String(entity.LocationInner{IP: "8.8.8.8", Platform: "gaojian", UserAgant: "user1"}))
		return req
	}

	getBufAndWriter := func(filePath string, fun func(w *multipart.Writer)) (*bytes.Buffer, *multipart.Writer) {
		// 读文件
		// 无文件扩展名
		fileName := filepath.Base(filePath)[:len(filepath.Base(filePath))-len(filepath.Ext(filePath))]
		by, err := os.ReadFile(filePath)
		if err != nil {
			panic(err)
		}
		// 2. 客户端算 SHA256
		sum := sha256.Sum256(by)
		sha256Hex := hex.EncodeToString(sum[:]) // 64 字符 hex

		// 3. 构造 multipart body
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		// 3.1 文件 part
		fw, err := w.CreateFormFile("File", fileName)
		if err != nil {
			panic(err)
		}
		if _, err := fw.Write(by); err != nil {
			panic(err)
		}
		_ = w.WriteField("SHA256", sha256Hex)
		fun(w)

		// 4. 关闭 writer，写入 boundary 结尾
		if err := w.Close(); err != nil {
			panic(err)
		}
		return &buf, w
	}

	r := NewRouter()
	// /GetLocation
	testCases := []struct {
		name   string
		getReq func() (req *http.Request)
		code   int
	}{
		// {name: "// 大于1mb跟", getReq: func() (req *http.Request) {
		// 	_ = entity.VideoFileUpload{}
		// 	bufPtr, w := getBufAndWriter(`D:\Animation\Minus8_Archive 一个性的时代的代表\Minus8 Archive\ANIMATIONS\!MP4s\minus zone.mp4`,
		// 		func(w *multipart.Writer) {
		// 			w.WriteField(DescriptionPlain, "gaojian")
		// 			by, _ := json.Marshal([]string{"asd", "zxc"})
		// 			w.WriteField(TagsPlain, string(by))
		// 		})
		// 	req = initReq(post, "", bufPtr)
		// 	req.Header.Set("Content-Type", w.FormDataContentType())
		// 	return
		// }, code: 200},
		{name: "// 小于1mb上传下", getReq: func() (req *http.Request) {
			bufPtr, w := getBufAndWriter(`D:\Animation\Night.mp4`,
				func(w *multipart.Writer) {
					w.WriteField(DescriptionPlain, "gaojian")
					by, _ := json.Marshal([]string{"asd", "zxc"})
					w.WriteField(TagsPlain, string(by))
				})
			req = initReq(post, "", bufPtr)
			req.Header.Set("Content-Type", w.FormDataContentType())

			return
		}, code: 200},
		{name: "// 小于1mb上传下", getReq: func() (req *http.Request) {
			bufPtr, w := getBufAndWriter(`D:\Animation\Minus8_Archive 一个性的时代的代表\Minus8 Archive\ANIMATIONS\!MP4s\Music addiction.mp4`,
				func(w *multipart.Writer) {
					w.WriteField(DescriptionPlain, "gaojian")
					by, _ := json.Marshal([]string{"asd", "zxc"})
					w.WriteField(TagsPlain, string(by))
				})
			req = initReq(post, "", bufPtr)
			req.Header.Set("Content-Type", w.FormDataContentType())

			return
		}, code: 200},
		{name: "// 小于1mb上传下", getReq: func() (req *http.Request) {
			bufPtr, w := getBufAndWriter(`D:\Animation\Minus8_Archive 一个性的时代的代表\Minus8 Archive\ANIMATIONS\!MP4s\YC6CvHa.mp4`,
				func(w *multipart.Writer) {
					w.WriteField(DescriptionPlain, "gaojian")
					by, _ := json.Marshal([]string{"asd", "zxc"})
					w.WriteField(TagsPlain, string(by))
				})
			req = initReq(post, "", bufPtr)
			req.Header.Set("Content-Type", w.FormDataContentType())

			return
		}, code: 200},
		// {name: "// 替换上传", getReq: func() (req *http.Request) {
		// 	bufPtr, w := getBufAndWriter(`D:\Animation\Minus8_Archive 一个性的时代的代表\Minus8 Archive\ANIMATIONS\!MP4s\Kirby and Tiff.mp4`,
		// 		func(w *multipart.Writer) {
		// 			w.WriteField(DescriptionPlain, "gaojian")
		// 			by, _ := json.Marshal([]string{"asd", "zxc"})
		// 			w.WriteField(TagsPlain, string(by))
		// 			w.WriteField("ReplacedID", "2")

		// 		})
		// 	req = initReq(post, "", bufPtr)
		// 	req.Header.Set("Content-Type", w.FormDataContentType())
		// 	return
		// }, code: 200},
		// {name: "// 携带系列", getReq: func() (req *http.Request) {
		// 	bufPtr, w := getBufAndWriter(`D:\Animation\Minus8_Archive 一个性的时代的代表\Minus8 Archive\ANIMATIONS\!MP4s\os2Q9ica2ZZR15Mg.mp4`,
		// 		func(w *multipart.Writer) {
		// 			w.WriteField(DescriptionPlain, "gaojian")
		// 			by, _ := json.Marshal([]string{"asd", "zxc"})
		// 			w.WriteField(TagsPlain, string(by))
		// 			w.WriteField(SeriesIDPlain, "1")
		// 		})
		// 	req = initReq(post, "", bufPtr)
		// 	req.Header.Set("Content-Type", w.FormDataContentType())
		// 	return
		// }, code: 200},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.getReq()
			code, _ := ServerReq(r, req)

			if code != tc.code {
				fmt.Println("not Good")
				fmt.Println(tc.name)
				fmt.Println("not Good")
			}
		})
	}
}

func TestCreateVideoSeries(t *testing.T) {
	// type VideoSeriesCreate struct {
	//     BaseOn      string  `form:"BaseOn" validate:"oneof=Series Video"`
	//     Name        string  `form:"Name" validate:"required"`
	//     Description string  `form:"Description" validate:"required"`
	//     IDsJson     string  `form:"IDsJson" validate:"required"`
	//     IDs         []int64 // 这里能指定关联的id也就是作为一个剧集该有的
	// }

	// 得进行一层的user检查

	getReq := func(query string, body io.Reader) *http.Request {
		local.TokenStorage.Store("asd", time.Now().Add(time.Minute))
		req := httptest.NewRequest(post, "/CreateVideoSeries?"+query, body)
		req.Header.Set("ContextID", "dwd")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set(constant.ContextToken, entity.ToUrlSafeBase64String(entity.ContextToken{UserID: 123, UserCode: "asd"}))
		req.Header.Set(constant.ContextLocation, entity.ToUrlSafeBase64String(entity.LocationInner{IP: "8.8.8.8", Platform: "gaojian", UserAgant: "user1"}))
		return req
	}

	r := NewRouter()
	// /GetLocation
	testCases := []struct {
		name   string
		getReq func() (req *http.Request)
		code   int
	}{
		{name: "// Series一个", getReq: func() (req *http.Request) {
			_ = entity.VideoSeriesCreate{}
			// 			 BaseOn      string  `form:"BaseOn" validate:"oneof=Series Video"`
			//     Name        string  `form:"Name" validate:"required"`
			//     Description string  `form:"Description" validate:"required"`
			//     IDsJson     string  `form:"IDs" validate:"required"`
			//     IDs         []int64 // 这里能指定关联的id也就是作为一个剧集该有的
			// }
			data := url.Values{}
			data.Set(BaseOnPlain, SeriesPlain)
			data.Set(NamePlain, "SeriesPlain18")
			data.Set(DescriptionPlain, "SeriesPlain18")
			data.Set(IDsPlain, "[18]")
			req = getReq("", strings.NewReader(data.Encode()))
			return
		}, code: 200},
		// {name: "// Series一个", getReq: func() (req *http.Request) {
		// 	_ = entity.VideoSeriesCreate{}
		// 	data := url.Values{}
		// 	data.Set(BaseOnPlain, SeriesPlain)
		// 	data.Set(NamePlain, "SeriesPlain2")
		// 	data.Set(DescriptionPlain, "SeriesPlain2")
		// 	data.Set(IDsPlain, "[17]")
		// 	req = getReq("", strings.NewReader(data.Encode()))
		// 	return
		// }, code: 200},
		// {name: "// SeriesID对应", getReq: func() (req *http.Request) {
		// 	_ = entity.VideoSeriesCreate{}
		// 	// 			 BaseOn      string  `form:"BaseOn" validate:"oneof=Series Video"`
		// 	//     Name        string  `form:"Name" validate:"required"`
		// 	//     Description string  `form:"Description" validate:"required"`
		// 	//     IDsJson     string  `form:"IDs" validate:"required"`
		// 	//     IDs         []int64 // 这里能指定关联的id也就是作为一个剧集该有的
		// 	// }
		// 	data := url.Values{}
		// 	data.Set(BaseOnPlain, SeriesPlain)
		// 	data.Set(NamePlain, "dwdwSeriesSeries1")
		// 	data.Set(DescriptionPlain, "SeriesPlain")
		// 	data.Set(IDsPlain, "[3]")
		// 	req = getReq("", strings.NewReader(data.Encode()))
		// 	return
		// }, code: 200},
		// {name: "// SeriesID不对应", getReq: func() (req *http.Request) {
		// 	_ = entity.VideoSeriesCreate{}
		// 	// 			 BaseOn      string  `form:"BaseOn" validate:"oneof=Series Video"`
		// 	//     Name        string  `form:"Name" validate:"required"`
		// 	//     Description string  `form:"Description" validate:"required"`
		// 	//     IDsJson     string  `form:"IDs" validate:"required"`
		// 	//     IDs         []int64 // 这里能指定关联的id也就是作为一个剧集该有的
		// 	// }
		// 	data := url.Values{}
		// 	data.Set(BaseOnPlain, SeriesPlain)
		// 	data.Set(NamePlain, "dwdwSeriesSeries1")
		// 	data.Set(DescriptionPlain, "SeriesPlain")
		// 	data.Set(IDsPlain, "[1,2]")
		// 	req = getReq("", strings.NewReader(data.Encode()))
		// 	return
		// }, code: 403},
		// {name: "// Video一个", getReq: func() (req *http.Request) {
		// 	data := url.Values{}
		// 	data.Set(BaseOnPlain, VideoPlain)
		// 	data.Set(NamePlain, "dwdwVideoSedwdries1awdw")
		// 	data.Set(DescriptionPlain, "SeriwdwesPlainawdw")
		// 	data.Set(IDsPlain, "[2]")
		// 	req = getReq("", strings.NewReader(data.Encode()))
		// 	return
		// }, code: 200},
		// {name: "// Video一个", getReq: func() (req *http.Request) {
		// 	data := url.Values{}
		// 	data.Set(BaseOnPlain, VideoPlain)
		// 	data.Set(NamePlain, "dwdwVidwdweoSeries2")
		// 	data.Set(DescriptionPlain, "SewdwriesPlain")
		// 	data.Set(IDsPlain, "[4]")
		// 	req = getReq("", strings.NewReader(data.Encode()))
		// 	return
		// }, code: 200},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.getReq()
			code, _ := ServerReq(r, req)

			if code != tc.code {
				fmt.Println("not Good")
				fmt.Println(tc.name)
				fmt.Println("not Good")
			}
		})
	}

}

func TestCreateVideoSeriesBatch(t *testing.T) {
	// type VideoSeriesCreate struct {
	//     BaseOn      string  `form:"BaseOn" validate:"oneof=Series Video"`
	//     Name        string  `form:"Name" validate:"required"`
	//     Description string  `form:"Description" validate:"required"`
	//     IDsJson     string  `form:"IDsJson" validate:"required"`
	//     IDs         []int64 // 这里能指定关联的id也就是作为一个剧集该有的
	// }

	// 得进行一层的user检查

	getReq := func(query string, body io.Reader) *http.Request {
		local.TokenStorage.Store("asd", time.Now().Add(time.Minute))
		req := httptest.NewRequest(post, "/CreateVideoSeries?"+query, body)
		req.Header.Set("ContextID", "dwd")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set(constant.ContextToken, entity.ToUrlSafeBase64String(entity.ContextToken{UserID: 123, UserCode: "asd"}))
		req.Header.Set(constant.ContextLocation, entity.ToUrlSafeBase64String(entity.LocationInner{IP: "8.8.8.8", Platform: "gaojian", UserAgant: "user1"}))
		return req
	}

	r := NewRouter()

	getReq1 := func(index int) (req *http.Request) {
		data := url.Values{}
		data.Set(BaseOnPlain, VideoPlain)
		data.Set(NamePlain, fmt.Sprint("dwdwVideoSedwdries1awdw", index))
		data.Set(DescriptionPlain, fmt.Sprint("SeriwdwesPlainawdw", index))
		data.Set(IDsPlain, fmt.Sprintf("[%d]", index))
		req = getReq("", strings.NewReader(data.Encode()))
		return
	}

	for i := 1; i < 7; i++ {
		req := getReq1(i)
		ServerReq(r, req)
	}

}

func TestUpdateVideoSeries(t *testing.T) {

	getReq := func(query string, body io.Reader) *http.Request {
		local.TokenStorage.Store("asd", time.Now().Add(time.Minute))
		req := httptest.NewRequest(post, "/UpdateVideoSeries?"+query, body)
		req.Header.Set("ContextID", "dwd")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set(constant.ContextToken, entity.ToUrlSafeBase64String(entity.ContextToken{UserID: 123, UserCode: "asd"}))
		req.Header.Set(constant.ContextLocation, entity.ToUrlSafeBase64String(entity.LocationInner{IP: "8.8.8.8", Platform: "gaojian", UserAgant: "user1"}))
		return req
	}

	r := NewRouter()
	// /GetLocation
	testCases := []struct {
		name   string
		getReq func() (req *http.Request)
		code   int
	}{
		// {name: "// 基本字段更新", getReq: func() (req *http.Request) {
		// 	_ = entity.VideoSeriesUpdate{}

		// 	data := url.Values{}
		// 	data.Set(IDPlain, "1")
		// 	data.Set(DescriptionPlain, "SeriesPlain// 基本字段更新")
		// 	data.Set(NamePlain, "dwdwSeriesSeries3")
		// 	req = getReq("", strings.NewReader(data.Encode()))
		// 	return
		// }, code: 200},
		// {name: "// 增加并删除video下层关联", getReq: func() (req *http.Request) {
		// 	_ = entity.VideoSeriesUpdate{}

		// 	data := url.Values{}
		// 	data.Set(IDPlain, "10")
		// 	data.Set(DescriptionPlain, "SeriesPlain// 基本字段更新,非delete增加并删除video关联")
		// 	data.Set(NamePlain, "dwdwSeriesSeries3,非delete增加并删除video关联")
		// 	data.Set(IDsAppendPlain, "[13]")
		// 	data.Set(IDsRemovePlain, "[1,2]")
		// 	req = getReq("", strings.NewReader(data.Encode()))
		// 	return
		// }, code: 200},

		// {name: "// 增加并删除video下层关联, 错误layer不一样", getReq: func() (req *http.Request) {
		// 	_ = entity.VideoSeriesUpdate{}

		// 	data := url.Values{}
		// 	data.Set(IDPlain, "23")
		// 	data.Set(DescriptionPlain, "SeriesPlain// 基本字段更新,非delete增加并删除video关联")
		// 	data.Set(NamePlain, "dwdwSeriesSeries3,非delete增加并删除video关联")
		// 	data.Set(IDsAppendPlain, "[24]")
		// 	data.Set(IDsRemovePlain, "[21]")
		// 	req = getReq("", strings.NewReader(data.Encode()))
		// 	return
		// }, code: 200},

		// {name: "// 非delete增加并删除series关联", getReq: func() (req *http.Request) {
		// 	_ = entity.VideoSeriesUpdate{}

		// 	data := url.Values{}
		// 	data.Set(IDPlain, "1")
		// 	data.Set(DescriptionPlain, "SeriesPlain// 基本字段更新")
		// 	data.Set(NamePlain, "dwdwSeriesSeries3")
		// 	req = getReq("", strings.NewReader(data.Encode()))
		// 	return
		// }, code: 200},
		// {name: "// 删除当前", getReq: func() (req *http.Request) {
		// 	_ = entity.VideoSeriesUpdate{}

		// 	data := url.Values{}
		// 	data.Set(IDPlain, "19")
		// 	data.Set(DeletePlain, "true")

		// 	req = getReq("", strings.NewReader(data.Encode()))
		// 	return
		// }, code: 200},
		{name: "// 级联删除当前", getReq: func() (req *http.Request) {
			_ = entity.VideoSeriesUpdate{}

			data := url.Values{}
			data.Set(IDPlain, "17")
			data.Set("IsCASCADE", "true")
			data.Set("Delete", "true")

			req = getReq("", strings.NewReader(data.Encode()))
			return
		}, code: 200},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.getReq()
			code, _ := ServerReq(r, req)

			if code != tc.code {
				fmt.Println("not Good")
				fmt.Println(tc.name)
				fmt.Println("not Good")
			}
		})
	}
}

func TestAppend(t *testing.T) {
	idsAppend := make([]int, 0, 10)
	idsAppend = append(idsAppend, 1, 2, 3)
	idsRemove := []int{4, 5, 6}
	ids := append(idsAppend, idsRemove...)
	fmt.Println(ids)
	fmt.Println(idsAppend)
}

func Add(column string) string {
	return column + " - ?"
}

func Subtract(column string) string {
	return column + " - ?"
}

func TestEx(t *testing.T) {
	_ = NewRouter()
	err := base.Mysql.Model(&entity.Series{}).
		Where("ID = ?", 24).Update(SumViewsCountPlain, gorm.Expr(Subtract(SumViewsCountPlain), 1)).Error
	if err != nil {
		fmt.Println(err)
	}
}

func TestAddVideosViews(t *testing.T) {
	getReq := func(query string, body io.Reader) *http.Request {
		local.TokenStorage.Store("asd", time.Now().Add(time.Minute))
		req := httptest.NewRequest(get, "/AddVideosViews?"+query, body)
		req.Header.Set("ContextID", "dwd")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set(constant.ContextToken, entity.ToUrlSafeBase64String(entity.ContextToken{UserID: 9, UserCode: "asd"}))
		req.Header.Set(constant.ContextLocation, entity.ToUrlSafeBase64String(entity.LocationInner{IP: "8.8.8.8", Platform: "gaojian", UserAgant: "user1"}))
		return req
	}

	r := NewRouter()
	// /GetLocation
	testCases := []struct {
		name   string
		getReq func() (req *http.Request)
		code   int
	}{
		// {name: "// 级联删除当前", getReq: func() (req *http.Request) {
		// 	_ = entity.VideoSeriesUpdate{}

		// 	data := url.Values{}
		// 	data.Set(IDPlain, "17")
		// 	data.Set("IsCASCADE", "true")
		// 	data.Set("Delete", "true")

		// 	req = getReq("", strings.NewReader(data.Encode()))
		// 	return
		// }, code: 200},
		{
			name: "normal", getReq: func() (req *http.Request) {
				req = getReq(VideoIDPlain+"=4", nil)
				return
			}, code: 200,
		},
		// {
		// 	name: "no Video", getReq: func() (req *http.Request) {
		// 		req = getReq(VideoIDPlain+"=100", nil)
		// 		return
		// 	}, code: 200,
		// },
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.getReq()
			code, _ := ServerReq(r, req)

			if code != tc.code {
				fmt.Println("not Good")
				fmt.Println(tc.name)
				fmt.Println("not Good")
			}
		})
	}
	time.Sleep(time.Second * 2)
}

func TestPostComment(t *testing.T) {
	getReq := func(query string, body io.Reader, userID int64) *http.Request {
		local.TokenStorage.Store("asd", time.Now().Add(time.Minute))
		req := httptest.NewRequest(post, "/PostComment?"+query, body)
		req.Header.Set("ContextID", "dwd")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set(constant.ContextToken, entity.ToUrlSafeBase64String(entity.ContextToken{UserID: userID, UserCode: "asd"}))
		req.Header.Set(constant.ContextLocation, entity.ToUrlSafeBase64String(entity.LocationInner{IP: "8.8.8.8", Platform: "gaojian", UserAgant: "user1"}))
		return req
	}

	r := NewRouter()
	// /GetLocation
	testCases := []struct {
		name   string
		getReq func() (req *http.Request)
		code   int
	}{
		// {name: "// 级联删除当前", getReq: func() (req *http.Request) {
		// 	_ = entity.VideoSeriesUpdate{}

		// 	data := url.Values{}
		// 	data.Set(IDPlain, "17")
		// 	data.Set("IsCASCADE", "true")
		// 	data.Set("Delete", "true")

		// 	req = getReq("", strings.NewReader(data.Encode()))
		// 	return
		// }, code: 200},
		// {
		// 	name: "normal", getReq: func() (req *http.Request) {
		// 		_ = entity.Comment{}
		// 		_ = entity.PostComment{}
		// 		data := url.Values{}
		// 		data.Set(VideoIDPlain, "4")
		// 		data.Set(CommentPlain, "gaojian666")
		// 		req = getReq("", strings.NewReader(data.Encode()), 9)
		// 		return
		// 	}, code: 200,
		// },
		// {
		// 	name: "normal child", getReq: func() (req *http.Request) {
		// 		_ = entity.Comment{}
		// 		_ = entity.PostComment{}
		// 		data := url.Values{}
		// 		data.Set(ParentIDPlain, "3")
		// 		data.Set(UserID2Plain, "9")
		// 		data.Set(VideoIDPlain, "4")
		// 		data.Set(CommentPlain, "guyuchun666")

		// 		req = getReq("", strings.NewReader(data.Encode()), 11)
		// 		return
		// 	}, code: 200,
		// },

		{
			name: "评论里没有这个人", getReq: func() (req *http.Request) {
				data := url.Values{}
				data.Set(ParentIDPlain, "3")
				data.Set(UserID2Plain, "123")
				data.Set(VideoIDPlain, "4")
				data.Set(CommentPlain, "guyuchun666")

				req = getReq("", strings.NewReader(data.Encode()), 11)
				return
			}, code: 400,
		},

		{
			name: "没有这个parent", getReq: func() (req *http.Request) {
				data := url.Values{}
				data.Set(ParentIDPlain, "100")
				data.Set(UserID2Plain, "9")
				data.Set(VideoIDPlain, "4")
				data.Set(CommentPlain, "guyuchun666")

				req = getReq("", strings.NewReader(data.Encode()), 11)
				return
			}, code: 404,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.getReq()
			code, _ := ServerReq(r, req)

			if code != tc.code {
				fmt.Println("not Good")
				fmt.Println(tc.name)
				fmt.Println("not Good")
			}
		})
	}
	time.Sleep(time.Second * 2)
}

func TestDeleteComment(t *testing.T) {
	getReq := func(query string, body io.Reader, userID int64) *http.Request {
		local.TokenStorage.Store("asd", time.Now().Add(time.Minute))
		req := httptest.NewRequest(get, "/DeleteComment?"+query, body)
		req.Header.Set("ContextID", "dwd")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set(constant.ContextToken, entity.ToUrlSafeBase64String(entity.ContextToken{UserID: userID, UserCode: "asd"}))
		req.Header.Set(constant.ContextLocation, entity.ToUrlSafeBase64String(entity.LocationInner{IP: "8.8.8.8", Platform: "gaojian", UserAgant: "user1"}))
		return req
	}

	r := NewRouter()
	// /GetLocation
	testCases := []struct {
		name   string
		getReq func() (req *http.Request)
		code   int
	}{

		{
			name: "是不是级联删除", getReq: func() (req *http.Request) {
				req = getReq(CommentIDPlain+"=3", nil, 9)
				return
			}, code: 200,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.getReq()
			code, _ := ServerReq(r, req)

			if code != tc.code {
				fmt.Println("not Good")
				fmt.Println(tc.name)
				fmt.Println("not Good")
			}
		})
	}
}

//

func TestOther(t *testing.T) {
	fmt.Println("gaojian")
	getReq := func(method, path, query string, body io.Reader, userID int64) *http.Request {
		local.TokenStorage.Store("asd", time.Now().Add(time.Minute))
		req := httptest.NewRequest(method, "/"+path+"?"+query, body)
		req.Header.Set("ContextID", "dwd")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set(constant.ContextToken, entity.ToUrlSafeBase64String(entity.ContextToken{UserID: userID, UserCode: "asd"}))
		req.Header.Set(constant.ContextLocation, entity.ToUrlSafeBase64String(entity.LocationInner{IP: "8.8.8.8", Platform: "gaojian", UserAgant: "user1"}))
		return req
	}

	r := NewRouter()
	// /GetLocation
	testCases := []struct {
		name   string
		getReq func() *http.Request
		code   int
	}{
		// {
		// 	// name: "简单过一下UserViewVideos", getReq: func() (req *http.Request) {
		// 	// 	req = getReq(get, "UserViewVideos", "VideoID"+"=5", nil, 9)
		// 	// 	return
		// 	// }, code: 200,
		// },
		// {
		// 	name: "UserActiveAdd", getReq: func() *http.Request {
		// 		return getReq(get, "UserActiveAdd", "", nil, 9)
		// 	}, code: 200,
		// },
		// {
		// 	name: "UserActiveGet", getReq: func() *http.Request {
		// 		return getReq(get, "UserActiveGet", UserIDPlain+"=9", nil, 9)
		// 	}, code: 200,
		// },
		// {
		// 	name: "FollowOther", getReq: func() *http.Request {
		// 		return getReq(get, "FollowOther", UserIDPlain+"=9", nil, 11)
		// 	}, code: 200,
		// },
		// {
		// 	name: "FollowOther", getReq: func() *http.Request {
		// 		return getReq(get, "FollowOther", UserIDPlain+"=11", nil, 9)
		// 	}, code: 200,
		// },
		// {
		// 	name: "DeFollowOther", getReq: func() *http.Request {
		// 		return getReq(get, "DeFollowOther", UserIDPlain+"=9", nil, 11)
		// 	}, code: 200,
		// },
		{
			name: "CollectingVideo", getReq: func() *http.Request {
				return getReq(get, "CollectingVideo", VideoIDPlain+"=4", nil, 11)
			}, code: 200,
		},

		// {
		// 	name: "DecollectingVideo", getReq: func() *http.Request {
		// 		return getReq(get, "DecollectingVideo", VideoIDPlain+"=4", nil, 11)
		// 	}, code: 200,
		// },
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.getReq()
			code, _ := ServerReq(r, req)

			if code != tc.code {
				fmt.Println("not Good")
				fmt.Println(tc.name)
				fmt.Println("not Good")
			}
		})
	}
	time.Sleep(time.Second * 2)
}

func TestSearchUserSecondTable(t *testing.T) {

	// LoginLocations
	// Videos
	// CollectedVideos
	fmt.Println("gaojian")
	// getReq := func(query string, body io.Reader, userID int64) *http.Request {
	// 	local.TokenStorage.Store("asd", time.Now().Add(time.Minute))
	// 	req := httptest.NewRequest(get, "/SearchUserSecondTable?"+query, body)
	// 	req.Header.Set("ContextID", "dwd")
	// 	// req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// 	req.Header.Set(constant.ContextToken, entity.ToUrlSafeBase64String(entity.ContextToken{UserID: userID, UserCode: "asd"}))
	// 	req.Header.Set(constant.ContextLocation, entity.ToUrlSafeBase64String(entity.LocationInner{IP: "8.8.8.8", Platform: "gaojian", UserAgant: "user1"}))
	// 	return req
	// }

	r := NewRouter()
	// /GetLocation
	testCases := []struct {
		name   string
		getReq func() *http.Request
		code   int
	}{
		// {
		// 	name: "LoginLocations", getReq: func() *http.Request {
		// 		_ = entity.UserV1SearchSecondTable{}
		// 		return getReq(SearchTypePlain+"=LoginLocations", nil, 11)
		// 	}, code: 200,
		// },

		// {
		// 	name: "Videos", getReq: func() *http.Request {
		// 		_ = entity.UserV1SearchSecondTable{}
		// 		return getReq(SearchTypePlain+"=Videos", nil, 9)
		// 	}, code: 200,
		// },

		// {
		// 	name: "LoginLocations", getReq: func() *http.Request {
		// 		_ = entity.UserV1SearchSecondTable{}
		// 		return getReq(SearchTypePlain+"=LoginLocations", nil, 11)
		// 	}, code: 200,
		// },

		// {
		// 	name: "CollectedVideosPlain", getReq: func() *http.Request {
		// 		_ = entity.UserV1SearchSecondTable{}
		// 		return getReq(SearchTypePlain+"=CollectedVideos", nil, 11)
		// 	}, code: 200,
		// },

		// {
		// 	name: "FollowingPlain", getReq: func() *http.Request {
		// 		_ = entity.UserV1SearchSecondTable{}
		// 		return getReq(SearchTypePlain+"=Following", nil, 11)
		// 	}, code: 200,
		// },

		// {
		// 	name: "FollowingPlain", getReq: func() *http.Request {
		// 		_ = entity.UserV1SearchSecondTable{}
		// 		return getReq(SearchTypePlain+"=Followers", nil, 11)
		// 	}, code: 200,
		// },
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.getReq()
			code, _ := ServerReq(r, req)

			if code != tc.code {
				fmt.Println("not Good")
				fmt.Println(tc.name)
				fmt.Println("not Good")
			}
		})
	}

}

func TestRandomRecommendVideo(t *testing.T) {
	// getReq := func(query string, body io.Reader, userID int64) *http.Request {
	// 	local.TokenStorage.Store("asd", time.Now().Add(time.Minute))
	// 	req := httptest.NewRequest(get, "/RandomRecommendVideo?"+query, body)
	// 	req.Header.Set("ContextID", "dwd")
	// 	// req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// 	if userID != 0 {
	// 		req.Header.Set(constant.ContextToken, entity.ToUrlSafeBase64String(entity.ContextToken{UserID: userID, UserCode: "asd"}))
	// 	}
	// 	req.Header.Set(constant.ContextLocation, entity.ToUrlSafeBase64String(entity.LocationInner{IP: "8.8.8.8", Platform: "gaojian", UserAgant: "user1"}))
	// 	return req
	// }

	r := NewRouter()
	// /GetLocation
	testCases := []struct {
		name   string
		getReq func() *http.Request
		code   int
	}{
		// {
		// 	name: "RandomRecommendVideo", getReq: func() *http.Request {
		// 		_ = entity.UserV1SearchSecondTable{}
		// 		return getReq("", nil, 9)
		// 	}, code: 200,
		// },
		// {
		// 	name: "RandomRecommendVideo 大众池子", getReq: func() *http.Request {
		// 		_ = entity.UserV1SearchSecondTable{}
		// 		return getReq("", nil, 0)
		// 	}, code: 200,
		// },
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.getReq()
			code, _ := ServerReq(r, req)

			if code != tc.code {
				fmt.Println("not Good")
				fmt.Println(tc.name)
				fmt.Println("not Good")
			}
		})
	}
	abc := local.VideoBuf
	fmt.Println(abc)
}

func TestBatchUsers(*testing.T) {
	NewRouterAndInitConf("./conf-mysql.yaml")
	// config.Init("./conf-mysql.yaml")
	mdb := base.Mysql
	usernames := []string{
		"qwe",
		"asd",
		"zxc",
		"rty",
		"fgh",
		"vbn",
		"uio",
		"jkl",
	}
	users := []entity.UserV1{}
	for i, v := range usernames {
		ip := net.ParseIP("8.8.8.8")
		record, _ := base.GeoipLoc.City(ip)
		city := record.City.Names["en"]
		aLoc := entity.ContextLocation{
			Location: city,
			LocationPlain: entity.LocationInner{
				IP:        "8.8.8.8",
				UserAgant: v,
				Platform:  v,
			},
		}
		aUser := entity.UserV1{
			ID:                  int64(i + 1),
			Name:                v,
			PasswordPlainString: v,
			PhonePlainString:    fmt.Sprintf("+86123456789123%d", i),
			EmailPlainString:    fmt.Sprintf("%s@abc.com", v),
			Status:              "happy",
			LoginLocations: []entity.LoginLocation{
				{Location: aLoc.Location, CodePlainText: base.SprintfSortStrings(aLoc.LocationPlain.IP, aLoc.LocationPlain.UserAgant, aLoc.LocationPlain.Platform)},
			},
		}
		users = append(users, *aUser.EncodeDB())
	}
	err := mdb.CreateInBatches(&users, len(users)).Error
	if err != nil {
		fmt.Println(err)
		return
	}
}

func TestAbc(*testing.T) {
	NewRouterAndInitConf("./conf-mysql.yaml")
	// config.Init("./conf-mysql.yaml")
	mdb := base.Mysql
	aSeries := entity.Series{
		ID:          1,
		UserID:      1,
		Layer:       0,
		Name:        "epic",
		Description: "epic",
		VideosCount: 2000,
		ViewsCount:  987,
	}
	err := mdb.Create(&aSeries).Error
	if err != nil {
		fmt.Println(err)
		return
	}

}

// func TestVideos(*testing.T) {
// 	dir := `D:\Data\Musical\Epic-slice\video`
// 	entries, err := os.ReadDir(dir)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	// 构建一堆video
// 	videos := []entity.VideoES{}
// 	choices := []string{"zxc", "asd", "qwe", "vbn", "fgh", "rty", "jkl", "uio"}
// 	var seriesID int64 = 1
// 	for i := 0; i < 100; i++ {
// 		curTags := []string{}
// 		for i := 0; i < 2; i++ {
// 			abc := choices[rand.Intn(len(choices))]
// 			if i == 1 && abc == curTags[0] {
// 				continue
// 			}
// 			curTags = append(curTags, abc)
// 		}
// 		videos = append(videos, entity.VideoES{
// 			SeriesID:      &seriesID,
// 			SeriesName:    "epic",
// 			ID:            int64(i),
// 			UserID:        int64(i%8 + 1),
// 			Name:          entries[i%len(entries)].Name(),
// 			Duration:      float64(i),
// 			Description:   entries[i%len(entries)].Name(),
// 			Path:          filepath.Join(dir, entries[i%len(entries)].Name(), "stream_0.m3u8"),
// 			ViewsCount:    int64(i),
// 			Tags:          curTags,
// 			CreatedAt:     time.Date(2026, 5, rand.Intn(32), rand.Intn(25), rand.Intn(61), rand.Intn(61), 0, time.UTC),,
// 			CommentsCount: int64(i),
// 		})
// 	}

// 	// for _, v := range entries {
// 	// 	fmt.Println(v.Name())
// 	// }
// }
