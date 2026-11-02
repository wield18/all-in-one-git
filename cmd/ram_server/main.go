package ramserver

import (
	"all-in-one/config"
	"all-in-one/internal/base"
	consumerouter "all-in-one/internal/consume_router"
	"all-in-one/internal/entity/validate"
	"all-in-one/internal/local"
	"all-in-one/internal/router"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

func ServerReq(r *gin.Engine, req *http.Request) (int, string) {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fmt.Println("GET 响应:")
	fmt.Println("状态码:", w.Code)
	fmt.Println("响应体:", w.Body.String())
	return w.Code, w.Body.String()
}

// 会主动启动所有的base服务
func NewRouter() *gin.Engine {
	base.InitAll()
	v := validate.New()
	p := consumerouter.NewConsumePool()
	base.SetRPV(p, v)
	local.InitLocal()
	return router.NewRouter(config.Conf.Server)
}

// 测试使用需传递配置文件路径
func NewRouterAndInitConf(confPath string) *gin.Engine {
	config.Init(confPath)
	base.InitAll()
	v := validate.New()
	p := consumerouter.NewConsumePool()
	base.SetRPV(p, v)
	local.InitLocal()
	return router.NewRouter(config.Conf.Server)
}
