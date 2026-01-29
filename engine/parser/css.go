package parser

import (
	"fmt"
	"strings"
)

type StyleRule struct {
	Selectors    []string
	Declarations []Declaration
}

type Declaration struct {
	Name  string
	Value string
}

type CSSParser struct {
	pos   int
	input string
}

func NewCSSParser(input string) *CSSParser {
	return &CSSParser{input: input}
}

func (p *CSSParser) Parse() []StyleRule {
	fmt.Printf("Starting CSS Parse, input length: %d\n", len(p.input))
	var rules []StyleRule
	for !p.eof() {
		p.consumeWhitespace()
		if p.eof() {
			break
		}
		rules = append(rules, p.parseRule())
	}
	return rules
}

func (p *CSSParser) parseRule() StyleRule {
	return StyleRule{
		Selectors:    p.parseSelectors(),
		Declarations: p.parseDeclarations(),
	}
}

func (p *CSSParser) parseSelectors() []string {
	var selectors []string
	for {
		p.consumeWhitespace()
		selectors = append(selectors, p.parseIdentifier())
		p.consumeWhitespace()
		if p.input[p.pos] == '{' {
			p.consumeChar()
			break
		}
		if p.input[p.pos] == ',' {
			p.consumeChar()
		} else {
			// Safety break or consume
			p.consumeChar()
		}
	}
	return selectors
}

func (p *CSSParser) parseDeclarations() []Declaration {
	var decls []Declaration
	for {
		p.consumeWhitespace()
		if p.input[p.pos] == '}' {
			p.consumeChar()
			break
		}
		decls = append(decls, p.parseDeclaration())
	}
	return decls
}

func (p *CSSParser) parseDeclaration() Declaration {
	name := p.parseIdentifier()
	p.consumeWhitespace()
	p.consumeChar() // ':'
	p.consumeWhitespace()
	value := p.parseValue()
	p.consumeWhitespace()
	p.consumeChar() // ';'
	return Declaration{Name: name, Value: value}
}

func (p *CSSParser) parseIdentifier() string {
	start := p.pos
	for !p.eof() && isIdentifierChar(p.input[p.pos]) {
		p.consumeChar()
	}
	return p.input[start:p.pos]
}

func (p *CSSParser) parseValue() string {
	start := p.pos
	for !p.eof() && p.input[p.pos] != ';' && p.input[p.pos] != '}' {
		p.consumeChar()
	}
	return strings.TrimSpace(p.input[start:p.pos])
}

func (p *CSSParser) consumeWhitespace() {
	for !p.eof() && isWhitespace(p.input[p.pos]) {
		p.consumeChar()
	}
}

func (p *CSSParser) consumeChar() {
	p.pos++
}

func (p *CSSParser) eof() bool {
	return p.pos >= len(p.input)
}

func isIdentifierChar(c byte) bool {
	return isLetterOrDigit(c) || c == '-' || c == '_' || c == '.' || c == '#'
}
