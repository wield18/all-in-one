// Package api 所有api的基本绑定跟依赖注入
package api

import (
	producerConsumer "github.com/wield18/all-in-one/my-ideas/producer-consumer"
	emailqq "github.com/wield18/all-in-one/pkg/email-qq"
	"github.com/wield18/all-in-one/pkg/mysql"
	"github.com/wield18/all-in-one/pkg/redisx"
)

var qqEmail = emailqq.GetQQEmail()
var poolServer = producerConsumer.GetPoolServer()
var mysqlServer = mysql.GetDBServer()
var redisServer = redisx.GetRedisServerSingle()
