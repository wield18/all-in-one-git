package mysqlx

import (
    "context"
    "errors"
    "log/slog"
    "time"

    "gorm.io/gorm/logger"
)

// SlogLogger 把 GORM 的日志接口转接到 slog
type SlogLogger struct {
    l *slog.Logger
}

func NewSlogLogger(l *slog.Logger) logger.Interface {
    return &SlogLogger{l: l}
}

// LogMode 返回自身，GORM 用它切换日志级别；这里统一 info，不做区分
func (s *SlogLogger) LogMode(logger.LogLevel) logger.Interface {
    return s
}

func (s *SlogLogger) Info(ctx context.Context, msg string, args ...any) {
    s.l.InfoContext(ctx, msg, args...)
}

func (s *SlogLogger) Warn(ctx context.Context, msg string, args ...any) {
    s.l.WarnContext(ctx, msg, args...)
}

func (s *SlogLogger) Error(ctx context.Context, msg string, args ...any) {
    s.l.ErrorContext(ctx, msg, args...)
}

// Trace 是 GORM 每次执行 SQL 后调用的核心方法
func (s *SlogLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
    elapsed := time.Since(begin)
    sql, rows := fc()

    // 如果需要忽略 record not found，可以在这里判断
    if err != nil && !errors.Is(err, logger.ErrRecordNotFound) {
        s.l.ErrorContext(ctx, "gorm sql error",
            "sql", sql,
            "rows", rows,
            "elapsed", elapsed,
            "error", err,
        )
        return
    }

    s.l.InfoContext(ctx, "gorm sql",
        "sql", sql,
        "rows", rows,
        "elapsed", elapsed,
    )
}