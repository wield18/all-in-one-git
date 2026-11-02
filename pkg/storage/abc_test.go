package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestReadFileBindJson(t *testing.T) {
	state := "state.ini"
	// _, err := os.ReadFile(state)
	// if err != nil {
	// 	os.Create(state)
	// }
	data, _ := json.Marshal(map[string]int{
		"gaojian":    123,
		"gaojian123": 123123,
	})
	_ = os.WriteFile(state, data, 0644)

}

func TestMutex(*testing.T) {
	// aMap := map[string]int{}
	// aMap["asd"] = 1
	// dd := aMap["asd"]
	// abc :=
	// fmt.Println(dd)
	state := "state.ini"
	ab := NewAbc(state)
	ab.RegisterStorage("gaojian")
	num := ab.aMap["gaojian"]
	ab.aMap["gaojian"]++
	fmt.Println(num)
	ab.SaveState()
}

type abc struct {
	aMap  map[string]int
	mu    sync.Mutex
	path  string
	state int64
}

func NewAbc(path string) *abc {
	a := &abc{aMap: make(map[string]int), path: path}
	a.init()
	a.state = 1
	return a
}

func (a *abc) init() {
	data, err := os.ReadFile(a.path)
	if err != nil { // 没有文件重新来
		return
	}
	json.Unmarshal([]byte(data), &a.aMap)
}

func (a *abc) GetIndex(name string) (int, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.state == 0 {
		return 0, false
	}
	// a.aMap[]
	return 0, false
}

func (a *abc) RegisterStorage(name string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.aMap[name]; ok { // 如果找得到
		return
	}
	a.aMap[name] = 0 // 没找到直接重计时
}

func (a *abc) SaveState() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.state = 0
	by, _ := json.Marshal(a.aMap)
	os.WriteFile(a.path, by, 0644)
}

// 我首先得有一堆的可能的无关的存储,有的可能开了有的没开

// func TestSaveBytes(t *testing.T) {
// 	abc := map[string][]byte{}
// 	by, _ := json.Marshal(FileStorageData{TmpPath: "wdw", StorePath: "dwd"})
// 	abc["gaojian"] = by
// 	by, _ = json.Marshal(abc)
// 	os.WriteFile("state.ini", by, 0644)
// }

// func TestReadBytes(t *testing.T) {
// 	by, _ := os.ReadFile("state.ini")

// 	abc := map[string][]byte{}
// 	json.Unmarshal(by, &abc)
// 	data := FileStorageData{}
// 	json.Unmarshal(abc["gaojian"], &data)
// 	fmt.Println(data)
// 	// by, _ := json.Marshal(FileStoragePath{TmpPath: "wdw", StorePath: "dwd"})
// 	// abc["gaojian"] = by
// 	// by, _ = json.Marshal(abc)
// 	// os.WriteFile("state.ini", by, 0644)
// }

func TestChan(t *testing.T) {
	ch := make(chan struct{})
	close(ch)
	close(ch)
}

func TestFileStorage(t *testing.T) {
	videoStorage := NewFileStorage(`/path/to/video`)
	by, err := os.ReadFile(`D:\Animation\Minus8_Archive 一个性的时代的代表\Minus8 Archive\ANIMATIONS\!MP4s\EehFwLdU4AAthXv.mp4`)
	sum := sha256.Sum256(by)
	sha256Hex := hex.EncodeToString(sum[:]) // 64 字符 hex

	if err != nil {
		fmt.Println(err)
		return
	}
	// videoStorage.Init()
	path := videoStorage.GetFilePath()
	fmt.Println(path)

	// 服务端重新算 SHA256
	sum1 := sha256.Sum256(by)
	serverHash1 := hex.EncodeToString(sum1[:])

	if serverHash1 != sha256Hex {
		fmt.Println("no way")
		return
	}
	// 因为一个文件一个dir,根本不可能重名
	err = os.WriteFile(filepath.Join(path, "0.mp4"), by, 0755)
	if err != nil {
		fmt.Println(err)
		return
	}

}

// 正常从file跟map来进行初始化
// 先开服务再进行register查看状态
// 使用saveState添加式重写

func TestStorageManager(t *testing.T) {
	f := NewFileStorage(`/path/to/video\5\video`)
	s := NewStorageManagerFromFileAndMap(`/path/to/video\5\state.json`, map[string]StorageInter{
		"video": f,
	})
	p := NewFileStorage(`/path/to/video\5\pic`)
	s.Register("pic", p)
	fmt.Println(f.GetFilePath())
	fmt.Println(f.GetFilePath())
	fmt.Println(f.GetFilePath())
	p.GetFilePath()
	p.GetFilePath()
	p.GetFilePath()
	s.SaveState()
}

func TestMap(t *testing.T) {
	aMap := make(map[string]interface{})
	abc := aMap["dwd"]
	fmt.Println(abc)
}
