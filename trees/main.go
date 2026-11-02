// package main 生成结构用的
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// 需要跳过的目录/文件
var skipNames = map[string]bool{
	".git":         true,
	"node_modules": true,
	"dist":         true,
	"build":        true,
	".idea":        true,
	".vscode":      true,
}

func main() {
	root := `D:\Data\Golang\season5\all-in-one-git`
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "解析路径失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(filepath.Base(absRoot) + "/")
	if err := walk(absRoot, ""); err != nil {
		fmt.Fprintf(os.Stderr, "遍历失败: %v\n", err)
		os.Exit(1)
	}
}

func walk(dir, prefix string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// 排序：目录在前，文件在后，同类按名字排序
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return entries[i].Name() < entries[j].Name()
	})

	var visible []os.DirEntry
	for _, e := range entries {
		if skipNames[e.Name()] {
			continue
		}
		visible = append(visible, e)
	}

	for i, e := range visible {
		isLast := i == len(visible)-1
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		name := e.Name()
		if e.IsDir() {
			fmt.Println(prefix + connector + name + "/")
			nextPrefix := prefix
			if isLast {
				nextPrefix += "    "
			} else {
				nextPrefix += "│   "
			}
			if err := walk(filepath.Join(dir, name), nextPrefix); err != nil {
				return err
			}
		} else {
			fmt.Println(prefix + connector + name)
		}
	}
	return nil
}
