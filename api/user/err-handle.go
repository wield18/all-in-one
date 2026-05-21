package user

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"github.com/wield18/all-in-one/ekit/fmtx"
)

var (
	errRedisServer = errors.New("redis server error") // Redis服务端错误
	errUnknown     = errors.New("unknown err")        // 未知err
)

// 返回http状态跟error,前提必须做nil检查,使用这个方法已经确定是个err
// 这里有context的检测主要还是它需要传context
func redisErrHandle_toString(err error) (int, string) {
	// 1. 检查是否是预期的“空值”
	if errors.Is(err, redis.Nil) {
		return 404, err.Error()
	}
	// 2. (v9.17.0+) 检查类型错误，
	var redisErr redis.Error
	if errors.As(err, &redisErr) {
		// 如果是 Redis 服务端返回的错误（如 WRONGTYPE）
		return 500, fmtx.SprintfInterfaces(errRedisServer, redisErr) // 【Redis服务端错误】
	}
	// 3. 处理连接、超时、上下文取消等系统错误
	if errors.Is(err, context.Canceled) {
		return 503, context.Canceled.Error() // 【客户端】请求已被取消
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return 504, context.DeadlineExceeded.Error() // 【超时】请求超时，建议稍后重试
	}
	return 500, fmtx.SprintfInterfaces(errUnknown, err)
}

func redisErrHandle_toError(err error) (int, error) {
	// 1. 检查是否是预期的“空值”
	if errors.Is(err, redis.Nil) {
		return 404, err
	}
	// 2. (v9.17.0+) 检查类型错误，
	var redisErr redis.Error
	if errors.As(err, &redisErr) {
		// 如果是 Redis 服务端返回的错误（如 WRONGTYPE）
		return 500, fmtx.Errorf(errRedisServer, redisErr) // 【Redis服务端错误】
	}
	// 3. 处理连接、超时、上下文取消等系统错误
	if errors.Is(err, context.Canceled) {
		return 503, context.Canceled // 【客户端】请求已被取消
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return 504, context.DeadlineExceeded // 【超时】请求超时，建议稍后重试
	}
	return 500, fmtx.Errorf(errUnknown, err)
}
