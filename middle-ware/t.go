// Package middleware 见文知意
package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/wield18/all-in-one/entity"
)

func T(c *gin.Context) {
	head := c.Request.Header
	for k, v := range head {
		fmt.Printf("key: %s, value: %s\n", k, v)
	}
	c.String(200, "ok")
}

// -----------------这里json绑定----------------
// 测试说明 Gin 会尝试用 HTTP json 参数的 key 匹配结构体的 字段名（大小写不敏感)
// IdString 能被匹配成 idstring IdString IDSTRING
// 可以多或少字段, 但不能类型不匹配
func T_json1(c *gin.Context) {
	t := &entity.WithoutJson{}
	if err := c.ShouldBindJSON(t); err != nil {
		c.String(400, err.Error())
		return
	}
	c.String(200, t.IdString)
}

func T_json2(c *gin.Context) {
	t := &entity.WithJson{}
	if err := c.ShouldBindJSON(t); err != nil {
		c.String(400, err.Error())
		return
	}
	c.String(200, t.IdString)
}

// -------------这里form表单绑定---------------
// IdString 只能匹配 IdString
// 可以多或少字段, 但不能类型不匹配
func T_form1(c *gin.Context) {
	t := &entity.WithoutForm{}
	if err := c.ShouldBind(t); err != nil {
		c.String(400, err.Error())
		return
	}
	c.String(200, t.IdString)
}

func T_form2(c *gin.Context) {
	t := &entity.WithForm{}
	if err := c.ShouldBind(t); err != nil {
		c.String(400, err.Error())
		return
	}
	c.String(200, t.IdString)
}
