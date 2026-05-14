// Package div 使用来真正生产html模板的地方
package div

import (
	"fmt"
	"strings"
)

var (
	panicNilString               = "Nil String: 空字符串"
	panicStringMustHasNoSpace    = "String Must Has No Space: 字符串不可有空格"
	panicInnerHtmlAndDivsAllHave = "InnerHtml And Divs All Have: 内部的html跟Div不能都有"
)

// 基本div类型
type Div struct {
	id      string
	classes string
	styles  map[string]string
	divs    map[*Div]bool // 因为查重得转map 不如直接map
	html    string
}

func NewDiv() *Div {
	return &Div{
		styles: make(map[string]string),
		divs:   make(map[*Div]bool),
	}
}

func (d *Div) WithId(id string) *Div {
	if id != "" {
		if strings.Contains(id, " ") {
			panic(fmt.Sprintf("id: %s\n", panicStringMustHasNoSpace))
		}
		d.id = id
		return d
	}
	panic(panicNilString)
}

func (d *Div) WithClass(classes ...string) *Div {
	d.classes = d.classes + " "
	for _, oneClass := range classes {
		if !strings.Contains(d.classes, oneClass) {
			if strings.Contains(oneClass, " ") {
				panic(fmt.Sprintf("oneClass: %s\n", panicStringMustHasNoSpace))
			}
			d.classes += oneClass + " "
		}
	}
	d.classes = d.classes[:len(d.classes)-1]
	return d
}

func (d *Div) WithStyle(styles map[string]string) *Div {
	for oneStyle, value := range styles {
		if strings.Contains(oneStyle, " ") {
			panic(fmt.Sprintf("oneStyle: %s\n", panicStringMustHasNoSpace))
		}
		d.styles[oneStyle] = value
	}
	return d
}

func (d *Div) WithDiv(divs ...*Div) *Div {
	for _, div := range divs {
		if _, ok := d.divs[div]; !ok {
			d.divs[div] = false
		}
	}
	return d
}

func (d *Div) WithInnerHtml(html string) *Div {
	d.html = html
	return d
}

func (d *Div) Build() string {
	str := "<div "
	if d.id != "" {
		str += `id="` + d.id + `" `
	}
	if d.classes != "" {
		str += `class="` + d.classes + `" `
	}
	if len(d.styles) != 0 {
		str += `style="`
		for style, value := range d.styles {
			str += fmt.Sprintf("%s: %s; ", style, value)
		}
		str = str[:len(str)-1]
		str += `" `
	}
	str = str[:len(str)-1] + ">"
	// 如果俩个都有的话直接panic吧先
	if d.html != "" && len(d.divs) != 0 {
		panic(panicInnerHtmlAndDivsAllHave)
	}
	// 内部html直接插入 子div
	if d.html != "" {
		str += d.html
	}
	// 或者直接调用 子div的Build
	if len(d.divs) != 0 {
		for div := range d.divs {
			div_str := div.Build()
			str += div_str
		}
	}
	str += "</div>"
	return str
}
