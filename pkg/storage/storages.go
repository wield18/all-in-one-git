package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type StorageManager struct {
	statePath string
	aMap      map[string]StorageInter
	byMap     map[string]interface{}
	running   bool
}

// 会自动初始化运行
func NewStorageManagerFromFileAndMap(file string, aMap map[string]StorageInter) *StorageManager {
	s := NewStorageManager()
	s.aMap = aMap
	s.InitFromFile(file)
	return s
}

func NewStorageManagerFromFile(file string) *StorageManager {
	s := NewStorageManager()
	s.InitFromFile(file)
	return s
}

func NewStorageManager() *StorageManager {
	return &StorageManager{aMap: make(map[string]StorageInter)}
}

func (s *StorageManager) InitFromFile(file string) {
	s.statePath = file
	by, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			dir := filepath.Dir(file)
			os.MkdirAll(dir, 0755)
			s.running = true
			fmt.Println("没有state,重新开始")
			return
		}
		panic(err)
	}
	aMap := map[string]interface{}{}
	err = json.Unmarshal(by, &aMap)
	if err != nil {
		panic(err)
	}
	// 在这里进行对应storage的初始化
	for k, v := range s.aMap {
		if data, ok := aMap[k]; ok {
			v.Init(data)
		}
	}

	s.running = true
	s.byMap = aMap
}

// 这里我也可以对于那种一次没开了的不能进行覆盖ok?
func (s *StorageManager) SaveState() {
	aMap := make(map[string]interface{})

	for k, aStorage := range s.aMap {
		aMap[k] = (aStorage.SaveState())
	}
	for k, v := range s.byMap {
		if _, ok := aMap[k]; !ok {
			aMap[k] = v
		}
	}
	if s.statePath == "" {
		fmt.Println(aMap)

	} else {

		// 有没有目录
		dir := filepath.Dir(s.statePath)
		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			os.MkdirAll(dir, 0755)
		}
		by, _ := json.Marshal(aMap)
		err = os.WriteFile(s.statePath, by, 0755)
		if err != nil {
			fmt.Println(err)
		}
	}
	s.running = false
}

func (s *StorageManager) Register(key string, storage StorageInter) {
	s.aMap[key] = storage
	if s.running {
		if s.byMap != nil {
			storage.Init(s.byMap[key])
		}
	}
}
