// 验证码图片
package captcha

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/mojocn/base64Captcha"
)

type CaptChaConf struct {
	Height   int     `yaml:"Height"`
	Width    int     `yaml:"Width"`
	Length   int     `yaml:"Length"`
	MaxSkew  float64 `yaml:"MaxSkew"`
	DotCount int     `yaml:"DotCount"`
}

func NewCaptcha(conf CaptChaConf) *base64Captcha.Captcha {
	if conf.Height == 0 {
		panic("conf.Height is empty parameter")
	}
	if conf.Width == 0 {
		panic("conf.Width is empty parameter")
	}
	if conf.Length == 0 {
		panic("conf.Length is empty parameter")
	}
	if conf.MaxSkew == 0 {
		panic("conf.MaxSkew is empty parameter")
	}
	if conf.DotCount == 0 {
		panic("conf.DotCount is empty parameter")
	}
	var store = base64Captcha.DefaultMemStore
	var driver = base64Captcha.NewDriverDigit(
		conf.Height,
		conf.Width,
		conf.Length,
		conf.MaxSkew,
		conf.DotCount,
	)
	var CaptchaGenerator = base64Captcha.NewCaptcha(driver, store)
	return CaptchaGenerator
}

// 本地存储
func SaveBase64Image(b64s, filename string) error {
	// 1. 去掉前缀
	parts := strings.SplitN(b64s, ",", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid base64 image format")
	}

	// 2. Base64 解码
	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return err
	}

	// 3. 写入文件
	return os.WriteFile(filename, data, 0644)
}
