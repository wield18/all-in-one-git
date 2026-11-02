package config

import (
	"all-in-one/pkg/ekit/captcha"
	"all-in-one/pkg/ekit/slogx"
	"all-in-one/pkg/emailqq"
	"all-in-one/pkg/es"
	"all-in-one/pkg/mysqlx"
	"all-in-one/pkg/redisx"
	"os"

	"go.yaml.in/yaml/v2"
)

// 全局引用
var Conf = &Config{}

// //go:embed config.yaml
// var configData []byte

// 配置初始化 改成手动调用
func Init(confPath string) {
	by, err := os.ReadFile(confPath)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(by, Conf)
	if err != nil {
		panic(err)
	}
}

// 总配文件
type Config struct {
	Server   Server              `yaml:"Server"`
	Keys     Keys                `yaml:"Keys"`
	TheOther TheOther            `yaml:"TheOther"`
	Mysql    mysqlx.Mysql        `yaml:"Mysql"`
	Redis    redisx.Redis        `yaml:"Redis"`
	QQ       emailqq.QQEmail     `yaml:"QQ"`
	CaptCha  captcha.CaptChaConf `yaml:"Captcha"`
	Slog     slogx.SlogConf      `yaml:"Slog"`
	ES       es.ESConfigSingle   `yaml:"ES"`
}

// 项目端口配置
type Server struct {
	Port         string `yaml:"Port"`
	Address      string `yaml:"Address"`
	Model        string `yaml:"Model"`
	ServerName   string `yaml:"ServerName"`
	FrontAddress string `yaml:"FrontAddress"`
}

type Keys struct {
	AesGcmDbKey string `yaml:"AES-GCM-DB"`
	HMAS        string `yaml:"HMAS"`
	CryptoKey   string `yaml:"CryptoKey"`
}

type TheOther struct {
	Geoip2MMDB       string `yaml:"Geoip2MMDB"` // 一个简单的ip位置服务只是需要读位置
	SuperID          string `yaml:"SuperID"`
	StoragePath      string `yaml:"StoragePath"`
	StorageStatePath string `yaml:"StorageStatePath"`
}
