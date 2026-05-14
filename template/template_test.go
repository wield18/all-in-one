package template

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/wield18/all-in-one/template/leaf"
	"github.com/wield18/all-in-one/template/math"

	// "github.com/wield18/all-in-one/template/math"
	"github.com/wield18/all-in-one/template/video"
)

// 这个不能用了
func TestRun(t *testing.T) {
	hVideoRoot := `C:/Games/Videos/The Other/Diives_Videos/ani/H`
	vVideoRoot := `C:/Games/Videos/The Other/Diives_Videos/ani/V`
	count := 5
	aMapCutNode := leaf.GetMapStringSliceCutNodeVCut(count)

	// 默认VCut必然一个
	var hVideoCount int
	var vVideoCount int
	var cutNode *leaf.CutNode
	for {
		vVideoCount = rand.Intn(count-2) + 2
		hVideoCount = count - vVideoCount

		cutNodeKey := leaf.FormatIntsToStats(hVideoCount, vVideoCount)
		// 真正选择的node
		if len(aMapCutNode[cutNodeKey]) == 1 {
			cutNode = aMapCutNode[cutNodeKey][0]
			break
		} else if len(aMapCutNode[cutNodeKey]) > 1 {
			cutNode = aMapCutNode[cutNodeKey][rand.Intn(len(aMapCutNode[cutNodeKey])-1)]
			break
		} else {
			// 这可能真找不到就在于我们第一切一定是竖切的原因 所以for一下吧直到找到
			fmt.Println("这基本可能发生")
			continue
		}
	}

	// aMap是真正要显示的video的info
	aMap := make(map[string]*[]video.FileInfo)
	// 这里的没对于所有的count做检验
	var hVideoInfos []video.FileInfo
	var vVideoInfos []video.FileInfo
	var err error
	hVideo, _ := video.GetAllFileInfo(hVideoRoot)
	if hVideoCount != 0 {
		hVideoInfos, err = video.GetRandomVideoFromSlice(hVideo[H], hVideoCount)
		if err != nil {
			fmt.Println(err)
			return
		}

		aMap["H"] = &hVideoInfos
	}
	vVideo, _ := video.GetAllFileInfo(vVideoRoot)
	if vVideoCount != 0 {
		vVideoInfos, err = video.GetRandomVideoFromSlice(vVideo[V], vVideoCount)
		if err != nil {
			fmt.Println(err)
			return
		}
		aMap["V"] = &vVideoInfos
	}
	fmt.Println("-------------------------------")
	fmt.Println(leaf.TreeToString(cutNode))
	maxW, maxH := 1920.0, 1000.0
	template := NewTemplate(cutNode, aMap, maxH, maxW)

	err = template.Node.OptimizeLayout()
	if err != nil {
		fmt.Println(err)
		return
	}
	leaves := template.Node.GetLeaves()
	for _, leaf := range leaves {
		fmt.Println("-------------------")
		fmt.Println(leaf.FilePath)
		fmt.Println(leaf.Scale)
		fmt.Println(leaf.Height, leaf.Width)
	}
	aMapFilepathWH := math.LeavesToMap(leaves)

	videoDiv := template.Div.GetVideoDiv()
	for _, div := range videoDiv {
		if oneWH, ok := aMapFilepathWH[div.FilePath]; ok {
			div.WithHeight(oneWH[0]).WithWidth(oneWH[1])
		}
	}
	str := template.Div.Build()
	fmt.Println(str)

}

func TestMap(t *testing.T) {
	aMap := map[string]string{
		"1": "1",
		"2": "1",
		"3": "1",
	}

	deleteMap(aMap, "1")
	fmt.Println(aMap)
}

func deleteMap(aMap map[string]string, attribute string) {
	delete(aMap, attribute)
}

func TestInitTempalte(t *testing.T) {
	// 现在cutNodeMap生成不对 不是链式
	// 1H4V
	// V(V(leaf,leaf),H(leaf,V(leaf,leaf)))
	InitTemplate(3, 5, "C:/Games/Videos/The Other/Diives_Videos/ani/H")
	for key, v := range aMapCutNode {
		fmt.Println(key)
		for _, cutNode := range v {
			fmt.Println(leaf.TreeToString(cutNode))
		}
	}

	for i := 0; i < 100; i++ {
		height, width := 1920.0, 1080.0
		str := GenerateVideoTemplate(height, width)
		fmt.Println(str)
	}
	fmt.Println(aMapCutNode, videoInfos_map)

}
