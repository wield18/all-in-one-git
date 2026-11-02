package main

import (
	"all-in-one/cmd/test_setver/internal"
	"all-in-one/config"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

var conf = config.Conf

func TestMain(*testing.T) {
	router := internal.InitRouter()
	// _, _, _, _, _, _ := StartAllService(conf, servers{slogv: true})
	// path := internal.Handle200Get.Path
	// internal.ServerReq(router, httptest.NewRequest(internal.Handle200Get.Method, path, nil))

	// 2. 序列化成 JSON 字节数组，并转成 Reader
	jsonData, _ := json.Marshal(internal.Gaojian{Str: "gaojian"})
	body := bytes.NewBuffer(jsonData)
	req := httptest.NewRequest(internal.HandleStructPost.Method, internal.HandleStructPost.Path, body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("ID", "application/json")

	internal.ServerReq(router, req)
}

func TestAbc(*testing.T) {
	abc, err := strconv.ParseInt("", 10, 64)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(abc)
	}
}

func TestForm(*testing.T) {
	router := internal.InitRouter()
	formData := url.Values{}
	formData.Set("abc", "zhangsan")

	// 2. 编码成字符串并转成 Reader
	encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
	fmt.Println(encoded)
	body := strings.NewReader(encoded)

	path := internal.HandleFormAndQuery.Path + "?gaojian=gaojianQuery"

	// 发送get
	req := httptest.NewRequest(internal.HandleFormAndQuery.Method, path, body)
	req.Header.Set("gaojian1", "abc")
	req.Header.Set("ID", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	internal.ServerReq(router, req)
}

type Goajian1 struct {
	Str  string `json:"str"`
	Str2 string `json:"str1"`
}

type Goajian2 struct {
	Str  string `json:"str"`
	Str2 string `json:"str1"`
	Str3 string `json:"str2"`
}

func TestJsonUnmarshal(*testing.T) {
	abc1 := &Goajian1{}
	abc := Goajian2{Str3: "dwdwd"}
	by, _ := json.Marshal(abc)
	fmt.Println(string(by))
	err := json.Unmarshal(by, abc1)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(abc1)
	}
}

func TestFileForm(*testing.T) {
	router := internal.InitRouter()
	// 1. 准备文件内容（实际从磁盘读）
	fileContent := []byte("hello world, this is file content")
	filename := "test.txt"

	// 2. 客户端算 SHA256
	sum := sha256.Sum256(fileContent)
	sha256Hex := hex.EncodeToString(sum[:]) // 64 字符 hex

	// 3. 构造 multipart body
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	// 3.1 文件 part
	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		panic(err)
	}
	if _, err := fw.Write(fileContent); err != nil {
		panic(err)
	}

	// 3.2 普通字段 part
	_ = w.WriteField("sha256", sha256Hex)
	_ = w.WriteField("title", "我的文件")
	_ = w.WriteField("user_id", "1001")
	_ = w.WriteField("remark", "测试上传")

	// 4. 关闭 writer，写入 boundary 结尾
	if err := w.Close(); err != nil {
		panic(err)
	}

	// 5. 发请求
	req, err := http.NewRequest(internal.HandleFile.Method, internal.HandleFile.Path, &buf)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("ID", "dwd")

	internal.ServerReq(router, req)

}

func TestSize(*testing.T) {
	// 1. 准备文件内容（实际从磁盘读）
	// fileContent := []byte("hello world, this is file content")
	// filename := "test.txt"

	// multipart.NewWriter()
	// // 2. 客户端算 SHA256
	// sum := sha256.Sum256(fileContent)
	// sha256Hex := hex.EncodeToString(sum[:]) // 64 字符 hex

	// // 3. 构造 multipart body
	// var buf bytes.Buffer
	// w := multipart.NewWriter(&buf)

	// // 3.1 文件 part
	// fw, err := w.CreateFormFile("file", filename)
	// if err != nil {
	// 	panic(err)
	// }
	// if _, err := fw.Write(fileContent); err != nil {
	// 	panic(err)
	// }

	// filePath := `D:\Animation\Minus8_Archive 一个性的时代的代表\Minus8 Archive\ANIMATIONS\!MP4s\EehFwLdU4AAthXv.mp4`
	// fileName := filepath
	// fmt.Println(fileName)

	// 目录: "/home/user/", 文件名: "file.txt"
}

// // 1. 构造请求体结构体
// reqBody := map[string]interface{}{
// 	"abc":      "zhangsan",
// 	"password": "123456",
// 	"age":      25,
// }

// // 2. 序列化成 JSON 字节数组，并转成 Reader
// jsonData, err := json.Marshal(reqBody)
// if err != nil {
// 	return
// }
// body := bytes.NewBuffer(jsonData)

// formData := url.Values{}
// formData.Set("abc", "zhangsan")
// formData.Set("def", "zhangsan")
// formData.Set("def", "zhangsan")

// // 2. 编码成字符串并转成 Reader
// encoded := formData.Encode() // 结果: "username=zhangsan&password=123456&age=25"
// fmt.Println(encoded)
// body := strings.NewReader(encoded)

// email := "3963099642@qq.com"

// path := "/user/captcha?email=" + email

// 发送get
// req := httptest.NewRequest("POST", path, body)
// req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
