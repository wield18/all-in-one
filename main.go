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
	// consumer "github.com/wield18/all-in-one/my-ideas/producer-consumer"
	// emailqq "github.com/wield18/all-in-one/pkg/email-qq"
	// "github.com/wield18/all-in-one/pkg/mysql"
	// "github.com/wield18/all-in-one/pkg/redisx"
	"github.com/wield18/all-in-one/router"
	"github.com/wield18/all-in-one/template"
)

func main() {
	conf := config.GetConfig()
	// conf的美化输出是可以用reflect写
	fmt.Println(conf)

	// 初始化template
	tempalteConfig := config.GetConfig().Template
	template.InitTemplate(tempalteConfig.MinCount, tempalteConfig.MaxCount, tempalteConfig.VideoRoot)

	// 其他服务
	// mysql.InitMDBServer(&conf.Mysql)
	// redisx.InitRedisServerSingle(&conf.Redis)
	// emailqq.InitQQEmail(&conf.QQ)
	// consumer.InitPool(&conf.Pool)

	// 启动并监听gin
	StartGin(&conf.Server)

}

// 启动并监听gin
func StartGin(conf *config.Server) {
	gin.SetMode(conf.Model)

	router := router.InitRouter() // 返回 *gin.Engine
	// gin.Engine 实现了 http.Handler 接口
	srv := &http.Server{
		Addr:           conf.Port,
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
		log.Printf("listen: %s \n", conf.Port)
	}()

	// 关闭服务
	quit := make(chan os.Signal, 1)
	// 监听消息
	signal.Notify(quit, os.Interrupt) // misuse of unbuffered os.Signal channel as argument to signal.Notify
	<-quit
	log.Println("Shutdown Server ...")
	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Println("Server Shutdown:", err)
	}
	log.Println("Server exiting")

}
