// 构建
package math

func NewNode() *Node {
	return &Node{}
}

func (n *Node) WithMaxHeightAndWidth(h, w float64) *Node {
	return n.WithMaxHeight(h).WithMaxWidth(w)
}

func (n *Node) WithHeightAndWidth(h, w float64) *Node {
	return n.WithHeight(h).WithWidth(w)
}

func (n *Node) WithFile(h, w float64, filepath string) *Node {
	return n.WithHeightAndWidth(h, w).WithFilePath(filepath)
}

func (n *Node) WithType(t string) *Node {
	n.Type = t
	return n
}

func (n *Node) WithLeft(aNode *Node) *Node {
	n.Left = aNode
	return n
}

func (n *Node) WithRight(aNode *Node) *Node {
	n.Right = aNode
	return n
}

func (n *Node) WithMaxHeight(h float64) *Node {
	n.MaxHeight = h
	return n
}

func (n *Node) WithMaxWidth(w float64) *Node {
	n.MaxWidth = w
	return n
}

func (n *Node) WithHeight(h float64) *Node {
	n.Height = h
	return n
}

func (n *Node) WithWidth(w float64) *Node {
	n.Width = w
	return n
}

func (n *Node) WithFilePath(path string) *Node {
	n.FilePath = path
	return n
}
