package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

// 首先自增一个id
// 然后根据所给的name创个文件夹呗
// 还有对应的state save

// 估计这里做h256的问题

// storage是否已经关闭得看返回的bool

type StorageInter interface {
	// Key() string           // 获得对应的key好返回各自的map
	Init(data interface{}) // 绑定自己的struct
	SaveState() (state interface{})
}

const (
	IndexPlain = "Index"
)

// 这个东西的tmp或store的path是根据主文件来的，而具体的在于用name来分一下
type FileStorage struct {
	path  string
	mu    sync.Mutex
	index int
}

func NewFileStorage(dir string) *FileStorage {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
	s := &FileStorage{path: dir}

	// 这里for一下,防意外save了一下但是还有人在写文件,导致后来启动的直接覆盖掉
	for err == nil {
		s.index++
		_, err = os.Stat(filepath.Join(dir, strconv.Itoa(s.index)))
	}
	return s
}

// // key
// func (s *FileStorage) Key() string {
// 	return filepath.Dir(s.path)
// }

// 无疑也就是写文件删文件移动文件
func (s *FileStorage) Init(data interface{}) {
	dat := data.(map[string]interface{})
	index, ok := dat[IndexPlain]
	if !ok {
		fmt.Printf("%s 没有找到index,重新记0", filepath.Dir(s.path))
		return
	}
	s.index = int(index.(float64))
}

func (s *FileStorage) SaveState() interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	return map[string]interface{}{
		IndexPlain: s.index,
	}
}

// 主动创建dir,返回要存储的文件路径,带文件名 0.mp4
func (s *FileStorage) GetFilePath() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	path := filepath.Join(s.path, strconv.Itoa(s.index))
	os.Mkdir(path, 0755)
	s.index++
	return path + "\\0.mp4"
}
