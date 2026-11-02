package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

var (
	get  = "GET"
	post = "POST"
)

func InitRouter() *gin.Engine {
	router := gin.New()
	router.MaxMultipartMemory = 500 << 20
	// 跌机是恢复

	router.Use(gin.Recovery())
	router.Use(IDValidate)
	register(router)
	return router
}
func IDValidate(c *gin.Context) {
	id := c.GetHeader("ID")
	if id == "" { // 没有id不是吗
		c.AbortWithStatusJSON(403, struct{ Gao string }{Gao: "gaojian"})
		return
	}
	c.Set("ID", id)
	c.Next()
}

func ServerReq(r *gin.Engine, req *http.Request) (int, string) {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fmt.Println("GET 响应:")
	fmt.Println("状态码:", w.Code)
	fmt.Println("响应体:", w.Body.String())
	return w.Code, w.Body.String()
}

type PathAndMethod struct {
	Path   string
	Method string
}

var (
	Handle200Get       = PathAndMethod{Path: "/200", Method: get}
	HandleAbcPost      = PathAndMethod{Path: "/abc", Method: post}
	HandleStructPost   = PathAndMethod{Path: "/struct", Method: post}
	HandleFormAndQuery = PathAndMethod{Path: "/FormAndQuery", Method: post}
	HandleFile         = PathAndMethod{Path: "/file", Method: post}
)

func register(r *gin.Engine) {
	r.GET(Handle200Get.Path, Handle)
	r.POST(HandleAbcPost.Path, abcJsonBind(&Gaojian{}), abcHandle)
	r.POST(HandleStructPost.Path, jsonBind(&Gaojian{}), structHandle)
	r.POST(HandleFormAndQuery.Path, Result, PostHandleFormAndQuery, Abc)
	r.POST(HandleFile.Path, UploadHandler)
}

func Abc(c *gin.Context) {
	fmt.Println("abcnooooooooooooooooooooooo")
}

func Handle(c *gin.Context) {
	c.JSON(200, "gaojian")
}

type Gaojian struct {
	Str string `json:"str"`
}

func (g *Gaojian) ToData() interface{} {
	return &gaojianData{Str: g.Str}
}

type gaojianData struct {
	Str string `json:"str"`
}

var (
	data = "data"
)

type form struct {
	Abc string `form:"abc"`
}

type AcceptType interface {
	ToData() interface{}
}

func abcJsonBind[v1 AcceptType](acceptStruct v1) gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		if err = c.ShouldBindJSON(acceptStruct); err != nil {
			c.JSON(400, "")
			c.Abort()
			return
		}
		c.Set(data, acceptStruct.ToData())
		c.Next()
	}
}

// c.ShouldBindJSON(acceptStruct) 这破方法转map了简直了,可能是看到我这个指针指向interface吧
func jsonBind(acceptStruct any) gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		if err = c.ShouldBindJSON(acceptStruct); err != nil {
			c.JSON(400, "")
			c.Abort()
			return
		}
		c.Set(data, acceptStruct)
		c.Next()
	}
}

func abcHandle(c *gin.Context) {
	abc := c.MustGet(data).(*gaojianData)
	c.JSON(200, abc)
}

func structHandle(c *gin.Context) {
	abc := c.MustGet(data).(*Gaojian)
	ab := c.GetHeader("gaojian")
	fmt.Println(ab)
	c.JSON(200, abc)
}

// func NewReq(method string, path string, body io.Reader) {
// 	req := httptest.NewRequest(Get, path, nil)
// }

func Result(c *gin.Context) {
	c.Next()
	c.AbortWithStatusJSON(200, "ok")
}

func PostHandleFormAndQuery(c *gin.Context) {
	// abc, ok := c.GetQuery("gaojian")
	// abc1 := c.GetHeader("gaojian1")
	// abcd := &form{}
	// err := c.ShouldBind(abcd)
	// if !ok || abc1 == "" || err != nil {
	// 	c.AbortWithStatusJSON(400, "400")
	// 	return
	// }
	// c.AbortWithStatusJSON(200, fmt.Sprint(abc, abc1, abcd.Abc))

}

// UploadReq 上传请求：文件 + 表单字段 + SHA256 防伪
type UploadReq struct {
	// 文件本体，Gin 用 *multipart.FileHeader 接收
	File *multipart.FileHeader `form:"file" binding:"required"`

	// 客户端声明的 SHA256（hex，64 字符），服务端要重新算一遍比对
	SHA256 string `form:"sha256" binding:"required,len=64"`

	// 其他业务字段，随便加
	Title  string `form:"title" binding:"max=128"`
	UserID int64  `form:"user_id" binding:"required,gt=0"`
	Remark string `form:"remark" binding:"max=512"`
}

func UploadHandler(c *gin.Context) {
	var req UploadReq

	// 1. 绑定（文件 + 表单字段一起）
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. 打开文件
	f, err := req.File.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "open file failed"})
		return
	}
	defer f.Close()

	// 3. 读取全部内容（大文件建议流式，见后文）
	data, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "read file failed"})
		return
	}

	// 4. 服务端重新算 SHA256
	sum := sha256.Sum256(data)
	serverHash := hex.EncodeToString(sum[:])

	// 5. 比对（大小写无关，统一小写）
	if serverHash != req.SHA256 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "sha256 mismatch",
			"client_hash": req.SHA256,
			"server_hash": serverHash,
		})
		return
	}

	// 6. 校验通过，后续处理：存盘 / 入库 / 上传 OSS ...
	//    这里只打印拿到的所有字段
	c.JSON(http.StatusOK, gin.H{
		"filename": req.File.Filename,
		"size":     req.File.Size,
		"sha256":   serverHash,
		"title":    req.Title,
		"user_id":  req.UserID,
		"remark":   req.Remark,
		"matched":  true,
	})
}
