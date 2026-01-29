package dom

type NodeType int

const (
	ElementNode NodeType = iota
	TextNode
)

type AttrMap map[string]string

type Node struct {
	Children   []*Node
	NodeType   NodeType
	// Element data
	TagName    string
	Attributes AttrMap
	// Text data
	Text       string
}

func Text(data string) *Node {
	return &Node{NodeType: TextNode, Text: data}
}

func Element(name string, attrs AttrMap, children []*Node) *Node {
	return &Node{
		TagName:    name,
		Attributes: attrs,
		Children:   children,
		NodeType:   ElementNode,
	}
}
