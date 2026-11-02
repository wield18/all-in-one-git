package videobuf

import (
	"encoding/json"
	"io"
	"strconv"
	"strings"
)

// searchVideo 按条件查询 video 索引，每个 UserID 只返回一条
func searchVideo(p searchVideoParams) []byte {
	// 用 map 构建 body，避免手写字符串导致转义/类型问题
	must := []any{}
	filter := []any{}

	if len(p.IncludedTags) != 0 {
		must = append(must, map[string]any{
			"terms": map[string]any{
				"Tags": p.IncludedTags,
			},
		})
	}
	if p.MinID > 0 {
		filter = append(filter, map[string]any{
			"range": map[string]any{
				"ID": map[string]any{
					"gt": p.MinID,
				},
			},
		})
	}
	if len(p.ExcludeUserIDs) != 0 {
		filter = append(filter, map[string]any{
			"bool": map[string]any{
				"must_not": []any{
					map[string]any{
						"terms": map[string]any{
							"UserID": p.ExcludeUserIDs,
						},
					},
				},
			},
		})
	}
	if p.MinViewsCount > 0 {
		filter = append(filter, map[string]any{
			"range": map[string]any{
				"ViewsCount": map[string]any{
					"gt": p.MinViewsCount,
				},
			},
		})
	}
	if p.Size < 1 {
		p.Size = 10
	}

	queryBool := map[string]any{}
	if len(must) != 0 {
		queryBool["must"] = must
	}
	if len(filter) != 0 {
		queryBool["filter"] = filter
	}
	body := map[string]any{
		"size": strconv.Itoa(p.Size),
		"query": map[string]any{
			"bool": queryBool,
		},
		"collapse": map[string]any{
			"field": "UserID",
		},
		"sort": []any{"ViewsCount"},
	}

	by, _ := json.Marshal(&body)
	return by
}

func ToIOReader(queries [][]byte) io.Reader {
	var sb strings.Builder
	for _, q := range queries {
		// 1. 写入 Header（因为已经用了 Msearch().Index("video")，这里可以传空对象 {}）
		sb.WriteString("{}\n")

		// 2. 写入 Query Body
		sb.Write(q)
		sb.WriteString("\n") // ⚠️ 每一行结束必须加换行符
	}
	abc := sb.String()
	return strings.NewReader(abc)
}

// searchVideoParams 查询参数
type searchVideoParams struct {
	IncludedTags   []string // Tags 至少匹配一个
	MinID          int      // ID > MinID
	ExcludeUserIDs []int64  // 排除这些 UserID
	MinViewsCount  int      // 最小播放次数
	Size           int
}

// 2. 定义单个 search 响应的结构
type SearchHit struct {
	Index  string          `json:"_index"`
	ID     string          `json:"_id"`
	Score  float64         `json:"_score"`
	Source json.RawMessage `json:"_source"` // 你的文档内容
}

type SearchHits struct {
	Total struct {
		Value int `json:"value"`
	} `json:"total"`
	Hits []SearchHit `json:"hits"`
}

type SingleSearchResponse struct {
	Hits SearchHits `json:"hits"`
	// 如果你还关心 took、timed_out 等，可以继续添加字段
}

// 3. 解析 msearch 的整体响应
func ParseMsearchResponse(reses interface{}) []SingleSearchResponse {
	// 将 any 序列化为 JSON 字节
	rawJSON, _ := json.Marshal(reses)

	// 定义顶层结构，用来接收整个 msearch 的响应体
	var msearchResp struct {
		Responses []json.RawMessage `json:"responses"`
	}

	json.Unmarshal(rawJSON, &msearchResp)

	// 遍历 responses 数组，逐个解析
	var results []SingleSearchResponse
	for _, raw := range msearchResp.Responses {
		var single SingleSearchResponse
		json.Unmarshal(raw, &single)
		results = append(results, single)
	}

	return results
}
