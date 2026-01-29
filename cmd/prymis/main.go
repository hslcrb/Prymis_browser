package main

import (
	"bufio"
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"prymis/engine/gui"
	"prymis/engine/layout"
	"prymis/engine/parser"
	"prymis/engine/render"
	"strings"
	"time"
)

func main() {
	fmt.Println("Prymis Browser - Launching GUI...")

	// 1. Initialize X11 Window
	win, err := gui.NewX11Window(800, 600)
	if err != nil {
		fmt.Printf("Error launching GUI: %v\n", err)
		fmt.Println("Falling back to headless mode...")
		runHeadless()
		return
	}
	defer win.Close()

	// 2. Browser State
	currentURL := "https://prymis.browser"
	html := `<html><body><div class="container"><div class="header">Prymis Navigation</div><div class="content"><div class="main">Type URL in terminal to browse (Demo)</div></div></div></body></html>`
	css := `
	.container { background-color: white; }
	.header { background-color: #282c34; color: white; }
	.content { background-color: #e5e5e5; }
	.main { background-color: white; }
	`

	// Channel for terminal input to change URL/Content
	inputChan := make(chan string)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			inputChan <- scanner.Text()
		}
	}()

	fmt.Println("Prymis is ready! Enter an HTTP URL to browse.")
	fmt.Print("> ")

	// 3. Main Loop
	needsRender := true
	for {
		select {
		case input := <-inputChan:
			currentURL = input
			if strings.HasPrefix(input, "http") {
				fmt.Printf("Navigating to: %s\n", input)
				resp, err := http.Get(input)
				if err == nil {
					body, _ := io.ReadAll(resp.Body)
					html = string(body)
					resp.Body.Close()
				} else {
					html = fmt.Sprintf("<html><body><h1>Error</h1><p>%v</p></body></html>", err)
				}
			} else {
				// Local search/input
				html = fmt.Sprintf("<html><body><div class='container'><div class='header'>Prymis Search</div><div class='content'><div class='main'>You entered: %s</div></div></div></body></html>", input)
			}
			needsRender = true
			fmt.Print("> ")

		default:
			if needsRender {
				p := parser.NewHTMLParser(html)
				domTree := p.Parse()
				cp := parser.NewCSSParser(css)
				rules := cp.Parse()
				styleTree := layout.NewStyledNode(domTree, rules)
				layoutTree := layout.NewLayoutTree(styleTree)
				viewport := layout.Dimensions{
					Content: layout.Rect{X: 0, Y: 100, Width: 800, Height: 0},
				}
				layoutTree.Layout(viewport)

				canvas := render.Paint(layoutTree, image.Rect(0, 0, 800, 600), currentURL)
				win.Draw(canvas)
				needsRender = false
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func runHeadless() {
	// ... (Previous file-based logic)
}
