package base

import (
	"all-in-one/config"
	"all-in-one/pkg/consumer"
	"all-in-one/pkg/ekit/captcha"
	"all-in-one/pkg/ekit/cipherx"
	"all-in-one/pkg/ekit/slogx"
	"all-in-one/pkg/emailqq"
	"all-in-one/pkg/es"
	"all-in-one/pkg/mysqlx"
	"all-in-one/pkg/redisx"
	"all-in-one/pkg/storage"
	"path/filepath"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/go-playground/validator/v10"
	"github.com/mojocn/base64Captcha"
	"github.com/oschwald/geoip2-golang"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// 所有的服务

// base里拿不到local的引用,也初始不了
// 那里边的服务得主文件调用初始化

var (
	Mysql          *gorm.DB
	Redis          *redis.Client
	QQDialer       *emailqq.QQDialer
	B64sCaptcha    *base64Captcha.Captcha
	SaveSlogState  func() // slog使用的最后state的save
	hmacKey        = []byte{}
	cryptoKey      = "" // CryptoSait
	SuperID        = ""
	CipherDB       *cipherx.Cipher // 存库使用
	GeoipLoc       *geoip2.Reader
	StorageManager *storage.StorageManager // 得主动调用里边的 SaveState
	VideoStorage   *storage.FileStorage
	PicStorage     *storage.FileStorage
	ES             *elasticsearch.TypedClient
	// 这三必须主文件指定
	// R *gin.Engine
	P *consumer.ConsumerPool // 需要主动调shutdown正常关闭
	V *validator.Validate
)

func InitAll() {
	var err error
	SaveSlogState = slogx.InitSlog(config.Conf.Slog)
	hmacKey, cryptoKey, CipherDB = NewKeys(config.Conf.Keys)
	Mysql = mysqlx.NewDB(config.Conf.Mysql)
	Redis = redisx.NewRedis(config.Conf.Redis)
	ES = es.NewESClient(config.Conf.ES)
	QQDialer = emailqq.NewDialer(config.Conf.QQ)
	B64sCaptcha = captcha.NewCaptcha(config.Conf.CaptCha)
	GeoipLoc, err = geoip2.Open(config.Conf.TheOther.Geoip2MMDB)
	SuperID = config.Conf.TheOther.SuperID

	VideoStorage = storage.NewFileStorage(filepath.Join(config.Conf.TheOther.StoragePath, "Video"))
	PicStorage = storage.NewFileStorage(filepath.Join(config.Conf.TheOther.StoragePath, "Pic"))
	StorageManager = storage.NewStorageManagerFromFileAndMap(
		config.Conf.TheOther.StorageStatePath,
		map[string]storage.StorageInter{
			"Video": VideoStorage,
			"Pic":   PicStorage,
		})
	if err != nil {
		panic(err)
	}
	// 简单服务

	// 下边这俩依赖前边的服务, 下边的不能在base直接初始化,得在主文件中指定
	// R = router.NewRouter(config.Conf.Server)
	// P = consumerouter.NewConsumePool()
}

func SetRPV(p *consumer.ConsumerPool, v *validator.Validate) {
	if p != nil {
		P = p
	}
	if v != nil {
		V = v
	}

}
