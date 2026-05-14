package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/wield18/all-in-one/config"
	"github.com/wield18/all-in-one/router"
	"github.com/wield18/all-in-one/template"
)

func main() {
	fmt.Println(config.Config)
	// 初始化template
	tempalteConfig := config.Config.Template
	template.InitTemplate(tempalteConfig.MinCount, tempalteConfig.MaxCount, tempalteConfig.VideoRoot)

	gin.SetMode(config.Config.Server.Model)

	router := router.InitRouter() // 返回 *gin.Engine
	// gin.Engine 实现了 http.Handler 接口
	srv := &http.Server{
		Addr:           config.Config.Server.Port,
		Handler:        router,
		ReadTimeout:    15 * time.Second, // 可以设置超时
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}
	// 启动服务
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("listen: %s \n", err)
		}
		log.Printf("listen: %s \n", config.Config.Server.Port)
	}()

	// 关闭服务
	quit := make(chan os.Signal, 1)
	// 监听消息
	signal.Notify(quit, os.Interrupt) // misuse of unbuffered os.Signal channel as argument to signal.Notify
	<-quit
	log.Println("Shutdown Server ...")
	// 创建上下文``
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Println("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}
