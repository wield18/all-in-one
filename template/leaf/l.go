// Package leaf 主要是创建所有切法 具体见tree.go 这里主要生成统计数据记录供其他包选择切法
package leaf

import (
	"fmt"
	"strings"

	"github.com/wield18/all-in-one/template/constant"
)

var (
	H    = constant.H
	V    = constant.V
	leaf = constant.Leaf
)

// 如果这个方法得传入node，isChainV
// isChainV 可以用 node.Type == V 来第一次传入
// 竖切链来作为是否是竖屏视频
func CountLeafParentsVCut(node *CutNode, isChainV bool) map[string]int {
	// 这只对第一次输入时不正常传入的检查
	if node == nil || node.Type == leaf {
		return map[string]int{}
	}

	// 当前节点是否是 leaf 的直接父节点
	result := make(map[string]int)
	// 当前是不是还是V包裹V
	isChainV = node.Type == V && isChainV

	// 这里不是保守这里可能俩个节点不一样
	leftIsLeaf := node.Left != nil && node.Left.Type == leaf // 是leaf时
	rightIsLeaf := node.Right != nil && node.Right.Type == leaf

	if leftIsLeaf {
		if isChainV {
			result[V]++ // 这里增加的直接leaf节点
		} else {
			result[H]++ // 这里增加的直接leaf节点
		}
	}
	if rightIsLeaf {
		if isChainV {
			result[V]++ // 这里增加的直接leaf节点
		} else {
			result[H]++ // 这里增加的直接leaf节点
		}
	}

	// 递归处理子节点（即使子节点是 leaf 也不用递归了）
	if node.Left != nil && node.Left.Type != leaf {
		leftResult := CountLeafParentsVCut(node.Left, isChainV)
		for k, v := range leftResult {
			result[k] += v
		}
	}
	if node.Right != nil && node.Right.Type != leaf {
		rightResult := CountLeafParentsVCut(node.Right, isChainV)
		for k, v := range rightResult {
			result[k] += v
		}
	}

	return result
}

// 生成统计字符串， 类似 "0V2H"
func FormatStats(stats map[string]int) string {
	if len(stats) == 0 {
		return "0"
	}

	// 按 H 在前，V 在后排序输出
	parts := []string{}
	if h, ok := stats[H]; ok {
		parts = append(parts, fmt.Sprintf("%dH", h))
	} else {
		parts = append(parts, fmt.Sprintf("%dH", 0))
	}
	if v, ok := stats[V]; ok {
		parts = append(parts, fmt.Sprintf("%dV", v))
	} else {
		parts = append(parts, fmt.Sprintf("%dV", 0))
	}
	return strings.Join(parts, "")
}

// 把高度跟宽度写成string
func FormatIntsToStats(h, v int) string {
	if v == 0 && h == 0 {
		return ""
	}
	// 按 H 在前，V 在后排序输出
	parts := []string{}
	parts = append(parts, fmt.Sprintf("%dH", h))
	parts = append(parts, fmt.Sprintf("%dV", v))
	return strings.Join(parts, "")
}

// blocks: 总共切几块
// 竖切链来作为是否是竖屏视频 获得开局连续俩刀竖切的
func GetMapStringSliceCutNodeVCut(blocks int) map[string][]*CutNode {
	records := GenerateCuts(blocks)

	// 按 "横切数x竖切数" 分组，同时去除结构重复的树
	groupMap := make(map[string]map[string]*CutRecord) // key: "HxV" -> 树签名 -> 记录
	// 这样计算的有个别重的 但输出满足需求只是美中不足
	for _, rec := range records {
		key := fmt.Sprintf("%d横%d竖", rec.HCnt, rec.VCnt)
		if groupMap[key] == nil {
			groupMap[key] = make(map[string]*CutRecord)
		}
		// sig去重了
		sig := TreeToString(rec.Tree)
		// 只保留不同的树结构
		if _, exists := groupMap[key][sig]; !exists {

			groupMap[key][sig] = rec
		}
	}

	final := make(map[string][]*CutNode)

	// // 输出结果
	for _, trees := range groupMap {
		i := 1
		for _, rec := range trees {
			// 这里对于 rec.Tree 第一切为横切的全部舍弃
			if rec.Tree.Type == H {
				continue
			} else if rec.Tree.Type == V {
				// 这里第一切必然是竖切 V
				// 这里只有有开头连续俩刀竖切才做记录
				if rec.Tree.Left.Type == V || rec.Tree.Right.Type == V {
					stats := CountLeafParentsVCut(rec.Tree, true)
					final[FormatStats(stats)] = append(final[FormatStats(stats)], rec.Tree)
					i++
				}
			} else {
				fmt.Println("type wrong: ", rec.Tree.Type)
			}

		}
	}
	return final
}
