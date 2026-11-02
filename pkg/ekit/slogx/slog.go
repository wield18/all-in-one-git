package slogx

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync/atomic"

	"gopkg.in/natefinch/lumberjack.v2"
)

func newErrEmptyParameter(key string) error {
	return fmt.Errorf("Lack of Parameter: %s", key)
}

type SlogConf struct {
	FileName      string `yaml:"FileName"`      // 日志文件路径
	MaxSize       int    `yaml:"MaxSize"`       // 单文件最大100MB
	MaxBackups    int    `yaml:"MaxBackups"`    // 保留30个备份文件
	MaxAge        int    `yaml:"MaxAge"`        // 保留30天
	StateFileName string `yaml:"StateFileName"` // State文件位置
}

// 这里返回的func是对于state.json的Save,需要主动存储
func InitSlog(conf SlogConf) func() {
	idHandler, SaveState, err := newIDHandlerV1(conf)
	if err != nil {
		panic(err)
	}
	logger := slog.New(idHandler)
	slog.SetDefault(logger)
	return SaveState
}

type IDManagerV1 struct {
	counter   uint64
	stateFile string
}

func newIDManagerV1(stateFile string) *IDManagerV1 {
	m := &IDManagerV1{
		stateFile: stateFile,
		counter:   0,
	}

	// 1. 尝试从状态文件恢复
	data, err := os.ReadFile(stateFile)
	if err == nil {
		var lastID uint64
		if json.Unmarshal(data, &lastID) == nil {
			m.counter = lastID
		}
	}

	return m
}

func (m *IDManagerV1) Next() uint64 {
	return atomic.AddUint64(&m.counter, 1)
}

func (m *IDManagerV1) SaveState() {
	data, _ := json.Marshal(m.counter)
	_ = os.WriteFile(m.stateFile, data, 0644)
}

// ============ 自定义Handler：注入ID ============
type IDHandlerV1 struct {
	inner slog.Handler
	idMgr *IDManagerV1
}

func newIDHandlerV1(conf SlogConf) (*IDHandlerV1, func(), error) {
	if conf.FileName == "" {
		return nil, nil, newErrEmptyParameter("FileName")
	}
	if conf.MaxSize == 0 {
		return nil, nil, newErrEmptyParameter("MaxSize")
	}
	if conf.MaxBackups == 0 {
		return nil, nil, newErrEmptyParameter("MaxBackups")
	}
	if conf.MaxAge == 0 {
		return nil, nil, newErrEmptyParameter("MaxAge")
	}
	if conf.StateFileName == "" {
		return nil, nil, newErrEmptyParameter("IdFileName")
	}
	// 2. 配置日志写入文件
	logFile := &lumberjack.Logger{
		Filename:   conf.FileName,
		MaxSize:    conf.MaxSize,
		MaxBackups: conf.MaxBackups,
		MaxAge:     conf.MaxAge,
		Compress:   true,
	}

	// 3. 创建JSON handler并注入ID
	baseHandler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelInfo, // info以上级别
		// AddSource: true,           // 添加代码上下文
	})

	imgr := newIDManagerV1(conf.StateFileName)

	return &IDHandlerV1{
		inner: baseHandler,
		idMgr: imgr,
	}, imgr.SaveState, nil
}

func (h *IDHandlerV1) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *IDHandlerV1) Handle(ctx context.Context, r slog.Record) error {
	// 注入自增ID
	id := h.idMgr.Next()
	r.AddAttrs(slog.Uint64("id", id))
	return h.inner.Handle(ctx, r)
}

func (h *IDHandlerV1) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &IDHandlerV1{
		inner: h.inner.WithAttrs(attrs),
		idMgr: h.idMgr,
	}
}

func (h *IDHandlerV1) WithGroup(name string) slog.Handler {
	return &IDHandlerV1{
		inner: h.inner.WithGroup(name),
		idMgr: h.idMgr,
	}
}
