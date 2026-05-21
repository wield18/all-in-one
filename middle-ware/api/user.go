package api

import (
	"github.com/gin-gonic/gin"
	"github.com/wield18/all-in-one/api/user"
	"github.com/wield18/all-in-one/entity"
)

func GetEmailCaptcha(c *gin.Context) {
	var email entity.CaptchaEmail
	if err := c.ShouldBindQuery(&email); err != nil {
		c.String(400, err.Error())
		return
	}
	user.GetEmailCaptcha(c, qqEmail, poolServer.Pool, &email)
}

func Register(c *gin.Context) {
	var userInfo entity.UserV1_allRequired
	// 解析查询参数到结构体
	if err := c.ShouldBind(&userInfo); err != nil {
		c.JSON(400, err.Error())
		return
	}
	user.Register(c, redisServer.RedisDb, mysqlServer.DB, &userInfo, poolServer.Pool)
}

func Update(c *gin.Context) {
	var userUpdateInfo entity.UserV1_update
	// 解析查询参数到结构体
	if err := c.ShouldBind(&userUpdateInfo); err != nil {
		c.JSON(400, err.Error())
		return
	}
	user.Update(c, redisServer.RedisDb, mysqlServer.DB, &userUpdateInfo, poolServer.Pool)
}

func LoginOut(c *gin.Context) {
	var query entity.UserV1_loginout
	// 解析查询参数到结构体
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, err.Error())
		return
	}
	user.LoginOut(c, redisServer.RedisDb, &query)
}

// 这里因为以后得做一下地区跟设备验证
// 所以先 name, password or email, password 先简单验证
// 如果通过 后端发送token 然后后端直接要求验证码 发个json就行 顺便发送验证码 用户输入验证码后 验证通过登录成功
// 所有临时token 验证码 都存redis 所有验证通过 所有都异步删除
func Login_pre(c *gin.Context) {
	var loginInfo entity.UserV1_passwordRequired
	// 解析查询参数到结构体
	if err := c.ShouldBind(&loginInfo); err != nil {
		c.JSON(400, err.Error())
		return
	}
	user.Login_pre(c, poolServer.Pool, mysqlServer.DB, &loginInfo, qqEmail)
}

func Login(c *gin.Context) {
	var query entity.UserV1_ExceptName
	// 解析查询参数到结构体
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, err.Error())
		return
	}
	user.Login(c, redisServer.RedisDb, mysqlServer.DB, &query, poolServer.Pool)
}

func Delete(c *gin.Context) {
	var query entity.UserV1_del
	// 解析查询参数到结构体
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, err.Error())
		return
	}
	user.Delete(c, redisServer.RedisDb, poolServer.Pool, mysqlServer.DB, &query, qqEmail)
}
