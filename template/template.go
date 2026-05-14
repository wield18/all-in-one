// Package template 整体调用组合的地方
package template

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/wield18/all-in-one/template/div"
	"github.com/wield18/all-in-one/template/leaf"
	"github.com/wield18/all-in-one/template/math"
	"github.com/wield18/all-in-one/template/video"
)

var (
	V    = math.V
	H    = math.H
	Leaf = math.Leaf
)

// var (
// 	errDisMatchChildCount = errors.New("DisMatch Child Count: 孩子div数量与当前调用的方法不一致")
// )

// 这里直接组合
type Template struct {
	// 这俩其中一个
	Node *math.Node
	Div  *div.CustomDiv
}

func NewTemplate(cutNode *leaf.CutNode, aMapFileInfo map[string]*[]video.FileInfo, maxH, maxW float64) *Template {
	// 校验
	if cutNode.Type == Leaf {
		return nil
	}
	t := &Template{}
	t.Div, t.Node = ConvDivAndNodeVCut(cutNode, cutNode.Type == V, aMapFileInfo)
	t.Node.WithMaxHeightAndWidth(maxH, maxW)
	return t
}

// ConvDivAndNode 这里math跟div是强关联的必须同时进行
func ConvDivAndNode(cutNode *leaf.CutNode, parentCutType string, aMapFileInfo map[string]*[]video.FileInfo) (*div.CustomDiv, *math.Node) {
	switch cutNode.Type {
	case Leaf:
		var file video.FileInfo
		if parentCutType == H {
			file = (*aMapFileInfo[H])[0]
			*aMapFileInfo[H] = (*aMapFileInfo[H])[1:]
		} else {
			file = (*aMapFileInfo[V])[0]
			*aMapFileInfo[V] = (*aMapFileInfo[V])[1:]
		}
		return div.NewCustomDiv().WithFilePath(file.FilePath),
			// &math.Leaf{OrigW: float64(file.Width), OrigH: float64(file.Height), Scale: 1, FilePath: file.FilePath}
			math.NewNode().WithType(Leaf).WithFile(float64(file.Height), float64(file.Width), file.FilePath)
	case H:
		divLeft, nodeLeft := ConvDivAndNode(cutNode.Left, cutNode.Type, aMapFileInfo)
		divRight, nodeRight := ConvDivAndNode(cutNode.Right, cutNode.Type, aMapFileInfo)
		return div.NewCustomDiv().WithChild(divLeft,
			divRight), math.NewNode().WithType(H).WithLeft(nodeLeft).WithRight(nodeRight)
		// &math.HNode{Left: nodeLeft, Right: nodeRight}
	default:
		divLeft, nodeLeft := ConvDivAndNode(cutNode.Left, cutNode.Type, aMapFileInfo)
		divRight, nodeRight := ConvDivAndNode(cutNode.Right, cutNode.Type, aMapFileInfo)
		return div.NewCustomDiv().WithChild(divLeft,
			divRight).WithDisplayFlex(true), math.NewNode().WithType(V).WithLeft(nodeLeft).WithRight(nodeRight)
		// &math.HNode{Left: nodeLeft, Right: nodeRight}
	}

}

// VCut链
func ConvDivAndNodeVCut(cutNode *leaf.CutNode, isChainV bool, aMapFileInfo map[string]*[]video.FileInfo) (*div.CustomDiv, *math.Node) {
	switch cutNode.Type {
	case Leaf:
		var file video.FileInfo
		if !isChainV {
			file = (*aMapFileInfo[H])[0]
			*aMapFileInfo[H] = (*aMapFileInfo[H])[1:]
		} else {
			file = (*aMapFileInfo[V])[0]
			*aMapFileInfo[V] = (*aMapFileInfo[V])[1:]
		}
		return div.NewCustomDiv().WithFilePath(file.FilePath), math.NewNode().WithType(Leaf).WithFile(float64(file.Height), float64(file.Width), file.FilePath)
	case H:
		divLeft, nodeLeft := ConvDivAndNodeVCut(cutNode.Left, false, aMapFileInfo)
		divRight, nodeRight := ConvDivAndNodeVCut(cutNode.Right, false, aMapFileInfo)
		return div.NewCustomDiv().WithChild(divLeft,
			divRight), math.NewNode().WithType(H).WithLeft(nodeLeft).WithRight(nodeRight)
	default:
		divLeft, nodeLeft := ConvDivAndNodeVCut(cutNode.Left, cutNode.Type == V && isChainV, aMapFileInfo)
		divRight, nodeRight := ConvDivAndNodeVCut(cutNode.Right, cutNode.Type == V && isChainV, aMapFileInfo)
		return div.NewCustomDiv().WithChild(divLeft,
			divRight).WithDisplayFlex(true), math.NewNode().WithType(V).WithLeft(nodeLeft).WithRight(nodeRight)
	}

}

// div基本初始化
func TempDiv(cutNode *leaf.CutNode, parentCutType string, aMapFileInfo map[string]*[]video.FileInfo) *div.CustomDiv {
	// 当前的cutNode
	switch cutNode.Type {
	case Leaf:
		var file video.FileInfo
		if parentCutType == H {
			file = (*aMapFileInfo[H])[0]
			*aMapFileInfo[H] = (*aMapFileInfo[H])[1:]
		} else {
			file = (*aMapFileInfo[V])[0]
			*aMapFileInfo[V] = (*aMapFileInfo[V])[1:]
		}
		return div.NewCustomDiv().WithFilePath(file.FilePath)
	case H:
		return div.NewCustomDiv().WithChild(TempDiv(cutNode.Left, cutNode.Type, aMapFileInfo),
			TempDiv(cutNode.Right, cutNode.Type, aMapFileInfo))
	default:
		return div.NewCustomDiv().WithChild(TempDiv(cutNode.Left, cutNode.Type, aMapFileInfo),
			TempDiv(cutNode.Right, cutNode.Type, aMapFileInfo)).WithDisplayFlex(true)
	}

}

var (
	aMapCutNode    = make(map[string][]*leaf.CutNode)
	videoInfos_map = make(map[string][]video.FileInfo)
	videoKeys      = make([]string, 0) // 对应的video布局可能表
	MinCount       = 3
)

func InitTemplate(minCount, maxCount int, videoRoot string) {
	if minCount < MinCount {
		minCount = MinCount
	}
	if maxCount < minCount {
		maxCount = minCount
	}
	// videoInfos_map
	var err error
	videoInfos_map, err = video.GetAllFileInfo(videoRoot)
	if err != nil {
		panic(err)
	}
	// aMapCutNode
	for i := minCount; i <= maxCount; i++ {
		tempMap := leaf.GetMapStringSliceCutNodeVCut(i)
		// 去掉tempMap多余的 0H...V
		for key := range tempMap {
			// key是不会重的
			videoKeys = append(videoKeys, key)
			if strings.Contains(key, "0H") {
				// 还有 aMapCutNode 的对应 只加一个对于这样的
				aMapCutNode[key] = []*leaf.CutNode{tempMap[key][0]}
			} else {
				// 这里的videokey很重复 如果要去重也得看对应的aMapCutNode的分布来去
				aMapCutNode[key] = append(aMapCutNode[key], tempMap[key]...)
			}
		}

	}

	fmt.Println("tmeplate 初始化完成")
}

// 只这里至少三个, 跟InitTemplate相对应
func GenerateVideoTemplate(h, w float64) string {
	// 默认VCut必然连续俩个
	var cutNode *leaf.CutNode
	var cutNodeKey string
	for {
		// 这里哈 在key中随便选个
		cutNodeKey = videoKeys[rand.Intn(len(videoKeys)-1)]
		// 真正选择的node
		l := len(aMapCutNode[cutNodeKey])
		if l == 1 {
			cutNode = aMapCutNode[cutNodeKey][0]
			break
		} else if l > 1 {
			cutNode = aMapCutNode[cutNodeKey][rand.Intn(l-1)]
			break
		} else {
			// 这可能真找不到就在于我们第一切一定是竖切的原因 所以for一下吧直到找到
			fmt.Println("这基本可能发生")
			continue
		}
	}
	aMap := make(map[string]*[]video.FileInfo)
	var hInfos []video.FileInfo
	var vInfos []video.FileInfo
	hCount, _ := strconv.ParseInt(string(cutNodeKey[0]), 10, 64)
	vCount, _ := strconv.ParseInt(string(cutNodeKey[2]), 10, 64)
	hVideoCount := int(hCount)
	vVideoCount := int(vCount)
	var err error
	if hVideoCount != 0 {
		hInfos, err = video.GetRandomVideoFromSlice((videoInfos_map[H]), hVideoCount)
		if err != nil {
			fmt.Println(err)
			return "hInfos, err = video.GetRandomVideoFromSlice((videoInfos_map[H]), hVideoCount)"
		}

		aMap["H"] = &hInfos
	}
	if vVideoCount != 0 {
		vInfos, err = video.GetRandomVideoFromSlice((videoInfos_map[V]), vVideoCount)
		if err != nil {
			fmt.Println(err)
			return "vInfos, err = video.GetRandomVideoFromSlice((videoInfos_map[V]), vVideoCount)"
		}
		aMap["V"] = &vInfos
	}

	fmt.Println("-------------------------------")
	fmt.Println(leaf.TreeToString(cutNode))

	template := NewTemplate(cutNode, aMap, h, w)

	err = template.Node.OptimizeLayout()
	if err != nil {
		fmt.Println(err)
		return "err = template.Node.OptimizeLayout()"
	}
	leaves := template.Node.GetLeaves()
	for _, leaf := range leaves {
		fmt.Println("-------------------")
		fmt.Printf("FilePath:%s ,Scale:%.3f ,Height:%.3f ,Width:%.3f\n",
			leaf.FilePath, leaf.Scale, leaf.Height, leaf.Width)
	}
	aMapFilepathWH := math.LeavesToMap(leaves)

	videoDiv := template.Div.GetVideoDiv()
	for _, div := range videoDiv {
		if oneWH, ok := aMapFilepathWH[div.FilePath]; ok {
			div.WithHeight(oneWH[0]).WithWidth(oneWH[1])
		}
	}
	return template.Div.Build()

}
