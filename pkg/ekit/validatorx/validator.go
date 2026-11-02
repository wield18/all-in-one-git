package validatex

import (
	"github.com/go-playground/validator/v10"
)

type CustomValidator struct {
	Tag      string
	Validate validator.Func
}

type CustomValidatorStruct struct {
	Structs  []interface{}
	Validate validator.StructLevelFunc
}

func NewValidator(validates []CustomValidator, validateStructs []CustomValidatorStruct) *validator.Validate {
	var validator = validator.New()
	applyValidates(validator, validates, validateStructs)
	return validator
}

func applyValidates(v *validator.Validate, validates []CustomValidator, validateStructs []CustomValidatorStruct) {
	var err error
	for _, one := range validates {
		if err = v.RegisterValidation(one.Tag, one.Validate); err != nil {
			panic(err)
		}
	}
	for _, one := range validateStructs {
		v.RegisterStructValidation(one.Validate, one.Structs...)
	}
}

func applyValidatesErr(v *validator.Validate, validates []CustomValidator) error {
	var err error
	for _, one := range validates {
		if err = v.RegisterValidation(one.Tag, one.Validate); err != nil {
			return err
		}
	}
	return nil
}

// func Apply[v any](abc *v, applys func(abc *v) error) {

// }
