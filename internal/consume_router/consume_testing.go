package consumerouter

import (
	"all-in-one/internal/entity"
	"all-in-one/pkg/consumer"
	"context"
	"fmt"
)

func ConsumeKeyValue(ctx context.Context, datas []*consumer.ConsumeData[entity.KeyValue]) {
	for _, data := range datas {
		fmt.Println(data.Data.Key)
		fmt.Println(data.Data.Value)
	}
}
