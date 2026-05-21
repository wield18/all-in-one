// Package entity 定义一些基本的网络传输数据结构
package entity

// Window 窗口参数获得
type Window struct {
	Height float64 `form:"height"`
	Width  float64 `form:"width"`
}

// 视频索引
type VideoIndex struct {
	Index      string `form:"index"`
	Resolution string `form:"resolution"`
}

// 视频索引对应切片
type VideoSlice struct {
	Index      string `form:"index"`
	Resolution string `form:"resolution"`
	Time       string `form:"time"`
}
