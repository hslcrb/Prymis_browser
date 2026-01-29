package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"prymis/engine/dom"
	"prymis/engine/layout"
	"prymis/engine/parser"
	"prymis/engine/render"
)

func main() {
	fmt.Println("Prymis Browser - Starting Engine...")

	// 1. Setup sample HTML and CSS
	html := `<html><body><div class="container"><div class="header">Prymis Browser</div><div class="content"><div class="sidebar">Menu</div><div class="main">Main Content Area</div></div><div class="footer">Built from scratch in Go</div></div></body></html>`

	css := `
	.container { background-color: white; }
	.header { background-color: blue; }
	.content { background-color: gray; }
	.sidebar { background-color: red; }
	.main { background-color: white; }
	.footer { background-color: green; }
	`

	// 2. Parse HTML
	fmt.Println("Parsing HTML...")
	p := parser.NewHTMLParser(html)
	domTree := p.Parse()

	// 3. Parse CSS
	fmt.Println("Parsing CSS...")
	cp := parser.NewCSSParser(css)
	rules := cp.Parse()

	// 4. Create Style Tree
	fmt.Println("Creating Style Tree...")
	styleTree := layout.NewStyledNode(domTree, rules)

	// 5. Create Layout Tree
	fmt.Println("Calculating Layout...")
	layoutTree := layout.NewLayoutTree(styleTree)
	viewport := layout.Dimensions{
		Content: layout.Rect{Width: 800, Height: 0},
	}
	layoutTree.Layout(viewport)

	// 6. Render to Image
	fmt.Println("Rendering to Image...")
	canvas := render.Paint(layoutTree, image.Rect(0, 0, 800, 600))

	// 7. Save output
	f, err := os.Create("output.png")
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer f.Close()
	png.Encode(f, canvas)

	fmt.Println("Successfully rendered Prymis output to output.png")

	// Print DOM tree for verification
	fmt.Println("\nDOM Tree Visualization:")
	printNode(domTree, 0)
}

func printNode(n *dom.Node, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	if n.NodeType == dom.TextNode {
		fmt.Printf("%s#text: %s\n", indent, n.Text)
	} else {
		fmt.Printf("%s<%s>\n", indent, n.TagName)
	}
	for _, child := range n.Children {
		printNode(child, depth+1)
	}
}
