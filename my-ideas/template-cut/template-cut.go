// Package templatecut html模板切法文件
// 最终形状为唯一判别
// todo: 这里像二叉树的存储,可以使用slice存一遍leaf节点的的Block指针好获得所有leaf.可以哈希加链表,以后的事
package templatecut

var (
	hCut = "h-cut"
	vCut = "v-cut"
	leaf = "leaf"
)

// 定义宽高比
type Template struct {
	Block                  *Block
	HeightVideoPersent_max float32
	HeightVideoPersent_min float32
	WidthVideoPersent_max  float32
	WidthVideoPersent_min  float32
	Blocks                 []Block // 存储所有的最终形状,切一次得去重一次
}

// 切法得在原来基础上继续切,所以得准备一个数组的前一个的所有切法,而所存的block不能相连

// 现在切法的形状并不能影响我们真正的最终形象
// 那个形象的印象得在放html时在影响,而且哈我们video的放置是一个个放的,并不能控制最终的输出,有的过小有的过大
// 控制标准我无法掌控一般,我只能给个最大限制,主要看最终生成的像素点跟所给的像素点

type Block struct {
	Type   string // 横切竖切还是leaf
	Width  float32
	Height float32
	Left   *Block
	Right  *Block
}

// 这里规定宽高比
// 这里就类似形成
// 宽视频的最大2,最小1
// 高视频的最大1,最小1/2
func NewTemplate(width, height, heightVideoPersent_max, decay float32) *Template {
	return &Template{
		Block: &Block{
			Type:   leaf,
			Width:  width,
			Height: height,
		},
		HeightVideoPersent_max: heightVideoPersent_max,
		HeightVideoPersent_min: heightVideoPersent_max * decay,
		WidthVideoPersent_max:  heightVideoPersent_max * decay,
		WidthVideoPersent_min:  heightVideoPersent_max * decay * decay,
	}
}

func (b *Block) RandomCut() {

}

// // 而且每切一次所有的宽高比是不是得适应一下

// func (t *Template) Hcut() bool {
// 	// 获得所有可以切的block
// 	blocks := t.Block.gerLeafBlocks()
// 	// 我这不是要获得切法吗,判断在于最终形状

// 	// 检查最大比

// }

// func (b *Template) Vcut() bool {

// }

// func (b *Block) randomChoose() *Block {

// }

// func (b *Block) gerLeafBlocks() []*Block {
// 	if b.Isleaf() {
// 		return []*Block{b}
// 	}
// 	leafs := []*Block{}
// 	if b.Left != nil && b.Right != nil {
// 		leafs = append(leafs, b.Left.gerLeafBlocks()...)
// 		leafs = append(leafs, b.Right.gerLeafBlocks()...)
// 	}
// 	return leafs
// }

// // 如果是nil指针的话,直接报错
// func (b *Block) Isleaf() bool {
// 	return b.Type == leaf
// }

// func (b *Block) hcut() {

// }

// func (b *Block) vcut() {

// }
