package entity

import (
	"all-in-one/internal/base"
	"all-in-one/internal/base/constant"
	"errors"
)

var (
	zeroField = "zeroField"
)

// KeyValue 可测试可基本使用
type KeyValue struct {
	Key   string `validate:"required"`
	Value string `validate:"required"`
}

// 上下文token
type ContextToken struct {
	ID string `json:"-"` // 这网关的直接先忽略
	// 这里必须使用crypto加密的token再跟对应的id进行cipher.AEAD的aes加密,传输时加密解密形成俩个字段
	UserID   int64
	UserCode string
}

func (c *ContextToken) checkUserColumn() error {
	if c.UserID == 0 {
		return errors.New(base.Sprintf(zeroField, constant.UserIDPlain))
	}
	if c.UserCode == "" {
		return errors.New(base.Sprintf(zeroField, constant.UserCodePlain))
	}
	return nil
}

// 说白了location我只是需要个code就行
// 需要使用aes进行加密一下,该解密就解开
type ContextLocation struct {
	// Location  string record, _ := base.GeoipLoc.City(ip)
	// Location: record.City.Names["en"],
	Location      string
	LocationPlain LocationInner
}

// 申请locationCode时使用
// 这里的得跟constant里对应
type LocationInner struct {
	IP        string `form:"IP" validate:"ip4_addr"`
	UserAgant string `form:"UserAgant" validate:"required"`
	Platform  string `form:"Platform" validate:"required"`
}

func (l *LocationInner) SprintfSortStrings() string {
	return base.SprintfSortStrings(l.IP, l.UserAgant, l.Platform)
}

// 这里的得跟constant里对应
// 这里的json主要是返回使用或者绑定
type Captcha struct {
	Type      string `json:"Type" validate:"oneof=CaptchaB64s CaptchaEmail CaptchaPhone"`
	Code      string `json:"Code" validate:"required"` // base.CipherDB.Encrypt(email or phone) or randx.RandCode
	Val       string `json:"-" validate:"required"`
	IsChecked bool   `json:"-"`
}
