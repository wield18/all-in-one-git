package entity

import (
	"all-in-one/internal/base"
	"all-in-one/internal/base/constant"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net"
)

var (
	emptyStr        = "empty str"
	lackOfArguments = "lack of arguments"
	typeErr         = "typeErr"
)

func errInvalidCode(info string) error {
	return errors.New("invalid code: " + info)
}

// rawCode = urlBase64(json(ContextToken))
func ToContextToken(id string, rawCode string) (ContextToken, error) {
	if id == "" {
		return ContextToken{}, errors.New(emptyStr)
	}
	if rawCode == "" { // 可以userToken不存在
		return ContextToken{ID: id}, nil
	}
	// base.HashCryptoCode(userID) 这样加密的一个string
	// 进行aes解密

	rawCodeByte, err := base64.RawURLEncoding.DecodeString(rawCode)
	if err != nil {
		return ContextToken{}, err
	}
	contextToken := ContextToken{}
	err = json.Unmarshal(rawCodeByte, &contextToken)
	if err != nil {
		return ContextToken{}, err
	}
	err = contextToken.checkUserColumn()
	if err != nil {
		return ContextToken{}, err
	}
	return contextToken, nil
}

// 所有字段必须非空存在
func ToContextLocation(rawCode string) (ContextLocation, error) {
	if rawCode == "" {
		return ContextLocation{}, errors.New(emptyStr)
	}
	// base64一下
	rawCodeByte, err := base64.RawURLEncoding.DecodeString(rawCode)
	if err != nil {
		return ContextLocation{}, err
	}
	// 解析json字符串
	aMap, err := base.JsonMarshalAndCheckEmpty(rawCodeByte, constant.IPPlain, constant.UserAgantPlain, constant.PlatformPlain)
	if err != nil {
		return ContextLocation{}, errInvalidCode(err.Error())
	}

	ip := net.ParseIP(aMap[constant.IPPlain])
	record, _ := base.GeoipLoc.City(ip)
	Location := record.City.Names["en"]
	return ContextLocation{
		Location: Location,
		LocationPlain: LocationInner{
			IP:        aMap[constant.IPPlain],
			UserAgant: aMap[constant.UserAgantPlain],
			Platform:  aMap[constant.PlatformPlain],
		},
	}, nil
}

// 仅仅只解构,不验证
func ToCaptchaPtr(rawCode string, captchaVal string) (*Captcha, error) {
	if rawCode == "" || captchaVal == "" {
		return nil, errors.New(emptyStr)
	}

	abc := &Captcha{}
	// urlBase64
	rawCodeByte, err := base64.RawURLEncoding.DecodeString(rawCode)
	if err != nil {
		return nil, err
	}
	// 解析json字符串
	json.Unmarshal(rawCodeByte, abc)
	abc.Val = captchaVal
	err = base.V.Struct(abc)
	if err != nil {
		return nil, errInvalidCode(err.Error())
	}
	return abc, nil

}
