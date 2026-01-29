package layout

import (
	"prymis/engine/dom"
	"prymis/engine/parser"
)

type StyledNode struct {
	Node             *dom.Node
	SpecifiedValues  map[string]string
	Children         []*StyledNode
}

func NewStyledNode(node *dom.Node, rules []parser.StyleRule) *StyledNode {
	values := make(map[string]string)
	for _, rule := range rules {
		for _, selector := range rule.Selectors {
			if matches(node, selector) {
				for _, decl := range rule.Declarations {
					values[decl.Name] = decl.Value
				}
			}
		}
	}

	var children []*StyledNode
	for _, child := range node.Children {
		children = append(children, NewStyledNode(child, rules))
	}

	return &StyledNode{
		Node:            node,
		SpecifiedValues: values,
		Children:        children,
	}
}

func matches(n *dom.Node, selector string) bool {
	if n.NodeType != dom.ElementNode {
		return false
	}
	if selector == n.TagName {
		return true
	}
	if selector[0] == '.' && n.Attributes["class"] == selector[1:] {
		return true
	}
	if selector[0] == '#' && n.Attributes["id"] == selector[1:] {
		return true
	}
	return false
}
