package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/wield18/all-in-one/config/consumer"
	"github.com/wield18/all-in-one/ekit/crypto"
	"github.com/wield18/all-in-one/ekit/fmtx"
	"github.com/wield18/all-in-one/entity"
	producerconsumer "github.com/wield18/all-in-one/my-ideas/producer-consumer"
	"gorm.io/gorm"
)

// 直接check user 基本验证password
func checkUser(db *gorm.DB, userInfo *entity.UserV1_passwordRequired) (*entity.UserV1, error) {
	if userInfo.Name == "" && userInfo.Email == "" {
		return nil, errors.New("empty name and email")
	}
	var aUser entity.UserV1

	result := db.Where(&entity.UserV1{Name: userInfo.Name, Email: userInfo.Email}).First(&aUser)
	if result.RowsAffected == 0 {
		return nil, errors.New("user not exist")
	}
	// 密码比较
	if !crypto.CheckPassword(userInfo.Password, aUser.Password) {
		return nil, errors.New("user name, email or password is wrong")
	}
	return &aUser, nil
}

// 就更新下密码
func updatePassword(db *gorm.DB, userUpdateInfo *entity.UserV1_update) (int, error) {
	var err error
	if err = db.Transaction(func(tx *gorm.DB) error {
		var aUser entity.UserV1
		result := db.Where(map[string]interface{}{entity.ID: userUpdateInfo.ID}).First(&aUser)
		if result.RowsAffected == 0 {
			return errNotExist
		}

		if !crypto.CheckPassword(userUpdateInfo.Password, aUser.Password) {
			return errValidatedWrong
		}
		password, _ := crypto.HashPassword(userUpdateInfo.NewPassword)
		result = db.Model(&aUser).Updates(entity.UserV1{Password: password})
		return result.Error
	}); err != nil {
		if err != errNotExist || err != errValidatedWrong {
			return 500, err
		} else {
			return 400, err
		}
	}
	return 200, nil
}

// 更新email 里边配合了redis的captcha获得跟删除
func updateEmail(rdb *redis.Client, db *gorm.DB, userUpdateInfo *entity.UserV1_update, pool *producerconsumer.ConsumerPool) (int, error) {
	ctx := context.TODO()
	// 查captcha
	val, err := rdb.Get(ctx, fmt.Sprintf("%s: %s", captcha, userUpdateInfo.NewEmail)).Result()
	if err != nil {
		return redisErrHandle_toError(err)
	}
	if val != userUpdateInfo.Captcha {
		return 400, err
	}

	// 再更新
	result := db.Model(&entity.UserV1{ID: userUpdateInfo.ID}).Updates(
		entity.UserV1{Email: userUpdateInfo.NewEmail})
	if result.Error != nil {
		return 500, err
	}
	if result.RowsAffected != 1 {
		return 400, err
	}
	// 最后在根据是否 NewEmail != "" 来异步删除 email的captcha
	pool.BlockSubmit(consumer.Topic_Del_Token, []string{fmtx.Sprintf(captcha, userUpdateInfo.NewEmail)}, func(consumeState int32, err error) {
		if err != nil {
			fmt.Println(err.Error())
		}
	})
	return 200, nil

}

// 返回http状态跟error 这里通常wrapper db,我这里的键名得传,所以说不好wrapper
