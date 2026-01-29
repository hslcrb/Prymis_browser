package layout

import "prymis/engine/dom"

type Rect struct {
	X, Y, Width, Height float32
}

type EdgeSizes struct {
	Top, Bottom, Left, Right float32
}

type Dimensions struct {
	Content Rect
	Padding EdgeSizes
	Border  EdgeSizes
	Margin  EdgeSizes
}

type LayoutBox struct {
	Dimensions Dimensions
	BoxType    BoxType
	StyledNode *StyledNode
	Children   []*LayoutBox
}

type BoxType int

const (
	BlockNode BoxType = iota
	InlineNode
	AnonymousBlock
)

func NewLayoutTree(node *StyledNode) *LayoutBox {
	root := &LayoutBox{
		StyledNode: node,
	}
	if node.SpecifiedValues["display"] == "inline" {
		root.BoxType = InlineNode
	} else {
		root.BoxType = BlockNode
	}

	for _, child := range node.Children {
		if child.SpecifiedValues["display"] == "none" {
			continue
		}
		root.Children = append(root.Children, NewLayoutTree(child))
	}
	return root
}

func (b *LayoutBox) Layout(containerDimensions Dimensions) {
	switch b.BoxType {
	case BlockNode:
		b.layoutBlock(containerDimensions)
	case InlineNode:
		// Primitive: treat as block for move
		b.layoutBlock(containerDimensions)
	}
}

func (b *LayoutBox) layoutBlock(container Dimensions) {
	// Child's width = container's width - margins/paddings/borders
	// For simplicity, just use container width
	b.Dimensions.Content.Width = container.Content.Width
	b.Dimensions.Content.X = container.Content.X
	b.Dimensions.Content.Y = container.Content.Y + container.Content.Height

	var totalHeight float32
	for _, child := range b.Children {
		child.Layout(b.Dimensions)
		b.Dimensions.Content.Height += child.Dimensions.Content.Height
		totalHeight += child.Dimensions.Content.Height
	}

	// If no children, give some default height if it's a div or something
	if len(b.Children) == 0 && b.StyledNode.Node.NodeType == dom.ElementNode {
		b.Dimensions.Content.Height = 20
	}
}
