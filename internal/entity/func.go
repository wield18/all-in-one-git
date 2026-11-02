package entity

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
)

func ToUrlSafeBase64String(c interface{}) string {
	by, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(by)
}

func ToSliceInt64(str string) ([]int64, error) {
	trimmed := strings.Trim(str, "[]")
	parts := strings.Split(trimmed, ",")
	aSlice := make([]int64, 0, len(parts))
	var err error
	for _, p := range parts {
		var aInt int64
		aInt, err = strconv.ParseInt(p, 10, 64)
		if err != nil {
			return nil, err
		}
		aSlice = append(aSlice, aInt)
	}
	return aSlice, err
}

func ToSliceString(str string) ([]string, error) {
	trimmed := strings.Trim(str, "[]")
	parts := strings.Split(trimmed, ",")
	aSlice := make([]string, 0, len(parts))
	var err error
	for _, p := range parts {
		aSlice = append(aSlice, p)
	}
	return aSlice, err
}
