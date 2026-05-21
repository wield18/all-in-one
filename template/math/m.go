// Package math 用来计算每个div的大小缩放
package math

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/wield18/all-in-one/template/constant"
)

// 缩放思路很简单看就看chains的
// 首先我们得分布一下块，获得chains
// 从上往下来最大优化，直接的leaf节点先满足，再满足直接的leaf节点跟孙子leaf节点
// 对于VNode的子节点的高度不能大于VNode的直接leaf，
// 子节点的俩个子节点得和子节点的兄弟节点差不多 V 对应 height H 对应 width
// HNode也是 这样丰富了限制
// 放大的时候得先仅上层leaf进行放大，尽量接近俩个最大，
// 例如 V(leaf1,H(leaf2,leaf3))
// 得先限制leaf1跟leaf2的width还有leaf1的高度，
// 这俩优先级当热是直接leaf1节点的高度先满足
// 竖切高度一致优先，横切宽度一致优先

// 算法有缺陷可能会导致外层很大,但内层很小
// 这里使用leaves的数量做为空间的分配

var (
	Leaf = constant.Leaf
	V    = constant.V
	H    = constant.H
	base = 0.06 // 空间基数
)

var (
	errUnSupportedType       = errors.New("UnSupported Type: 不支持type")
	errLackOfFirstConstraint = errors.New("Lack Of First Constraint: 没有第一约束条件")
)

// 这里的node最外层必传 MaxHeight MaxWidth Type
type Node struct {
	Type      string // "V" "H" "leaf"
	Left      *Node
	Right     *Node
	MaxHeight float64
	MaxWidth  float64 // 当前最大
	Height    float64
	Width     float64 // 如果是叶子节点一定有height width Scale
	Scale     float64
	FilePath  string // 如果是leaf的话一定得有filepath
}

func (n *Node) IsLeaf() bool { return true }

func (n *Node) GetLeaves() []*Node {
	if n.Type == Leaf {
		return []*Node{n}
	}
	return append(n.Left.GetLeaves(), n.Right.GetLeaves()...)
}

func (n *Node) GetLeavesCount() int {
	if n.Type == Leaf {
		return 1
	}
	return n.Left.GetLeavesCount() + n.Right.GetLeavesCount()
}

func (n *Node) OptimizeLayout() error {
	if n.Type == Leaf {
		return fmt.Errorf("%w:%s", errUnSupportedType, n.Type)
	}
	// 横切竖切的第一约束没有直接return
	if n.MaxHeight == 0 || n.MaxWidth == 0 {
		return fmt.Errorf("%w:%.3f\n%w:%.3f", errLackOfFirstConstraint, n.MaxHeight,
			errLackOfFirstConstraint, n.MaxWidth)
	}
	switch n.Type {
	case V:

		// 如果是竖切，那子节点如果是leaf的话，直接优化子节点
		// 俩个必须同时优化
		left, right := n.Left, n.Right
		// 这里有四种情况
		// 左leaf右node
		// 左node右leaf
		// 左右都node
		// 左右都leaf
		if left.Type == Leaf {
			// 左右都leaf
			if right.Type == Leaf {
				return n.HandleTwoLeafV()
			} else { // // 左leaf右node
				return n.HandleLeftLeafV(left, right)
			}
		} else {
			// 左node右leaf
			if right.Type == Leaf {
				// 左右调换一下
				left, right = right, left
				return n.HandleLeftLeafV(left, right)
			} else { // 左右都node
				// 我想这随便选一个最大化node
				if index := rand.Intn(2); index == 0 {
					return n.HandleTwoNodeV(left, right)
				} else {
					left, right = right, left
					return n.HandleTwoNodeV(left, right)
				}
			}
		}
	case H:
		left, right := n.Left, n.Right
		if left.Type == Leaf {
			// 左右都leaf
			if right.Type == Leaf {
				return n.HandleTwoLeafH()
			} else { // // 左leaf右node
				return n.HandleLeftLeafH(left, right)
			}
		} else {
			// 左node右leaf
			if right.Type == Leaf {
				// 左右调换一下
				left, right = right, left
				return n.HandleLeftLeafH(left, right)
			} else { // 左右都node
				// 我想这随便选一个最大化node
				if index := rand.Intn(2); index == 0 {
					return n.HandleTwoNodeH(left, right)
				} else {
					left, right = right, left
					return n.HandleTwoNodeH(left, right)
				}
			}
		}
	default:
		return fmt.Errorf("%w:%s", errUnSupportedType, n.Type)
	}
}

func (n *Node) HandleTwoLeafV() error {
	left, right := n.Left, n.Right
	left.Scale = n.MaxHeight / left.Height
	right.Scale = n.MaxHeight / right.Height
	width := left.Width*left.Scale + right.Width*right.Scale
	if width > n.MaxWidth {
		left.Scale, right.Scale = left.Scale*n.MaxWidth/width, right.Scale*n.MaxWidth/width
	}
	if left.Height > right.Height {
		n.MaxHeight = left.Height * left.Scale
	}
	return nil
}

func (n *Node) HandleTwoLeafH() error {
	left, right := n.Left, n.Right
	left.Scale = n.MaxWidth / left.Width
	right.Scale = n.MaxWidth / right.Width
	height := left.Height*left.Scale + right.Height*right.Scale
	if height > n.MaxHeight {
		left.Scale, right.Scale = left.Scale*n.MaxHeight/height, right.Scale*n.MaxHeight/height
	}
	if left.Width > right.Width {
		n.MaxWidth = left.Width * left.Scale
	}
	return nil
}

func (n *Node) HandleLeftLeafV(left, right *Node) error {
	// 右节点的叶子数
	rightCount := right.GetLeavesCount()

	left.Scale = n.MaxHeight / left.Height
	// 左叶子最大有1/rightCount + Base
	// 这不能根据count来决定scale吧
	maxScale := 1.0/(float64(rightCount)+1.0) + base
	if left.Scale > maxScale {
		left.Scale = maxScale
	}
	right.MaxWidth = n.MaxWidth - left.Scale*left.Width // 剩下的都给子节点
	right.MaxHeight = n.MaxHeight
	err := right.OptimizeLayout()
	if err != nil {
		return err
	}
	// 在右边先优化完
	// 获得右边的宽度
	// 来获得左边的最大显示
	if n.MaxWidth-right.MaxWidth > left.Scale*left.Width { // 有剩余空间,先占满最大宽度,高度判断
		left.Scale = (n.MaxWidth - right.MaxWidth) / left.Width // 这是占满
	}
	if left.Height*left.Scale > n.MaxHeight {
		left.Scale = n.MaxHeight / left.Height // 再满足最大高度限制
	}
	return nil
}

func (n *Node) HandleLeftLeafH(left, right *Node) error {
	rightCount := right.GetLeavesCount()
	maxScale := 1.0/(float64(rightCount)+1.0) + base
	left.Scale = n.MaxWidth / left.Width // scale 是当前视频得乘以多少才能满足当前最大的宽度或高度
	if left.Scale > maxScale {
		left.Scale = maxScale
	}
	right.MaxHeight = n.MaxHeight - left.Scale*left.Height // 剩下的都给子节点
	right.MaxWidth = n.MaxWidth
	err := right.OptimizeLayout()
	if err != nil {
		return err
	}
	// 高度限制优先
	// if n.MaxWidth-right.MaxWidth > left.Scale*left.Width { // 有剩余空间,先占满最大宽度,高度判断
	// 	left.Scale = (n.MaxWidth - right.MaxWidth) / left.Width // 这是占满
	// }
	// if left.Height*left.Scale > n.MaxHeight {
	// 	left.Scale = n.MaxHeight / left.Height // 再满足最大高度限制
	// }
	if n.MaxHeight-right.Height > left.Scale*left.Width {
		left.Scale = (n.MaxHeight - right.MaxHeight) / left.Height
	}
	if left.Width*left.Scale > n.MaxWidth {
		left.Scale = n.MaxWidth / left.Width
	}
	return nil
}

func (n *Node) HandleTwoNodeV(left, right *Node) error {
	leftCount := left.GetLeavesCount()
	rightCount := right.GetLeavesCount()
	maxScale := float64(leftCount)/(float64(rightCount)+float64(leftCount)) + base
	right.MaxHeight, left.MaxHeight = n.MaxHeight, n.MaxHeight
	left.MaxWidth = n.MaxWidth * maxScale
	err := left.OptimizeLayout()
	if err != nil {
		return err
	}
	// 等等 这里不嫩直接拿到leaf的scale因为它不是leaf
	// 想个办法拿当前所有宽度链的最大的那个，获得其中的宽度就行, 然后对其leaf整体运用宽度的缩放
	// 留下空间给他的兄弟node

	// 然后这里必然不能超过
	right.MaxWidth = n.MaxWidth - left.Width*left.Scale
	return right.OptimizeLayout()
}
func (n *Node) HandleTwoNodeH(left, right *Node) error {
	leftCount := left.GetLeavesCount()
	rightCount := right.GetLeavesCount()
	maxScale := float64(leftCount)/(float64(rightCount)+float64(leftCount)) + base
	right.MaxWidth, left.MaxWidth = n.MaxWidth, n.MaxWidth
	left.MaxHeight = n.MaxHeight * maxScale
	err := left.OptimizeLayout()
	if err != nil {
		return err
	}
	right.MaxHeight = n.MaxHeight - left.Height*left.Scale
	return right.OptimizeLayout()
}

// filepath:[height, width]
func LeavesToMap(n []*Node) map[string][]int {
	aMap := make(map[string][]int)
	for _, node := range n {
		aMap[node.FilePath] = []int{int(node.Height * node.Scale), int(node.Width * node.Scale)}
	}
	return aMap
}
