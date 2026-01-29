package main

import (
	"fmt"
	"image"
	"io"
	"net/http"
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
		return
	}
	defer win.Close()

	// 2. Browser State
	currentURL := "https://prymis.browser"
	typingBuffer := ""
	html := `<html><body><div class="container"><div class="header">Prymis Navigation</div><div class="content"><div class="main">Prymis Engine is Ready. Type a URL in the browser window!</div></div></div></body></html>`
	css := `
	.container { background-color: white; }
	.header { background-color: #282c34; color: white; }
	.content { background-color: #e5e5e5; }
	.main { background-color: white; padding: 20px; }
	`

	fmt.Println("Prymis is ready! Interface is inside the window.")

	// 3. Main Loop
	needsRender := true
	for {
		// handle X11 events
		if ev := win.PollEvent(); ev != nil {
			if ev.Type == gui.KeyPress {
				if ev.Key == 13 { // Enter
					currentURL = typingBuffer
					typingBuffer = ""
					fmt.Printf("Navigating to: %s\n", currentURL)
					if strings.HasPrefix(currentURL, "http") {
						resp, err := http.Get(currentURL)
						if err == nil {
							body, _ := io.ReadAll(resp.Body)
							html = string(body)
							resp.Body.Close()
						} else {
							html = fmt.Sprintf("<html><body><h1>Error</h1><p>%v</p></body></html>", err)
						}
					}
				} else if ev.Key == 8 { // Backspace
					if len(typingBuffer) > 0 {
						typingBuffer = typingBuffer[:len(typingBuffer)-1]
					}
				} else if ev.Key != 0 {
					typingBuffer += string(ev.Key)
				}
				needsRender = true
			} else if ev.Type == gui.Expose {
				needsRender = true
			}
		}

		if needsRender {
			// Safety: Recover from parser/layout panics
			canvas := func() *image.RGBA {
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("⚠️ Render Panic: %v\n", r)
					}
				}()

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

				// Use typingBuffer if active, otherwise currentURL
				displayText := currentURL
				if typingBuffer != "" {
					displayText = typingBuffer + "_"
				}
				return render.Paint(layoutTree, image.Rect(0, 0, 800, 600), displayText)
			}()

			if canvas != nil {
				win.Draw(canvas)
			}
			needsRender = false
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func runHeadless() {
	// ... (Previous file-based logic)
}
