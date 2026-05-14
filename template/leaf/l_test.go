package leaf

import (
	"fmt"
	"testing"
)

func TestMain(t *testing.T) {
	blocks := 4
	aMap := GetMapStringSliceCutNodeVCut(blocks)
	for tag, trees := range aMap {
		fmt.Println(tag)
		for _, tree := range trees {
			fmt.Println(TreeToString(tree))
		}
	}
}
