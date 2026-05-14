// Package div 使用来真正生产html模板的地方
// custom-div.go 只是 div.go 的一个简单封装使用
package div

import (
	"fmt"
	"strconv"
)

// CustomDiv 个性化Div
// 最外层得定义height
// 里边的div使用也使用绝对的,这里的绝对使用里面video的缩放后对应的绝对值
// 只有最里边的video使用百分比
type CustomDiv struct {
	Display       bool
	Width, Height int
	Childs        map[*CustomDiv]bool
	D             *Div
	FilePath      string
}

func NewCustomDiv() *CustomDiv {
	return &CustomDiv{
		Childs: make(map[*CustomDiv]bool),
		D:      NewDiv(),
	}
}

func (cd *CustomDiv) WithDisplayFlex(isflex bool) *CustomDiv {
	if isflex {
		cd.Display = true
	}
	return cd
}

func (cd *CustomDiv) WithWidth(val int) *CustomDiv {
	cd.Width = val
	return cd
}

func (cd *CustomDiv) WithHeight(val int) *CustomDiv {
	cd.Height = val
	return cd
}

func (cd *CustomDiv) WithFilePath(val string) *CustomDiv {
	cd.FilePath = val
	return cd
}

func (cd *CustomDiv) WithChild(childs ...*CustomDiv) *CustomDiv {
	for _, child := range childs {
		if _, ok := cd.Childs[child]; !ok {
			cd.Childs[child] = false
		}
	}
	return cd
}

func (cd *CustomDiv) WithOuterDivSepcialFlex() *CustomDiv {
	cd.D.WithStyle(outerStyle)
	return cd
}

func (cd *CustomDiv) GetVideoDiv() []*CustomDiv {
	// 这就说明它下面就是video
	if cd.FilePath != "" {
		return []*CustomDiv{cd}
	}
	res := make([]*CustomDiv, 0)
	for child := range cd.Childs {
		res = append(res, child.GetVideoDiv()...)
	}

	return res
}

// 啥child都没有的话直接build会认为是video的包裹div
func (cd *CustomDiv) Build() string {
	aMap := make(map[string]string)
	if cd.Display {
		aMap = flexStyle
	}
	if cd.Height != 0 {
		aMap["height"] = strconv.FormatInt(int64(cd.Height), 10) + "px"
	}
	if cd.Width != 0 {
		aMap["width"] = strconv.FormatInt(int64(cd.Width), 10) + "px"
	}
	innerStr := ""
	if len(cd.Childs) != 0 {
		for child := range cd.Childs {
			innerStr += child.Build()
		}
	} else if cd.FilePath != "" {
		innerStr = fmt.Sprintf(videoHtml, cd.FilePath)
	} else {
		innerStr = deFaultVideoHtml
	}
	// 现在得把 style 跟 videoHtml 注入 cd.d 里
	return cd.D.WithInnerHtml(innerStr).WithStyle(aMap).Build()
}
