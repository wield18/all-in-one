package math

import (
	"fmt"
	"testing"
)

func TestMain(t *testing.T) {

	w, h := 1920.0, 1580.0
	file1 := NewNode().WithType(Leaf).WithFile(1920, 1008, "zxc1") // 竖
	file2 := NewNode().WithType(Leaf).WithFile(1080, 1920, "zxc2") // 横视频
	file3 := NewNode().WithType(Leaf).WithFile(520, 720, "qwe3")
	file4 := NewNode().WithType(Leaf).WithFile(640, 660, "asd4")
	// 构造布局
	node := NewNode().WithType(V).WithMaxHeightAndWidth(h, w).
		WithLeft(file1).WithRight(
		NewNode().WithType(H).WithLeft(file2).WithRight(
			NewNode().WithType(V).WithLeft(file3).WithRight(file4)))
	err := node.OptimizeLayout()
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, leaf := range node.GetLeaves() {
		fmt.Println("-------------------")
		fmt.Println(leaf.FilePath)
		fmt.Println(leaf.Scale)
		fmt.Println(leaf.Height, leaf.Width)
	}
	// fmt.Printf("开始优化，目标尺寸: %.0f x %.0f\n", maxW, maxH)
	// fmt.Println("初始布局:")
	// w, h, _, _ := root.ComputeWithSensitivity()
	// fmt.Printf("  尺寸: %.0f x %.0f\n", w, h)

	// success := OptimizeLayout(root, maxW, maxH, maxIter)

	// fmt.Printf("\n优化结果: %v\n", success)
	// w, h, _, _ = root.ComputeWithSensitivity()
	// fmt.Printf("最终布局尺寸: %.0f x %.0f\n", w, h)
	// fmt.Printf("叶子缩放:\n")
	// fmt.Printf("  视频1 W%.3f H%.3f: %.3f\n", leaf1.OrigW, leaf1.OrigH, leaf1.Scale)
	// fmt.Printf("  视频2 W%.3f H%.3f: %.3f\n", leaf2.OrigW, leaf2.OrigH, leaf2.Scale)
	// fmt.Printf("  视频3 W%.3f H%.3f: %.3f\n", leaf3.OrigW, leaf3.OrigH, leaf3.Scale)
	// fmt.Printf("  视频4 W%.3f H%.3f: %.3f\n", leaf4.OrigW, leaf4.OrigH, leaf4.Scale)

	// // 验证缩放是否在允许范围内
	// fmt.Println("\n约束检查:")
	// for i, leaf := range []*Leaf{leaf1, leaf2, leaf3, leaf4} {
	// 	minScale := 0.5
	// 	maxScale := 2.0
	// 	valid := leaf.Scale >= minScale && leaf.Scale <= maxScale
	// 	fmt.Printf("  视频%d: 缩放=%.3f, 范围[%.1f,%.1f], %v\n",
	// 		i+1, leaf.Scale, minScale, maxScale, valid)
	// }
	// 现在哈
	// 我进行我的链式选择
	// 我得判断当前节点是V还是H

}
