package videobuf

import (
	"fmt"
	"testing"
)

func TestMap(t *testing.T) {
	aMap := map[string]struct{}{}
	operateMap(aMap)
	fmt.Println(aMap)
}

func operateMap(aMap map[string]struct{}) {
	aMap["ad"] = struct{}{}
	for i := range aMap {
		delete(aMap, i)
	}
	aMap["adasd"] = struct{}{}
}

// func TestGetMap(t *testing.T) {
// 	aMaps := New(redisx.NewRedis(config.Conf.Redis), es.NewESClient(config.Conf.ES))

// 	res, err := aMaps.GetRes(1, &entity.VideoBytes{
// 		IncludedTags: []entity.VideoBytesTag{
// 			{Tag: "zxc", Ratio: 0.5},
// 			{Tag: "fgh", Ratio: 0.3},
// 			{Tag: "uio", Ratio: 0.19},
// 			{Tag: "asd", Ratio: 0.01},
// 		},
// 		ExcludedUserIDs: []int64{1, 2, 3, 11, 12, 13, 21, 22, 23},
// 	})
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	fmt.Println(res)
// 	res, err = aMaps.GetRes(1, &entity.VideoBytes{
// 		IncludedTags: []entity.VideoBytesTag{
// 			{Tag: "zxc", Ratio: 0.5},
// 			{Tag: "fgh", Ratio: 0.3},
// 			{Tag: "uio", Ratio: 0.19},
// 			{Tag: "asd", Ratio: 0.01},
// 		},
// 		ExcludedUserIDs: []int64{1, 2, 3, 11, 12, 13, 21, 22, 23},
// 	})
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	fmt.Println(res)

// 	// res, err := aMaps.GetESRes([]string{"asd", "qwe"}, aMaps.GetVideoBuf(123), []int64{})
// 	// if err != nil {
// 	// 	fmt.Println(err)
// 	// 	return
// 	// }
// 	// fmt.Println(res)
// 	// fmt.Println(len(res))

// 	// res, err := aMaps.GetRedisRes([]string{"asd", "qwe"})
// 	// if err != nil {
// 	// 	fmt.Println(err)
// 	// 	return
// 	// }
// 	// fmt.Println(res)
// }

// func TestVideoBufMap(t *testing.T) {
// 	abc := VideoBufMap{aMap: make(map[string]struct{}), region: make([][]int, 0)}
// 	abc.AddRegions([]int{0, 100}, []int{1, 50}, []int{40, 200}, []int{123123, 13123213})
// 	fmt.Println(abc.region)
// }

func TestAbcd(t *testing.T) {
	aS := [][2]int{}
	aS = append(aS, [2]int{1})
	ac := aS[0]
	ac[0] = 123
	fmt.Println(aS[0][0])
}
