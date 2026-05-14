// Package leaf 切法生产
package leaf

import (
	"fmt"
)

// 切法节点
type CutNode struct {
	Type  string   // H 横切, V 竖切
	Left  *CutNode // 切后的左边/上边部分
	Right *CutNode // 切后的右边/下边部分
}

// 切法记录（带统计）
type CutRecord struct {
	Tree *CutNode
	HCnt int      // 横切次数
	VCnt int      // 竖切次数
	Seq  []string // 切法序列（前序遍历）
}

func (node *CutNode) IsLeaf() bool {
	return node.Type == leaf
}

// 递归生成所有切法，直到切成 n 块
func GenerateCuts(blocks int) []*CutRecord {
	// 一块
	if blocks == 1 {
		return []*CutRecord{{
			Tree: &CutNode{Type: leaf},
			HCnt: 0,
			VCnt: 0,
			Seq:  []string{},
		}}
	}

	var results []*CutRecord
	// 左子树分出 k 块，右子树分出 blocks-k 块
	for k := 1; k < blocks; k++ {
		// 这里分的横切跟竖切 是当前节点的事 疯狂便利子cut 拿出子cut的所有情况append当前result
		// 横切
		for _, left := range GenerateCuts(k) { // 这里返回了一个slice 它便利的也是slice
			// 获得一个 leftnode
			for _, right := range GenerateCuts(blocks - k) {
				// 遍历获得 获得slice一个

				// 创建一个横切的块
				tree := &CutNode{
					Type:  H,
					Left:  left.Tree,
					Right: right.Tree,
				}
				// 左右切法合并加上 当前切法
				seq := append([]string{H}, append(left.Seq, right.Seq...)...)
				results = append(results, &CutRecord{
					Tree: tree,
					HCnt: left.HCnt + right.HCnt + 1,
					VCnt: left.VCnt + right.VCnt,
					Seq:  seq,
				})
			}
		}

		// 竖切
		for _, left := range GenerateCuts(k) {
			for _, right := range GenerateCuts(blocks - k) {
				tree := &CutNode{
					Type:  V,
					Left:  left.Tree,
					Right: right.Tree,
				}
				seq := append([]string{V}, append(left.Seq, right.Seq...)...)
				results = append(results, &CutRecord{
					Tree: tree,
					HCnt: left.HCnt + right.HCnt,
					VCnt: left.VCnt + right.VCnt + 1,
					Seq:  seq,
				})
			}
		}
	}
	return results
}

// 将树结构转为字符串（用于去重：相同的切法结构算一种）
func TreeToString(node *CutNode) string {
	if node.Type == leaf {
		return leaf
	}
	return fmt.Sprintf("%s(%s,%s)", node.Type, TreeToString(node.Left), TreeToString(node.Right))
}

func (node *CutNode) GetLeaves() []*CutNode {
	if node.Type == leaf {
		return []*CutNode{node}
	}
	return append(node.Left.GetLeaves(), node.Right.GetLeaves()...)
}

func (node *CutNode) GetLeavesCount() int {
	if node.Type == leaf {
		return 1
	}
	return node.Left.GetLeavesCount() + node.Right.GetLeavesCount()
}

func treeToMap(node *CutNode) map[string]interface{} {
	aMap := make(map[string]interface{})
	if node.Type == leaf {

	}
	return aMap
}
