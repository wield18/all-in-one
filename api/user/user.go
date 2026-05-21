// Package user
package user

import (
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/wield18/all-in-one/config/consumer"
	"github.com/wield18/all-in-one/ekit/crypto"
	"github.com/wield18/all-in-one/ekit/fmtx"
	"github.com/wield18/all-in-one/ekit/randx"
	"github.com/wield18/all-in-one/entity"
	producerconsumer "github.com/wield18/all-in-one/my-ideas/producer-consumer"
	emailqq "github.com/wield18/all-in-one/pkg/email-qq"
	"gorm.io/gorm"
)

var (
	errCaptchaWrong     = errors.New("captcha Wrong")
	errDbConflict       = errors.New("db Conflict")
	errNotExist         = errors.New("not exist")
	errValidatedWrong   = errors.New("validated wrong")
	errLackOrWrongData  = errors.New("lack or wrong data")
	errUserNoPermission = errors.New("user has no permission")
)

var (
	preLoginToken = "pre-login-token"
	captcha       = "captcha"
	userToken     = "user-token"
)

func GetEmailCaptcha(c *gin.Context, qqEmail *emailqq.QQEmail, pool *producerconsumer.ConsumerPool, email *entity.CaptchaEmail) {
	var err error
	code, _ := randx.RandCode(6, randx.TypeDigit)
	// 邮箱发送
	err = qqEmail.SendCaptcha(email.Email, code)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	// redis存储 使用本地异步池子发送
	err = pool.BlockSubmit(consumer.Topic_Captcha, &consumer.Captcha{Key: fmtx.Sprintf(captcha, email.Email), Captcha: code}, func(consumeState int32, err error) {})
	if err != nil {
		c.String(500, err.Error())
		return
	}
	c.String(200, "ok")
}

// 保持登录状态主要是本地存储的事，得需要前端本地存储登录信息，到后端redis啥的进行验证，如果网页刷新，本地存储不掉，但我的前端只是几张html
func Register(c *gin.Context, rdb *redis.Client, db *gorm.DB, userInfo *entity.UserV1_allRequired, pool *producerconsumer.ConsumerPool) {
	ctx := c.Request.Context()
	// 验证他的验证码
	// 这里使用redis对比
	if captcha, err := rdb.Get(ctx, fmt.Sprintf("%s: %s", captcha, userInfo.Email)).Result(); err != nil {
		c.String(redisErrHandle_toString(err))
		return
	} else {
		if captcha != userInfo.Captcha {
			c.String(400, errCaptchaWrong.Error())
			return
		}
	}

	password, _ := crypto.HashPassword(userInfo.Password)
	aUser := entity.UserV1{
		Name:      userInfo.Name,
		Password:  password,
		Email:     userInfo.Email,
		Thumbnail: []byte("猜猜这是啥"),
	}
	if result := db.Create(&aUser); result.Error != nil {
		c.String(409, fmt.Sprintf("%v: %s", errDbConflict, result.Error.Error()))
		return
	}

	// 删下captcha
	pool.BlockSubmit(consumer.Topic_Del_Token, []string{fmtx.Sprintf(captcha, userInfo.Email)}, func(consumeState int32, err error) {})

	c.String(200, "ok")

}

// 只有一个password需要查库看看一不一样, 其他随便
func Update(c *gin.Context, rdb *redis.Client, db *gorm.DB, userUpdateInfo *entity.UserV1_update, pool *producerconsumer.ConsumerPool) {

	ctx := c.Request.Context()
	// redis基本验证 只要他传过来的token找到的token是 如果他使用别人的id跟token,他如果还知道密码就能改密码
	token_rdb, err := rdb.Get(ctx, fmt.Sprintf("%s: %d", userToken, userUpdateInfo.ID)).Result()
	if err != nil {
		c.String(redisErrHandle_toString(err))
		return
	}
	if token := c.GetHeader(userToken); token_rdb != token {
		c.String(400, errCaptchaWrong.Error())
		return
	}

	// 一次只能改一个
	count := 0
	if userUpdateInfo.NewPassword != "" {
		code, err := updatePassword(db, userUpdateInfo)
		if err != nil {
			c.String(code, err.Error())
			return
		}
		c.String(200, "ok")
		return
	}
	if userUpdateInfo.NewEmail != "" {
		code, err := updateEmail(rdb, db, userUpdateInfo, pool)
		if err != nil {
			c.String(code, err.Error())
			return
		}
		c.String(200, "ok")
		return
	}
	// 这仨随便改了...
	if userUpdateInfo.NewName != "" {
		count++
	}
	if userUpdateInfo.NewStatus != "" {
		count++
	}
	if userUpdateInfo.NewThumbnail != "" {
		count++
	}
	if count != 1 {
		c.String(400, errLackOrWrongData.Error())
		return
	}

	// 再更新
	result := db.Model(&entity.UserV1{ID: userUpdateInfo.ID}).Updates(
		entity.UserV1{Name: userUpdateInfo.NewName,
			Email:     userUpdateInfo.NewEmail,
			Status:    userUpdateInfo.NewStatus,
			Thumbnail: []byte(userUpdateInfo.NewThumbnail)})
	if result.Error != nil {
		c.String(500, result.Error.Error())
		return
	}

	c.String(200, "ok")

}

// 我想着哈,他这个每次都得user验证一下,无论发没发token,是不是可以改成token替换 Password Email Name?
// 五个选项 选一个 name email password status Thumbnail
// 选择了 name status Thumbnail 直接输入就行
// email 输入新email 发验证就行
// password 输入旧密码跟新密码就行
// 需要 id name status Thumbnail email newEmail password NewPassword captcha
// 根本不需要pre

// 可以是可以但我们得去redis查值获得,应该得用到HSet,那他的token的value该用啥呢,user-token:16位随机数字,字母吗?,可以我只是说,不返回email了
// 说干就干,避免不了最后一次查库,但开头的check是可以避免的

// 真正login完成的部分,这里必有email
func Login(c *gin.Context, rdb *redis.Client, db *gorm.DB, loginInfo *entity.UserV1_ExceptName, pool *producerconsumer.ConsumerPool) {
	token := c.GetHeader(preLoginToken)
	if token == "" {
		c.String(403, fmtx.Sprintf(errUserNoPermission.Error(), "token"))
		return
	}
	ctx := c.Request.Context()
	// token 跟 captcha 检查
	vals, err := rdb.MGet(ctx, fmt.Sprintf("%s: %s", preLoginToken, loginInfo.Email),
		fmt.Sprintf("%s: %s", captcha, loginInfo.Email)).Result()
	if err != nil {
		c.String(redisErrHandle_toString(err))
		return
	}
	if vals[0] != token || vals[1] != loginInfo.Captcha {
		c.String(403, fmtx.Sprintf(errUserNoPermission.Error(), "token or captcha is wrong\n"),
			fmtx.SprintfInterfaces(vals[0], token, vals[1], loginInfo.Captcha))
		return
	}

	// 最后一次user查库
	var aUser entity.UserV1
	result := db.Where(&entity.UserV1{Email: loginInfo.Email}).First(&aUser)
	if result.RowsAffected == 0 {
		c.String(404, fmt.Sprintf("%v: %s", errNotExist, result.Error.Error()))
		return
	}
	// 密码比较
	if !crypto.CheckPassword(loginInfo.Password, aUser.Password) {
		c.String(400, fmt.Sprintf("%v: %s", errValidatedWrong, "user or password is wrong"))
		return
	}

	// 也得生成用户token
	token, _ = randx.RandCode(16, randx.TypeDigit|randx.TypeLowerCase)
	aStruct := &consumer.Token{Key: fmt.Sprintf("%s: %d", userToken, aUser.ID), Token: token, Timeout: time.Hour} // 存redis时携带id
	// 存token
	if err := pool.BlockSubmit(consumer.Topic_Token, aStruct, func(consumeState int32, err error) {}); err != nil {
		c.String(500, err.Error())
		return
	}

	// 最后删一下redis对应的key,无论正确与否
	pool.BlockSubmit(consumer.Topic_Del_Token, []string{fmtx.Sprintf(preLoginToken, loginInfo.Email),
		fmtx.Sprintf(captcha, loginInfo.Email)}, func(consumeState int32, err error) {})

	c.JSON(
		200, gin.H{userToken: token, entity.ID: aUser.ID},
	)

}

// todo: 第一次验证可以直接打db去 也可以先查redis 现在只查db
func Login_pre(c *gin.Context, pool *producerconsumer.ConsumerPool, db *gorm.DB, loginInfo *entity.UserV1_passwordRequired, qq *emailqq.QQEmail) {
	aUser, err := checkUser(db, loginInfo)
	if err != nil {
		c.String(400, fmtx.Sprintf(errValidatedWrong.Error(), err.Error()))
		return
	}
	// 是否生成token, 这得判断机子跟地区 现在先直接生成
	if needToken := CheckSafety(loginInfo, aUser); needToken {
		// 生成token
		val, _ := randx.RandCode(16, randx.TypeDigit|randx.TypeLowerCase)
		token := &consumer.Token{Key: fmt.Sprintf("%s: %s", preLoginToken, aUser.Email), Token: val} // 存redis时携带email
		// 存token
		if err := pool.BlockSubmit(consumer.Topic_Token, token, func(consumeState int32, err error) {}); err != nil {
			c.String(500, err.Error())
			return
		}

		code, _ := randx.RandCode(6, randx.TypeDigit)
		// 发验证码得
		if err := qq.SendCaptcha(aUser.Email, code); err != nil {
			c.String(500, err.Error())
			return
		}
		// 存验证码
		if err := pool.BlockSubmit(
			consumer.Topic_Captcha, &consumer.Captcha{Key: fmt.Sprintf("%s: %s", captcha, aUser.Email), Captcha: code},
			func(consumeState int32, err error) {}); err != nil {
			c.String(500, err.Error())
			return
		}
		c.JSON(200, gin.H{
			entity.Email:  aUser.Email,
			preLoginToken: val,
		})
		return
	}

	// 不需要
	// 也得生成用户token,只是token得存下用户信息
	token, _ := randx.RandCode(16, randx.TypeDigit|randx.TypeLowerCase)
	aStruct := &consumer.Token{Key: fmt.Sprintf("%s: %d", userToken, aUser.ID), Token: token, Timeout: time.Hour} // 存redis时携带id
	// 存token
	if err := pool.BlockSubmit(consumer.Topic_Token, aStruct, func(consumeState int32, err error) {}); err != nil {
		c.String(500, err.Error())
		return
	}
	c.JSON(
		200, gin.H{userToken: token, entity.ID: aUser.ID},
	)

}

func CheckSafety(loginInfo *entity.UserV1_passwordRequired, aUser *entity.UserV1) bool {
	return true
}

// 删除本地cookie,redis直接同步删,结束
func LoginOut(c *gin.Context, rdb *redis.Client, id_del *entity.UserV1_loginout) {
	ctx := c.Request.Context()
	token := c.GetHeader(userToken)
	id := id_del.ID
	if token == "" || id == 0 {
		c.String(400, errLackOrWrongData.Error())
		return
	}
	token_rds, err := rdb.Get(ctx, fmtx.Sprintf(userToken, fmt.Sprint(id))).Result()
	if err != nil {
		c.String(redisErrHandle_toString(err))
		return
	}
	if token != token_rds {
		c.String(403, fmtx.Sprintf(errUserNoPermission.Error(), "token"))
		return
	}
	err = rdb.Del(ctx, fmtx.Sprintf(userToken, fmt.Sprint(id))).Err()
	if err != nil {
		c.String(redisErrHandle_toString(err))
		return
	}

	c.String(200, "ok")
}

// 这是删除用户哈
// qq直接就能申请注销
// 这里直接跟注册一样的只不过全部要求captcha的验证吧,基本的password,需要用户名或email,如果用户名,就得查询库获得(可以redis或mysql),然后就是申请验证码
// 这里跟login做法其实基本一致, 只是必须申请验证码
// 这里delete直接必须输入email完事吧
// id email password captcha 这四个还有 userToken 验证一下 就行
func Delete(c *gin.Context, rdb *redis.Client, pool *producerconsumer.ConsumerPool, db *gorm.DB, deleteInfo *entity.UserV1_del, qq *emailqq.QQEmail) {
	ctx := c.Request.Context()
	token := c.GetHeader(userToken)
	id := deleteInfo.ID
	// 基本验证一下
	if token == "" || id == 0 {
		c.String(400, errLackOrWrongData.Error())
		return
	}
	token_rds, err := rdb.Get(ctx, fmtx.Sprintf(userToken, fmt.Sprint(id))).Result()
	if err != nil {
		c.String(redisErrHandle_toString(err))
		return
	}
	if token != token_rds {
		c.String(403, fmtx.Sprintf(errUserNoPermission.Error(), "token"))
		return
	}
	if err = db.Transaction(func(tx *gorm.DB) error {
		var aUser entity.UserV1
		result := tx.Where(&entity.UserV1{ID: id}).First(&aUser)
		if result.RowsAffected == 0 {
			return errNotExist
		}
		if result.Error != nil {
			return result.Error
		}

		result = tx.Delete(&entity.UserV1{}, id)
		if result.Error != nil {
			return result.Error
		}
		return nil
	}); err != nil {
		if err == errNotExist {
			c.String(404, err.Error())
			return
		}
		c.String(500, err.Error())
		return
	}
	c.String(200, "ok")
}
