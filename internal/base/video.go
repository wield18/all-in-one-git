package base

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

// FFProbeResult 对应 ffprobe -show_format -show_streams 的 JSON 输出
type FFProbeResult struct {
	Format struct {
		Duration string `json:"duration"` // 容器时长（秒，字符串）
		BitRate  string `json:"bit_rate"` // 容器总码率（bit/s）
	} `json:"format"`
}

func JustDurationSeconds(path string) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 使用 -v quiet 抑制横幅，-print_format json 结构化输出
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		path,
	)

	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w", err)
	}
	var result FFProbeResult
	if err := json.Unmarshal(out, &result); err != nil {
		return 0, fmt.Errorf("parse ffprobe json: %w", err)
	}

	// 时长：优先用 format.duration，有些容器格式为空
	if result.Format.Duration != "" {
		return strconv.ParseFloat(result.Format.Duration, 64)
	}
	return 0, errors.New("Duration is empty")
}
