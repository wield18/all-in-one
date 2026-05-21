// Package redis
package redisx

import (
	"context"
	"fmt"
	"time"

	"github.com/wield18/all-in-one/config"

	"github.com/redis/go-redis/v9"
)

type RedisServerSingle struct {
	RedisDb *redis.Client
}

// 启动多个redis实例
var rdb = &RedisServerSingle{}

func GetRedisServerSingle() *RedisServerSingle {
	return rdb
}

func InitRedisServerSingle(config *config.Redis) *RedisServerSingle {
	// 创建 Redis 客户端
	rdb.RedisDb = redis.NewClient(&redis.Options{
		Addr:     config.Address,  // Redis 地址
		Password: config.Password, // 密码（与 redis.conf 中的 requirepass 一致）
		DB:       config.Db,       // 使用哪个数据库（0-15）

		// 可选：连接池配置
		PoolSize:     config.PoolSize, // 连接池大小
		MinIdleConns: 5,               // 最小空闲连接数
		MaxRetries:   3,               // 最大重试次数

		// 超时配置
		DialTimeout:  5 * time.Second, // 连接超时
		ReadTimeout:  3 * time.Second, // 读超时
		WriteTimeout: 3 * time.Second, // 写超时
	})

	// 测试连接
	ctx := context.Background()
	err := rdb.RedisDb.Ping(ctx).Err()
	if err != nil {
		panic(fmt.Sprintf("Redis 连接失败: %v", err))
	}
	fmt.Println("Redis 连接成功!")
	return rdb

}

func (rdb *RedisServerSingle) Shutdown() error {
	// 使用完后关闭连接
	// defer rdb.Close()
	rdb.RedisDb.Close()
	return nil
}
