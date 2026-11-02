package base

import (
	"all-in-one/internal/base/constant"
	httperr "all-in-one/pkg/ekit/http_err"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
)

func Sprintf(keyStart string, segs ...any) string {
	if keyStart == "" {
		return ""
	}
	for _, v := range segs {
		switch v := v.(type) {
		case uint, uint8, uint16, uint32, uint64,
			int, int8, int16, int32, int64:
			keyStart += fmt.Sprintf(":%v", v)
		case string:
			keyStart += ":" + v
		case []byte:
			keyStart += ":" + string(v)
		default:
			// 不支持的类型，可以选择跳过、panic 或返回错误
			continue
		}
	}
	return keyStart
}

// 升序排序生成string
func SprintfSortStrings(strs ...string) string {
	l := len(strs)
	if l == 0 {
		return ""
	}
	sort.Strings(strs)
	for i := 1; i < len(strs); i++ {
		strs[0] += strs[i]
	}
	return strs[0]
}

// 必须命中所有传入的key
// 所有字段必须非空存在
func CheckMapKeys(aMap map[string]string, keys ...string) error {
	for _, v := range keys {
		val, ok := aMap[v]
		// 非零而且找到
		if !ok || val == "" {
			return fmt.Errorf("key: %s's not found", v)
		}

	}
	return nil
}

// 简单json跟empty一下
// 只能全是string
func JsonMarshalAndCheckEmpty(rawCode []byte, keys ...string) (map[string]string, error) {
	aMap := map[string]string{}
	err := json.Unmarshal(rawCode, &aMap)
	if err != nil {
		return map[string]string{}, fmt.Errorf("unable json unmarshal: %w", err)
	}
	// 检查一下, 如果没有命中所有需要的keys那就是非法的
	return aMap, CheckMapKeys(aMap, keys...)
}

func ResultGin(c *gin.Context, data interface{}) {
	switch data := data.(type) {
	case string, map[string]interface{}:
		c.Set(constant.Result, data)
	default:
		panic(fmt.Sprintf("type is not string or map, is: %#v", data))
	}
}

func IsToday(t time.Time) bool {
	now := time.Now()
	y1, m1, d1 := t.Date()
	y2, m2, d2 := now.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// 添加,所以说默认就得拿到 ContextIDAndPath
func AddResultIErrAndArgsByAppendArgs(c *gin.Context, ierr httperr.IError, args SlogArgs) {
	ierrs, ok := c.Get(constant.IErrAndArgs)
	appendedArgs := NewSlogArgs(6).SlogContextIDAndPath(c)
	appendedArgs = append(appendedArgs, args)
	if !ok {
		abc := make([]*IErrAndArgs, 1, 3)
		abc[0] = &IErrAndArgs{IErr: ierr, Args: appendedArgs}
		c.Set(constant.IErrAndArgs, abc)
		return
	}
	ierrAndArgs := ierrs.([]*IErrAndArgs)
	ierrAndArgs = append(ierrAndArgs, &IErrAndArgs{IErr: ierr, Args: appendedArgs})
	c.Set(constant.IErrAndArgs, ierrAndArgs) // 重新set
}

// 添加
func AddResultIErrAndArgs(c *gin.Context, ierr httperr.IError) {
	ierrs, ok := c.Get(constant.IErrAndArgs)
	if !ok {
		abc := make([]*IErrAndArgs, 1, 3)
		abc[0] = &IErrAndArgs{IErr: ierr, Args: NewSlogArgs(4).SlogContextIDAndPath(c)}
		c.Set(constant.IErrAndArgs, abc)
		return
	}
	ierrAndArgs := ierrs.([]*IErrAndArgs)
	ierrAndArgs = append(ierrAndArgs, &IErrAndArgs{IErr: ierr, Args: NewSlogArgs(4).SlogContextIDAndPath(c)})
	c.Set(constant.IErrAndArgs, ierrAndArgs) // 重新set
}

func AddResultIErrAndArgsPure(c *gin.Context, ierrArgs *IErrAndArgs) {
	if ierrArgs == nil {
		return
	}
	ierrs, ok := c.Get(constant.IErrAndArgs)
	if !ok {
		abc := make([]*IErrAndArgs, 1, 3)
		abc[0] = ierrArgs
		c.Set(constant.IErrAndArgs, abc)
		return
	}
	ierrAndArgs := ierrs.([]*IErrAndArgs)
	ierrAndArgs = append(ierrAndArgs, ierrArgs)
	c.Set(constant.IErrAndArgs, ierrAndArgs) // 重新set
}

// 添加
func AddResultIErrAndArgsWithSourceArgs(c *gin.Context, ierr httperr.IError) {
	ierrs, ok := c.Get(constant.IErrAndArgs)
	if !ok {
		abc := make([]*IErrAndArgs, 1, 3)
		abc[0] = &IErrAndArgs{IErr: ierr, Args: NewSlogArgs(6).SlogContextIDAndPath(c).SlogSource()}
		c.Set(constant.IErrAndArgs, abc)
		return
	}
	ierrAndArgs := ierrs.([]*IErrAndArgs)
	ierrAndArgs = append(ierrAndArgs, &IErrAndArgs{IErr: ierr, Args: NewSlogArgs(6).SlogContextIDAndPath(c).SlogSource()})
	c.Set(constant.IErrAndArgs, ierrAndArgs) // 重新set
}

// 添加
func AddResultIErrAndArgsByBatch(c *gin.Context, ierrBatch ...*IErrAndArgs) {
	ierrs, ok := c.Get(constant.IErrAndArgs)
	if !ok {
		abc := make([]*IErrAndArgs, 1, 3)
		abc = append(abc, ierrBatch...)
		c.Set(constant.IErrAndArgs, abc)
		return
	}
	ierrAndArgs := ierrs.([]*IErrAndArgs)
	ierrAndArgs = append(ierrAndArgs, ierrBatch...)
	c.Set(constant.IErrAndArgs, ierrAndArgs) // 重新set
}

// 添加
func AddResultIErrAndArgsByBatchDefaultArgs(c *gin.Context, ierrBatch []httperr.IError) {
	ierrs, ok := c.Get(constant.IErrAndArgs)
	if !ok {
		ierrAndArgs := make([]*IErrAndArgs, 1, 3)
		for _, ierr := range ierrBatch {
			ierrAndArgs = append(ierrAndArgs, &IErrAndArgs{IErr: ierr, Args: NewSlogArgs().SlogContextIDAndPath(c)})
		}
		c.Set(constant.IErrAndArgs, ierrAndArgs)
		return
	}
	ierrAndArgs := ierrs.([]*IErrAndArgs)
	for _, ierr := range ierrBatch {
		ierrAndArgs = append(ierrAndArgs, &IErrAndArgs{IErr: ierr, Args: NewSlogArgs().SlogContextIDAndPath(c)})
	}
	c.Set(constant.IErrAndArgs, ierrAndArgs) // 重新set
}
