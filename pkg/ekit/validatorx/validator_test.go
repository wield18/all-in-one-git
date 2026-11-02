package validatex

// func initValidatorErr() {
// 	validates := []CustomValidator{
// 		{
// 			Tag:      noPunctTag,
// 			Validate: noPunct,
// 		},
// 	}
// 	err := applyValidatesErr(ValidatorEntity, validates)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// }

// func TestValidateUserV1Update(t *testing.T) {
// 	initValidator()
// 	testCases := []struct {
// 		name   string
// 		data   UserV1UpdateOne
// 		hasErr bool
// 	}{
// 		{
// 			name: "1of5 2 failed",
// 			data: UserV1UpdateOne{
// 				Name:        "gaojian",
// 				Password:    "dwdw",
// 				OldPassword: "dwd",
// 			},
// 			hasErr: true,
// 		},
// 		{
// 			name:   "1of5 0",
// 			data:   UserV1UpdateOne{},
// 			hasErr: true,
// 		},
// 		{
// 			name: "1of5 succeeed",
// 			data: UserV1UpdateOne{Name: "wdwdw"},
// 		},
// 		{
// 			name:   "password failed",
// 			data:   UserV1UpdateOne{Password: "dwdwd"},
// 			hasErr: true,
// 		},
// 		{
// 			name:   "password succeed",
// 			data:   UserV1UpdateOne{Password: "dwdw", OldPassword: "dwdw", T: "b64s", Captcha: "dww", CaptchaToken: "dwdw"},
// 			hasErr: false,
// 		},
// 		{
// 			name: "phone succeed",
// 			data: UserV1UpdateOne{Phone: "+8612345678912", OldPhone: "+8612345678912", T: Phone, Captcha: "dw", CaptchaToken: "dwdw"},
// 		},
// 		{
// 			name:   "Status failed",
// 			data:   UserV1UpdateOne{Status: "anxiousss"},
// 			hasErr: true,
// 		},
// 		{
// 			name: "status succeed",
// 			data: UserV1UpdateOne{Status: "anxious"},
// 		},
// 		{
// 			name:   "password captcha nil failed",
// 			data:   UserV1UpdateOne{Password: "dwdw", OldPassword: "dwdw"},
// 			hasErr: true,
// 		},
// 		{
// 			name:   "password captcha not match failed",
// 			data:   UserV1UpdateOne{Password: "dwdw", OldPassword: "dwdw", T: Email, Captcha: "dw", CaptchaToken: "dw"},
// 			hasErr: true,
// 		},
// 	}
// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			fmt.Println()
// 			fmt.Println(tc.name)
// 			tc.data.ID = 1233
// 			err := ValidatorEntity.Struct(&tc.data)
// 			if err == nil {
// 				err = tc.data.Validate()
// 			}
// 			if (err != nil) == tc.hasErr { // 有err时,如果同时真的有err才命中
// 				fmt.Println("命中")
// 				fmt.Printf("err: %v\n", err)
// 			} else {
// 				fmt.Println("未命中")
// 				fmt.Printf("name: %s\n", tc.name)
// 				if err != nil {
// 					fmt.Printf("err: %v\n", err)
// 				} else {
// 					fmt.Println("no err")
// 				}
// 			}
// 		})
// 	}
// }

// // func TestValidateUserV1Login(t *testing.T) {
// // 	initValidator()
// // 	testCases := []struct {
// // 		name   string
// // 		data   UserV1Login
// // 		hasErr bool
// // 	}{
// // 		{
// // 			name: "name login",
// // 			data: UserV1Login{
// // 				Name:              "gaojian",
// // 				PasswordPlainTest: "wddw",
// // 			},
// // 			hasErr: false,
// // 		},
// // 		{
// // 			name: "email login",
// // 			data: UserV1Login{
// // 				Email:             "123123@gmail.com",
// // 				PasswordPlainTest: "dwdw",
// // 			},
// // 			hasErr: false,
// // 		},
// // 		{
// // 			name: "email invalid",
// // 			data: UserV1Login{
// // 				Email:             "dwdw",
// // 				PasswordPlainTest: "dw",
// // 			},
// // 			hasErr: true,
// // 		},
// // 		{
// // 			name: "email and phone",
// // 			data: UserV1Login{
// // 				Email:             "123@gmail.com",
// // 				Phone:             "+8612345678912",
// // 				PasswordPlainTest: "dwd",
// // 			},
// // 			hasErr: false,
// // 		},
// // 	}
// // 	for _, tc := range testCases {
// // 		t.Run(tc.name, func(t *testing.T) {
// // 			fmt.Println()
// // 			tc.data.UserAgent = "fw"
// // 			tc.data.Platform = "dwd"
// // 			tc.data.Timezone = "dwdwe"
// // 			err := ValidatorEntity.Struct(&tc.data)
// // 			if (err != nil) == tc.hasErr { // 有err时,如果同时真的有err才命中
// // 				fmt.Println("命中")
// // 				fmt.Printf("err: %v\n", err)
// // 			} else {
// // 				fmt.Println("未命中")
// // 				fmt.Printf("name: %s\n", tc.name)
// // 				if err != nil {
// // 					fmt.Printf("err: %v\n", err)
// // 				} else {
// // 					fmt.Println("no err")
// // 				}
// // 			}
// // 		})
// // 	}
// // }

// func TestMAb(t *testing.T) {
// 	type aStruct struct {
// 		// 当email没有时才必须要求phone有,俩都有也行
// 		// 这个不是多选一,这是多项至少有一个
// 		Phone string `form:"phone" validate:"required_without=Email,omitempty,e164"`
// 		Email string `form:"email" validate:"required_without=Phone,omitempty,email"`
// 	}
// 	a := aStruct{Phone: "+8612345678912"}
// 	if err := ValidatorEntity.Struct(&a); err != nil {
// 		fmt.Println("good")
// 		fmt.Println(err)
// 	}
// }

// type contact struct {
// 	Email  string `exactlyone:"contact" validate:"email"`
// 	Phone  string `exactlyone:"contact" validate:"e164"`
// 	WeChat string `exactlyone:"contact"`
// }

// func TestCustom(*testing.T) {
// 	validate := validator.New()
// 	validate.RegisterStructValidation(validateExactlyOne, contact{})

// 	// 通过: 恰好一个
// 	fmt.Println(validate.Struct(&contact{Email: "a@b.com"}))
// 	fmt.Println()

// 	// 失败: 两个都有值
// 	fmt.Println(validate.Struct(&contact{Email: "a@b.com", Phone: "123"}))
// 	fmt.Println()

// 	// 失败: 一个都没有
// 	fmt.Println(validate.Struct(&contact{Email: "231"}))
// 	fmt.Println()
// }

// // func TestValidateUserV1Register(t *testing.T) {
// // 	initValidatorErr()
// // 	testCases := []struct {
// // 		name   string
// // 		data   UserV1Register
// // 		hasErr bool
// // 	}{
// // 		{
// // 			name: "name",
// // 			data: UserV1Register{
// // 				Name:              "g..",
// // 				PasswordPlainTest: "123",
// // 				Email:             "qwe@qq.com",
// // 				Phone:             "+861564846464",
// // 				Captcha:           "123456",
// // 				CaptchaToken:      "1234567890123456",
// // 				UserAgent:         "123",
// // 				Platform:          "456",
// // 				Timezone:          "Asia/Shanghai",
// // 				T:                 "email",
// // 			},
// // 			hasErr: true,
// // 		},
// // 		{
// // 			name: "email and phone",
// // 			data: UserV1Register{
// // 				Name:              "gao",
// // 				PasswordPlainTest: "dw",
// // 				Captcha:           "123456",
// // 				CaptchaToken:      "1234567890123456",
// // 				UserAgent:         "123",
// // 				Platform:          "456",
// // 				Timezone:          "Asia/Shanghai",
// // 				T:                 "email",
// // 			},
// // 			hasErr: true,
// // 		},
// // 		{
// // 			name: "has email",
// // 			data: UserV1Register{
// // 				Name:              "gao",
// // 				PasswordPlainTest: "dw",
// // 				Email:             "qwe@qq.com",
// // 				Captcha:           "123456",
// // 				CaptchaToken:      "1234567890123456",
// // 				UserAgent:         "123",
// // 				Platform:          "456",
// // 				Timezone:          "Asia/Shanghai",
// // 				T:                 "email",
// // 			},
// // 			hasErr: false,
// // 		},
// // 		{
// // 			name: "captcha len",
// // 			data: UserV1Register{
// // 				Name:              "gao",
// // 				PasswordPlainTest: "dw",
// // 				Email:             "qwe@qq.com",
// // 				Captcha:           "1234567",
// // 				CaptchaToken:      "1234567890123456",
// // 				UserAgent:         "123",
// // 				Platform:          "456",
// // 				Timezone:          "Asia/Shanghai",
// // 				T:                 "email",
// // 			},
// // 			hasErr: true,
// // 		},
// // 	}
// // 	for _, v := range testCases {
// // 		t.Run(v.name, func(t *testing.T) {
// // 			err := ValidatorEntity.Struct(v.data)
// // 			if err != nil {
// // 				fmt.Println(err)
// // 			}
// // 			if (err != nil) != v.hasErr {
// // 				fmt.Println("未命中", v.name)
// // 			} else {
// // 				fmt.Println("过", v.name)

// // 			}
// // 			// if (err == nil) == v.hasErr {
// // 			// }
// // 		})
// // 	}

// // }

// func TestRequestToken(t *testing.T) {
// 	err := ValidatorEntity.Struct(RequestToken{
// 		ID: "dwdw",
// 	})
// 	if err != nil {
// 		fmt.Println(1)
// 		fmt.Println(err)
// 		return
// 	}
// 	err = ValidatorEntity.Struct(RequestToken{
// 		ID:     "dwdfer",
// 		UserID: 1,
// 	})
// 	if err != nil {
// 		fmt.Println(2)
// 		fmt.Println(err)
// 		err = ValidatorEntity.Struct(RequestToken{
// 			ID:        "dwdfer",
// 			UserID:    2,
// 			UserToken: "dwdw",
// 		})
// 		if err != nil {
// 			fmt.Println(3)
// 			fmt.Println(err)
// 		}
// 	}

// }
