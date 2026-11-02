package httpserver

import (
	"all-in-one/config"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"all-in-one/internal/base"
	consumerouter "all-in-one/internal/consume_router"
	"all-in-one/internal/entity/validate"
	"all-in-one/internal/local"
	"all-in-one/internal/router"

	"github.com/gin-gonic/gin"
)

func NewRouter(conf config.Server) *gin.Engine {
	base.InitAll()
	v := validate.New()
	p := consumerouter.NewConsumePool()
	base.SetRPV(p, v)
	local.InitLocal()
	return router.NewRouter(conf)
}

// 启动所有服务跟http网关
func StartGin(conf config.Server) {
	gin.SetMode(conf.Model)

	router := NewRouter(conf) // 返回 *gin.Engine
	// gin.Engine 实现了 http.Handler 接口
	srv := &http.Server{
		Addr:           conf.Port,
		Handler:        router,
		ReadTimeout:    0, // 可以设置超时
		WriteTimeout:   0,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}
	// 启动服务
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("listen: %s \n", err)
		}
		log.Printf("listen: %s \n", conf.Port)
	}()

	// 关闭服务
	quit := make(chan os.Signal, 1)
	// 监听消息
	signal.Notify(quit, os.Interrupt) // misuse of unbuffered os.Signal channel as argument to signal.Notify
	<-quit
	log.Println("Shutdown Server ...")
	// 创建上下文
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Println("Server Shutdown:", err)
	}
	log.Println("Server exiting")

}
