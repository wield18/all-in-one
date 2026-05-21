// 基本结构
package consumer

import (
	"fmt"
	"time"
)

// topic
const (
	TopicCaptcha = "captcha"
)

// 基本redis存储的key-val结构, timeout 默认 5*time.Minute
type Token struct {
	Key     string        `json:"key"`
	Token   string        `json:"token"`
	Timeout time.Duration `json:"timeout"`
}

func (t *Token) FormatKey(key interface{}) *Token {
	if t.Key != "" {
		t.Key = fmt.Sprintf("%s: %v", t.Key, key)
	} else {
		t.Key = fmt.Sprintf("%v", key)
	}
	return t
}

// RedisSetCaptcha 真正绑定的数据
type Captcha struct {
	Key     string
	Captcha string
}

// 存储结构体, 我得写个interface都得有个ToMapToken的方法, 这里struct使用指针传递 默认timeout 5*time.Second
type RedisStruct struct {
	Key     string
	Token   string
	Struct  CanToMapToken
	Timeout time.Duration
}

type CanToMapToken interface {
	ToMapToken(token string) map[string]string
}
