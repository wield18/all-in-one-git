package base

import "regexp"

var emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func IsEmail(s string) bool {
	return emailRe.MatchString(s)
}

var e164Re = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

func IsE164(s string) bool {
	return e164Re.MatchString(s)
}
