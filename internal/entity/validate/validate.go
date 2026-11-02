package validate

import (
	"all-in-one/internal/entity"
	"reflect"
	"regexp"

	"github.com/go-playground/validator/v10"
)

var punctRegex = regexp.MustCompile(`[!\"#$%&'()*+,\-./:;<=>?@[\\\]^_{|}~，。！？；：·、]`)

// 所有tag
var (
	noPunct = "noPunct"
	// exactlyOne = "exactlyOne"
)

func New() *validator.Validate {
	v := validator.New()
	register(v)
	return v
}

func register(v *validator.Validate) {
	v.RegisterValidation(noPunct, noPunctFunc)
	v.RegisterStructValidation(validateExactlyOne, entity.UserV1Update{})
}

// noPunct 校验字符串是否不包含标点，没有任何标点返回 true
func noPunctFunc(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	if len(str) == 0 {
		return false
	}
	return !punctRegex.MatchString(str)
}

type fieldEntry struct {
	name string
	val  reflect.Value
}

// validateExactlyOne 是一个“通用”结构体级校验。
// 它扫描当前结构体中所有带 exactlyone 标签的字段，
// 同一分组内有值的字段恰好为 1 个则通过。
func validateExactlyOne(sl validator.StructLevel) {
	v := sl.Current()
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	// 按标签值分组

	groups := make(map[string][]fieldEntry)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("exactlyone")
		if tag == "" || tag == "-" {
			continue
		}
		groups[tag] = append(groups[tag], fieldEntry{field.Name, v.Field(i)})
	}

	for groupName, fields := range groups {
		count := 0
		for _, fv := range fields {
			if hasValue(fv.val) {
				count++
			}
		}
		if count != 1 {
			// 把错误挂到该组第一个字段上
			sl.ReportError(fields[0].val, // reflect.Value
				fields[0].name, // 字段名，如 "Email"
				fields[0].name, // 结构体字段名
				"exactlyone",
				groupName)
		}
	}
}

// hasValue 判断一个 reflect.Value 是否"有值"。
// 覆盖字符串、数字、指针、切片、map、接口等常见类型。
func hasValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() > 0
	case reflect.Ptr, reflect.Interface:
		return !v.IsNil()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return v.Float() != 0
	case reflect.Bool:
		return v.Bool()
	default:
		// 其它类型(如 struct)视为有值
		return true
	}
}
