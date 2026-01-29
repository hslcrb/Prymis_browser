package parser

import (
	"prymis/engine/dom"
	"strings"
)

type HTMLParser struct {
	pos   int
	input string
	depth int
}

func NewHTMLParser(input string) *HTMLParser {
	return &HTMLParser{input: input}
}

func (p *HTMLParser) Parse() *dom.Node {
	nodes := p.parseNodes()
	if len(nodes) == 1 {
		return nodes[0]
	}
	return dom.Element("html", dom.AttrMap{}, nodes)
}

func (p *HTMLParser) parseNodes() []*dom.Node {
	var nodes []*dom.Node
	for {
		p.consumeWhitespace()
		if p.eof() || strings.HasPrefix(p.input[p.pos:], "</") || p.depth > 1000 {
			break
		}
		nodes = append(nodes, p.parseNode())
	}
	return nodes
}

func (p *HTMLParser) parseNode() *dom.Node {
	if p.input[p.pos] == '<' {
		return p.parseElement()
	}
	return p.parseText()
}

func (p *HTMLParser) parseElement() *dom.Node {
	p.depth++
	defer func() { p.depth-- }()

	// Start tag
	p.consumeChar() // '<'
	tagName := p.parseTagName()
	attrs := p.parseAttributes()
	if !p.eof() && p.input[p.pos] == '>' {
		p.consumeChar()
	}

	// Void elements (simplified)
	if tagName == "img" || tagName == "br" || tagName == "hr" || tagName == "meta" || tagName == "link" || tagName == "input" {
		return dom.Element(tagName, attrs, nil)
	}

	// Children
	children := p.parseNodes()

	// End tag
	if !p.eof() && p.input[p.pos] == '<' {
		p.consumeChar()
		if !p.eof() && p.input[p.pos] == '/' {
			p.consumeChar()
			p.consumeTagName()
			if !p.eof() && p.input[p.pos] == '>' {
				p.consumeChar()
			}
		}
	}

	return dom.Element(tagName, attrs, children)
}

func (p *HTMLParser) parseText() *dom.Node {
	start := p.pos
	for !p.eof() && p.input[p.pos] != '<' {
		p.consumeChar()
	}
	return dom.Text(p.input[start:p.pos])
}

func (p *HTMLParser) parseTagName() string {
	start := p.pos
	for !p.eof() && isLetterOrDigit(p.input[p.pos]) {
		p.consumeChar()
	}
	return p.input[start:p.pos]
}

func (p *HTMLParser) consumeTagName() {
	for !p.eof() && isLetterOrDigit(p.input[p.pos]) {
		p.consumeChar()
	}
}

func (p *HTMLParser) parseAttributes() dom.AttrMap {
	attrs := make(dom.AttrMap)
	for {
		p.consumeWhitespace()
		if p.eof() || p.input[p.pos] == '>' {
			break
		}
		name, value := p.parseAttribute()
		attrs[name] = value
	}
	return attrs
}

func (p *HTMLParser) parseAttribute() (string, string) {
	name := p.parseTagName()
	p.consumeWhitespace()
	if !p.eof() && p.input[p.pos] == '=' {
		p.consumeChar()
		p.consumeWhitespace()
		value := p.parseAttributeValue()
		return name, value
	}
	return name, ""
}

func (p *HTMLParser) parseAttributeValue() string {
	if p.eof() {
		return ""
	}
	quote := p.input[p.pos]
	if quote != '"' && quote != '\'' {
		// Unquoted attribute value?
		start := p.pos
		for !p.eof() && !isWhitespace(p.input[p.pos]) && p.input[p.pos] != '>' {
			p.consumeChar()
		}
		return p.input[start:p.pos]
	}
	p.consumeChar()
	start := p.pos
	for !p.eof() && p.input[p.pos] != quote {
		p.consumeChar()
	}
	value := p.input[start:p.pos]
	if !p.eof() {
		p.consumeChar()
	}
	return value
}

func (p *HTMLParser) consumeWhitespace() {
	for !p.eof() && isWhitespace(p.input[p.pos]) {
		p.consumeChar()
	}
}

func (p *HTMLParser) consumeChar() {
	if p.pos < len(p.input) {
		p.pos++
	}
}

func (p *HTMLParser) eof() bool {
	return p.pos >= len(p.input)
}

func isLetterOrDigit(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

func isWhitespace(c byte) bool {
	return c == ' ' || c == '\n' || c == '\t' || c == '\r'
}
