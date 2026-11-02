package videobuf

import (
	"all-in-one/internal/entity"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math/rand/v2"
	"sort"
	"sync"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/redis/go-redis/v9"
)

type VideoBufServer struct {
	mapBufs map[int64]*VideoBufMap
	rmu     sync.RWMutex
	// rdb     *redis.Client              // 供随机热门提供
	// es      *elasticsearch.TypedClient // 基本的tag跟userid搜索
	resSize int // 返回的大小

	// 内部 VideoBufMap 设置
	maxSize      int // map最大大小
	minRegionGap int // 区间合并的最小间隔

	// es 部分
	es         *elasticsearch.TypedClient
	esForCount int // 最多可以几个tags
	esStart    int // 从哪个id开始
	esStep     int // 一步多少
	esCount    int // 查几次
	esSize     int // 返回大小
	// redis 部分
	rdb            *redis.Client
	redisForCount  int      // 随机循环的最大次数
	prefix         string   // 前缀
	choices        []string // 随机的键名从哪里来
	resRedisLen    int      // 要拿多少个redis数据
	randRedisRange int      // 真正可以随机的范围
}

var ()

type VideoBufMap struct {
	aMap   map[string]json.RawMessage
	region [][]int
}

func New(
	rdb *redis.Client, es *elasticsearch.TypedClient,
	options ...vBufOption) *VideoBufServer {
	return &VideoBufServer{
		mapBufs: make(map[int64]*VideoBufMap),
		// rdb:     rdb,
		// es:      es,
		resSize: 20,

		maxSize:      100,
		minRegionGap: 10,

		rdb:            rdb,
		redisForCount:  3,
		prefix:         "video:",
		choices:        []string{"zxc", "asd", "qwe", "vbn", "fgh", "rty", "jkl", "uio"},
		randRedisRange: 91,
		resRedisLen:    9,

		es:         es,
		esForCount: 5,
		esStart:    1000,
		esStep:     500,
		esSize:     20,
		esCount:    2,
	}

}

// 大众池子
func (v *VideoBufServer) GeneralRes() ([]json.RawMessage, error) {
	aMap := v.GetMap(0)
	if len(aMap) == 0 {
		// 初始
		for _, choice := range v.choices {
			res, err := v.rdb.ZRangeArgs(context.TODO(), redis.ZRangeArgs{
				Key:   v.prefix + choice,
				Start: 0,
				Stop:  9,
			}).Result()
			if err != nil {
				return nil, err
			}

			for _, one := range res {
				js := json.RawMessage(one)
				aMap[v.hashRawMsg(js)] = js
			}
		}
	}
	res := []json.RawMessage{}
	var i int
	for _, abc := range aMap {
		if i == (v.resSize) {
			break
		}
		res = append(res, abc)
		i++
	}
	return res, nil
}

// 简单哈希键
func (v *VideoBufServer) hashRawMsg(raw json.RawMessage) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]) // 64 字符

}

func (v *VideoBufServer) GetRes(userID int64, bytes *entity.VideoBytes) ([]json.RawMessage, error) {
	// 这里做下 includedTags 为空的情况
	if len(bytes.IncludedTags) == 0 {
		// 跳大众池子
		return v.GeneralRes()
	}
	aBuf := v.GetVideoBuf(userID)
	// 看看当前的大小够不够返回的
	if len(aBuf.aMap) > v.resSize {
		res := []json.RawMessage{}
		var i int
		for key, abc := range aBuf.aMap {
			if i == (v.resSize) {
				break
			}
			res = append(res, abc)
			delete(aBuf.aMap, key)
			i++
		}
		return res, nil
	}

	includedTags := v.GetIncludedTags(bytes.IncludedTags)

	resRdb, err := v.GetRedisRes(includedTags)
	if err != nil {
		return nil, err
	}
	resES, err := v.GetESRes(includedTags, aBuf, bytes.ExcludedUserIDs)
	if err != nil {
		return nil, err
	}
	l1, l2 := len(resRdb), len(resES)
	for i := 0; i < l1; i++ {
		aBuf.aMap[v.hashRawMsg(resRdb[i])] = resRdb[i]
	}
	for i := 0; i < l2; i++ {
		aBuf.aMap[v.hashRawMsg(resES[i])] = resES[i]
	}
	res := []json.RawMessage{}
	var i int
	for key, abc := range aBuf.aMap {
		if i == (v.resSize) {
			break
		}
		res = append(res, abc)
		delete(aBuf.aMap, key)
		i++
	}
	return res, nil
}

// 新增map部分
// 这里我不认为可能有人多机登录并同时获得map
func (v *VideoBufServer) GetMap(UserID int64) map[string]json.RawMessage {
	v.rmu.RLock()
	if OneBuf, ok := v.mapBufs[UserID]; ok {
		v.rmu.RUnlock()
		return OneBuf.aMap
	}
	v.rmu.RUnlock()

	v.rmu.Lock()
	defer v.rmu.Unlock()
	if aMap, ok := v.mapBufs[UserID]; ok {
		return aMap.aMap
	}
	aMap := make(map[string]json.RawMessage, v.maxSize)
	v.mapBufs[UserID] = &VideoBufMap{aMap: aMap, region: make([][]int, 0, 10)}
	return aMap
}

func (v *VideoBufServer) GetVideoBuf(UserID int64) *VideoBufMap {
	v.rmu.RLock()
	if OneBuf, ok := v.mapBufs[UserID]; ok {
		v.rmu.RUnlock()
		return OneBuf
	}
	v.rmu.RUnlock()

	v.rmu.Lock()
	defer v.rmu.Unlock()
	if aMap, ok := v.mapBufs[UserID]; ok {
		return aMap
	}
	buf := &VideoBufMap{aMap: make(map[string]json.RawMessage, v.maxSize), region: make([][]int, 0, 10)}
	v.mapBufs[UserID] = buf
	return buf
}

func (v *VideoBufServer) GetRedisRes(alreadyIncludedTags []string) ([]json.RawMessage, error) {
	// 随机id来查redis
	randChoice := ""
	for i := 0; i < v.redisForCount; i++ {
		randChoice = v.choices[rand.IntN(len(v.choices))]
		has := false
		for _, v := range alreadyIncludedTags {
			if randChoice == v {
				has = true
				break
			}
		}
		if has {
			continue
		}
		break
	}

	randNewId := rand.IntN(v.randRedisRange)
	res, err := v.rdb.ZRangeArgs(context.TODO(), redis.ZRangeArgs{
		Key:   v.prefix + randChoice,
		Start: randNewId,
		Stop:  randNewId + v.resRedisLen,
	}).Result()
	if err != nil {
		return nil, err
	}
	bytes := []json.RawMessage{}
	for _, v := range res {
		bytes = append(bytes, json.RawMessage(v))
	}
	return bytes, err

}

// 内部进行 region 的更新
func (v *VideoBufServer) GetESRes(includedTags []string, aBuf *VideoBufMap, excludedUserIDs []int64) ([]json.RawMessage, error) {
	id := v.esStart + rand.IntN(v.esStep)
	bytes := [][]byte{}
	for i := 0; i < v.esCount; i++ {
		bytes = append(bytes, searchVideo(searchVideoParams{IncludedTags: includedTags, ExcludeUserIDs: excludedUserIDs, MinID: id, Size: v.esSize}))
		id += v.esStep
	}

	// 随机id来查es
	model := entity.VideoES{}
	resES, err := v.es.Msearch().Index(model.IndexName()).Raw(ToIOReader(bytes)).Do(context.TODO())
	data := ParseMsearchResponse(resES)
	dat := []json.RawMessage{}
	for i := 0; i < len(data); i++ {
		// 这里可以直接获得区间问题
		if len(data[i].Hits.Hits) > 1 {
			dst, src := 0, 0
			dataStruct := entity.VideoES{}
			err := json.Unmarshal(data[i].Hits.Hits[0].Source, &dataStruct)
			if err != nil {
				return nil, err
			}
			dst = int(dataStruct.ID)

			err = json.Unmarshal(data[i].Hits.Hits[len(data[i].Hits.Hits)-1].Source, &dataStruct)
			if err != nil {
				return nil, err
			}
			src = int(dataStruct.ID)
			// 添加区间
			aBuf.region = append(aBuf.region, []int{dst, src})
		}
		for _, v1 := range data[i].Hits.Hits {
			// by, _ := json.Marshal(&v1.Source)
			dat = append(dat, v1.Source)
		}
	}
	v.sortRegions(aBuf)
	return dat, err
}

// 不到二倍直接断掉
func (v *VideoBufServer) GetIncludedTags(tags []entity.VideoBytesTag) []string {
	// 所有需要的tags
	if len(tags) == 0 {
		return []string{}
	}
	curRatio := tags[0].Ratio
	includedTags := []string{tags[0].Tag}
	count := 0
	for i := 1; i < len(tags) && count < v.esForCount; i++ {
		if curRatio/2 > tags[i].Ratio {
			break
		}
		curRatio = tags[i].Ratio
		includedTags = append(includedTags, tags[i].Tag)
		// 循环次数
		count++
	}
	return includedTags
}

// 必须
// len(aRegion) == 2
// []int dst < src
// 可传可不传 regions 可以直接操作VideoBufMap传进去再直接 AddRegions
func (vS *VideoBufServer) sortRegions(v *VideoBufMap, regions ...[]int) {
	// 这里直接append并sort排序下
	if len(regions) != 0 {
		v.region = append(v.region, regions...)
	}
	sort.Slice(v.region, func(i, j int) bool { return v.region[i][0] < v.region[j][0] })

	// 这里找需要保存的索引
	// 更改对应的数值
	previousOne := v.region[0]
	vaildedIds := []int{0}
	for i := 1; i < len(v.region); i++ {
		if previousOne[1]+vS.minRegionGap >= v.region[i][0] {
			if previousOne[1] < v.region[i][1] { // 大于等于下个的左边界但小于下个的右边界 --> 更新
				previousOne[1] = v.region[i][1]
			}
			continue
		}
		// 如果不大于左边界加gap --> 更新previousOne,vaildedIds
		previousOne = v.region[i]
		vaildedIds = append(vaildedIds, i)
	}

	aSlice := make([][]int, len(vaildedIds))
	for i, regionI := range vaildedIds {
		aSlice[i] = v.region[regionI]
	}
	v.region = aSlice
}
