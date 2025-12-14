package trans

//  ---  Interfaces

// -- Basic
type IMdhtmlLineNode interface {
	Block() string
}

// Node with transformations
type IMdhtmlTransfNode interface {
	IMdhtmlLineNode
	Transform(templDir string) error
	AddParamString(parVal string) error
	AddblockHtml(val string) error
	HasJsonBlock() bool
	JsonBlock() string
	JsonBlockType() string
}

// -- Basic, implements IMdhtmlLineNode
type MdhtLineNode struct {
	block       string
	before_link string
	after_link  string
}

func (n *MdhtLineNode) Block() string {
	return n.block
}

func NewMdhtLineNode(line string) *MdhtLineNode {
	ln := MdhtLineNode{block: line}
	return &ln
}
