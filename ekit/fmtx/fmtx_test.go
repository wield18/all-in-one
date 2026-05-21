package fmtx

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSprintf(t *testing.T) {
	a := Sprintf("gaojian", "token")
	fmt.Println(a)
	fmt.Println(len(a))

}

func TestSprintfInterfaces(t *testing.T) {
	a := SprintfInterfaces("gaojian", 123, '1') // gaojian: 123: 49 len(16)
	fmt.Println(a)
	fmt.Println(len(a))
}

// 可以直接string跟任何类型比较
func TestStringCompare(t *testing.T) {
	str := "qwe"
	aSlice := []interface{}{123, str}
	if aSlice[0] == str {
		fmt.Println("good")
	}
}

func TestStringCompareS(t *testing.T) {
	Sprintf()
}

func TestSprintf_1(t *testing.T) {
	val := "one"
	testCases := []struct {
		name     string
		wantVal  string
		testStrs []string
	}{
		{
			name:     "test-1-empty",
			wantVal:  "",
			testStrs: make([]string, 0),
		},
		{
			name:     "test-one",
			wantVal:  fmt.Sprintf("%s", val),
			testStrs: []string{val},
		},
		{
			name:     "test-two",
			wantVal:  fmt.Sprintf("%s: %s", val, val),
			testStrs: []string{val, val},
		},
		{
			name:     "test-three",
			wantVal:  fmt.Sprintf("%s: %s: %s", val, val, val),
			testStrs: []string{val, val, val},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, Sprintf(tc.testStrs...), tc.wantVal)
		})
	}
}
