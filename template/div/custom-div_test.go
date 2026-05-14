package div

import (
	"fmt"
	"testing"
)

func TestCustomDiv_Build(t *testing.T) {
	child := NewCustomDiv().WithChild(NewCustomDiv().WithWidth(100).WithHeight(200)).WithWidth(300)
	d := NewCustomDiv()
	fmt.Println(d.WithDisplayFlex(true).WithHeight(100).WithWidth(500).WithHeight(300).
		WithChild(child).Build())
}
