package mysqlx

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
    "log/slog"
)

// 数据库配置
type Mysql struct {
	Dialects string `yaml:"Dialects"`
	Host     string `yaml:"Host"`
	Port     int    `yaml:"Port"`
	DB       string `yaml:"DB"`

	Username string `yaml:"Username"`
	Password string `yaml:"Password"`
	Charset  string `yaml:"Charset"`
	MaxIdle  int    `yaml:"MaxIdle"`
	MaxOpen  int    `yaml:"MaxOpen"`
}

// 新创建一个 *gorm.DB 实例
func NewDB(dbConfig Mysql) *gorm.DB {
	url := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		dbConfig.Username,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.DB,
		dbConfig.Charset)

	// 使用GORM打开MySQL数据库连接
	db, err := gorm.Open(mysql.Open(url), &gorm.Config{
		// 设置日志模式为Info级别，显示执行的SQL语句
		Logger: NewSlogLogger(slog.Default()),
	})

	// 检查连接错误
	if err != nil {
		panic(err)
	}

	// 检查GORM错误
	if db.Error != nil {
		panic(db.Error)
	}

	// 获取底层的sql.DB对象以设置连接池参数
	sqlDB, _ := db.DB()

	// 设置连接池参数
	// SetMaxIdleConns: 设置最大空闲连接数
	sqlDB.SetMaxIdleConns(dbConfig.MaxIdle)
	// SetMaxOpenConns: 设置最大打开连接数
	sqlDB.SetMaxOpenConns(dbConfig.MaxOpen)
	// log.Println("数据库连接正常")

	return db
}
