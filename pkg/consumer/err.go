package consumer

import (
	httperr "all-in-one/pkg/ekit/http_err"
	"fmt"
)

func NewTopicNotFound(topic string) error {
	return fmt.Errorf("topic: %s not found", topic)
}

var (
	errCtxDone = httperr.NewIError(500, 1000, "ctx done already")
	errNoData  = httperr.NewIError(500, 1000, "no data has been transmitted | 没有数据传过来")
)
