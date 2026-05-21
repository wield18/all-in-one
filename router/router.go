// Package router
package router

import (
	"github.com/gin-gonic/gin"

	"github.com/wield18/all-in-one/api/video"
	middleware "github.com/wield18/all-in-one/middle-ware"
	"github.com/wield18/all-in-one/middle-ware/api"
)

func InitRouter() *gin.Engine {
	router := gin.New()
	// 跌机是恢复
	router.Use(gin.Recovery())
	router.Use(middleware.Cors())
	register(router)
	return router
}

func register(router *gin.Engine) {

	// post
	router.POST("/delete", api.Delete)
	router.POST("/update", api.Update)
	router.POST("/login-pre", api.Login_pre)
	router.POST("/login", api.Login)
	router.POST("/register", api.Register)
	router.POST("/test1", middleware.T_json1)
	router.POST("/test2", middleware.T_json2)
	router.POST("/test3", middleware.T_form1)
	router.POST("/test4", middleware.T_form2)

	// get
	router.GET("/login-out", api.LoginOut)
	router.GET("/test", middleware.T)
	router.GET("/random-videos", video.V)
	router.GET("/video-index", video.VideoIndex) // 获得对应video 的 index
	router.GET("/video-slice", video.VideoSlice)
	router.GET("/captcha", api.GetEmailCaptcha)
}
