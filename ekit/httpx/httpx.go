// Package httpx http包的简单封装
package httpx

import (
	"net/http"
)

func GetHeaderValue(header http.Header, key string) []string {
	for k, v := range header {
		if k == key {
			return v
		}
	}
	return []string{}
}
