package redisx

import (
	"context"
	"errors"
	"fmt"
	"time"
    "log/slog"

	"github.com/redis/go-redis/v9"
)

// loggingHook 实现 redis.Hook 接口
type loggingHook struct{}

func (h loggingHook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (h loggingHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmd)
		duration := time.Since(start)

		// 在这里打印日志，格式可以自己定
		slog.Info(fmt.Sprintf("[Redis] cmd=%s args=%v duration=%s err=%v \n\n",
			cmd.Name(), cmd.Args(), duration, err))
		return err
	}
}

func (h loggingHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmds)
		duration := time.Since(start)

		// 记录整个 pipeline 的信息
		slog.Info(fmt.Sprintf("[Redis Pipeline] cmds=%d duration=%s err=%v \n", len(cmds), duration, err))

		// 如果需要，也可以遍历 cmds 检查每条命令的错误
		for _, cmd := range cmds {
			args := cmd.Args()
			slog.Info(fmt.Sprint(args...))
			if cmdErr := cmd.Err(); cmdErr != nil && !errors.Is(cmdErr, redis.Nil) {
				slog.Info(fmt.Sprintf("[Redis Pipeline Cmd] name=%s err=%v \n", cmd.Name(), cmdErr))
			}
		}
		fmt.Println()
		return err
	}
}
