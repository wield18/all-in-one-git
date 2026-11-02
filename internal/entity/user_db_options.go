package entity

// --------------UserV1

// 检查需要
func (*UserV1) CheckPasswordSelect() []string {
	return []string{"ID", "PasswordCrypto"}
}

// ------------------LoginLocation

func (v *LoginLocation) ToMapForTransmitMany() map[string]interface{} {
	return map[string]interface{}{
		LocationPlain:  v.Location,
		LoginTimePlain: v.LoginTime,
	}
}
