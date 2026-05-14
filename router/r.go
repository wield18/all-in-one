// Package router
package router

import (
	"github.com/gin-gonic/gin"

	"github.com/wield18/all-in-one/api/video"
	middleware "github.com/wield18/all-in-one/middle-ware"
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
	router.GET("/test", middleware.T)
	router.GET("/random-videos", video.V)
	router.GET("/video-index", video.VideoIndex) // 获得对应video 的 index
	router.GET("/video-slice", video.VideoSlice)

}
