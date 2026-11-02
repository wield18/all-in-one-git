package captcha

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
)

type abc struct {
	// === 分组1：二选一 (Phone 和 Email 至少填一个) ===
	Phone string `validate:"required_without=Email"`
	Email string `validate:"required_without=Phone"`

	// === 分组2：三选一 (Wechat, QQ, Telegram 至少填一个) ===
	Wechat   string `validate:"required_without_all=QQ Telegram"`
	QQ       string `validate:"required_without_all=Wechat Telegram"`
	Telegram string `validate:"required_without_all=Wechat QQ"`
}

func TestValidate(t *testing.T) {
	a := abc{
		Wechat: "rfgrgr",
	}
	err := validator.New().Struct(a)
	if err != nil {
		fmt.Println(err)
	}
}

func TestCap(t *testing.T) {
	abc := make([]int, 0, 10)
	for _, v := range abc {
		fmt.Println(v)
	}
	fmt.Println(abc)
	abc = append(abc, 1)
	fmt.Println(abc)
}
