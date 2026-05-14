package div

import (
	"fmt"
	"testing"
)

func TestDiv_Build(t *testing.T) {
	c := NewDiv()
	d := NewDiv()
	e := d.WithId("gaojian").WithClass("one", "two").
		WithClass("one", "three").WithId("guyuchun").
		WithStyle(map[string]string{
			"width":  "100px",
			"height": "200px",
		}).WithStyle(map[string]string{
		"width": "10%",
	}).WithDiv(c.WithId("abc")).WithDiv(c.WithId("abc")).Build()
	fmt.Println(e)
}

func TestDiv_Withxxx(t *testing.T) {
	// d := NewDiv()
}
