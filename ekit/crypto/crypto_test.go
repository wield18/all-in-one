package crypto

import (
	"fmt"
	"testing"
)

func TestHashPassword(t *testing.T) {
	a, _ := HashPassword("abc")
	b, _ := HashPassword("abc")
	c, _ := HashPassword("abc")
	fmt.Println(CheckPassword("abc", a))
	fmt.Println(CheckPassword("abc", b))
	fmt.Println(CheckPassword("abc", c))
	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(c)
}
