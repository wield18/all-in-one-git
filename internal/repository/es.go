package repository

import (
	"all-in-one/internal/base/constant"
	"time"
)

// return body,aFilter
func PagingES(body map[string]any, limitTime int64, page int, orderBy string, desc bool) (map[string]any, map[string]any) {
	body["from"] = page * constant.PageSize
	body["size"] = constant.PageSize
	if orderBy != "" {
		if desc {
			body["sort"] = []any{map[string]any{orderBy: map[string]any{"order": "desc"}}}
		} else {
			body["sort"] = []any{orderBy}
		}
	}
	aFilter := map[string]any{}
	if limitTime != 0 {
		aFilter = map[string]any{
			"range": map[string]any{CreatedAtPlain: time.Now().Add(-time.Duration(limitTime) * time.Second)},
		}
	}
	return body, aFilter
}
