// Package consumer 主要用于本地异步池的consumerFunc的注册,也就是这里写所有的topic跟func
package consumer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wield18/all-in-one/my-ideas/producer-consumer/consumer"
	"github.com/wield18/all-in-one/pkg/redisx"
)

var rdb = redisx.GetRedisServerSingle()

var (
	errAssert = errors.New("Assert Wrong")
)

// Captcha
var Topic_Captcha = "captcha"

// Token 基本token
var Topic_Token = "token"

// 测试
var Topic_TestRedisWrapper = "test-redis-wrapper"

// RedisStruct
var Topic_StructToken = "struct-token"

// Del Token
var Topic_Del_Token = "del-token"

func RedisSetCaptcha(ctx context.Context, data *Captcha, db *redis.Client) error {
	err := db.Set(ctx, data.Key, data.Captcha, time.Minute*5).Err()
	if err != nil {
		return err
	}
	fmt.Printf("成功设置: %s---%s\n", data.Key, data.Captcha)
	return err
}

func RedisSetToken(ctx context.Context, data *Token, db *redis.Client) error {
	var err error
	if data.Timeout != 0 {
		err = db.Set(ctx, data.Key, data.Token, data.Timeout).Err()
	} else {
		err = db.Set(ctx, data.Key, data.Token, 5*time.Minute).Err()
	}
	if err != nil {
		return err
	}
	fmt.Printf("成功设置: %s---%s\n", data.Key, data.Token)
	return err
}

// 设置struct带着token 默认过期5分钟
func RedisSetStructWithToken(ctx context.Context, data *RedisStruct, db *redis.Client) error {
	aMap := data.Struct.ToMapToken(data.Token)

	// 创建pipe
	pipe := db.Pipeline()
	pipe.HSet(ctx, data.Key, aMap)
	if data.Timeout != 0 {
		pipe.Expire(ctx, data.Key, data.Timeout)
	} else {
		pipe.Expire(ctx, data.Key, 5*time.Minute)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("成功设置: %s---%v\n", data.Key, aMap)
	return nil
}

func RedisDelToken(ctx context.Context, key []string, db *redis.Client) error {
	return db.Del(ctx, key...).Err()
}

// 所谓的redisSet泛型wrapper
type RedisOperationFuncWrapper[T any] struct {
	process func(context.Context, T, *redis.Client) error
}

func (r *RedisOperationFuncWrapper[T]) Consume(ctx context.Context) error {
	db := rdb.RedisDb
	// 这不是得存Token
	data := ctx.Value(consumer.DATA)
	if token, ok := data.(T); ok {
		return r.process(ctx, token, db)
	} else {
		fmt.Printf("断言失败，data 是: %+v\n", data) // 直接打印，会自动解引用
		fmt.Printf("data 类型: %T\n", data)
		return fmt.Errorf("%w: %T\n", errAssert, data)
	}
}

var aMapConsumer = map[string]consumer.Consumer{
	Topic_Token: &RedisOperationFuncWrapper[*Token]{
		process: RedisSetToken,
	},
	Topic_Captcha: &RedisOperationFuncWrapper[*Captcha]{
		process: RedisSetCaptcha,
	},
	Topic_StructToken: &RedisOperationFuncWrapper[*RedisStruct]{
		process: RedisSetStructWithToken,
	},
	Topic_Del_Token: &RedisOperationFuncWrapper[[]string]{
		process: RedisDelToken,
	},
}

func GetConsumerMap() map[string]consumer.Consumer {
	return aMapConsumer
}
