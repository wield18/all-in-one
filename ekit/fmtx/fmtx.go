// Package fmtx 对fmt标准库简单封装
package fmtx

import "fmt"

// 这里可以啥都不传
func Sprintf(vals ...string) string {
	key := ""
	for _, v := range vals {
		key += v + ": "
	}
	if key == "" {
		return key
	}
	return key[:len(key)-2]
}

func SprintfInterfaces(vals ...interface{}) string {
	key := ""
	for i := 0; i < len(vals); i++ {
		key += "%v: "
	}
	if key == "" {
		return key
	}
	key = key[:len(key)-2]
	return fmt.Sprintf(key, vals...)
}

// 传入error不支持其他类型
// 不做任何检查
func Errorf(vals ...interface{}) error {
	key := ""
	for i := 0; i < len(vals); i++ {
		key += "%w: "
	}
	if key == "" {
		var t error
		return t
	}
	key = key[:len(key)-2]
	return fmt.Errorf(key, vals...)
}
