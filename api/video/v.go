package video

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/wield18/all-in-one/entity"
	"github.com/wield18/all-in-one/template"
)

func V(c *gin.Context) {
	var query entity.Window
	// 解析查询参数到结构体
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, err.Error())
		return
	}
	str := template.GenerateVideoTemplate(query.Height, query.Width)
	fmt.Println("---------------------------")
	fmt.Println(query.Width, query.Height)
	fmt.Println(str)
	c.String(http.StatusOK, str)
}

func VideoIndex(c *gin.Context) {
	var query entity.VideoIndex
	// 解析查询参数到结构体
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, err.Error())
		return
	}
	fmt.Println("query.Index: ", query.Index)
	fmt.Println("query.Resolution: ", query.Resolution)
	byt, _ := os.ReadFile(`D:\Data\Golang\season5\all-in-one\data\output.m3u8`)
	c.Header("Content-Type", "application/vnd.apple")
	c.String(200, string(byt))
}

func VideoSlice(c *gin.Context) {
	var query entity.VideoSlice
	// 解析查询参数到结构体
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, err.Error())
		return
	}
	fmt.Println("query.Index: ", query.Index)
	fmt.Println("query.Resolution: ", query.Resolution)
	fmt.Println("query.Time: ", query.Time)

	slicePath := fmt.Sprintf(`D:\Data\Golang\season5\all-in-one\data\output%s.ts`, query.Time)
	c.File(slicePath)
}
