package render

import (
	"image"
	"image/color"
	"image/draw"
	"prymis/engine/dom"
	"prymis/engine/layout"
)

func Paint(root *layout.LayoutBox, bounds image.Rectangle, url string) *image.RGBA {
	canvas := image.NewRGBA(bounds)

	// 1. Draw Browser Window Frame (Dark Theme)
	frameColor := color.RGBA{33, 37, 43, 255}
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{frameColor}, image.Point{}, draw.Src)

	// 2. Draw Window Controls (Red, Yellow, Green circles)
	drawCircle(canvas, 20, 20, 6, color.RGBA{255, 95, 87, 255})
	drawCircle(canvas, 40, 20, 6, color.RGBA{255, 189, 46, 255})
	drawCircle(canvas, 60, 20, 6, color.RGBA{39, 201, 63, 255})

	// 4. Draw Address Bar
	addressBarRect := image.Rect(100, 50, 700, 85)
	draw.Draw(canvas, addressBarRect, &image.Uniform{color.RGBA{30, 33, 39, 255}}, image.Point{}, draw.Src)
	drawBorder(canvas, addressBarRect, color.RGBA{100, 100, 100, 255})

	// Draw URL simulation (dots/lines representing text)
	drawTextSimulation(canvas, 120, 67, url, color.RGBA{180, 180, 180, 255})

	// 5. Draw Prymis Logo (Stylized 'P')
	drawLogo(canvas, 750, 25)

	// 5. Draw Content Area Background
	contentArea := image.Rect(0, 100, 800, 600)
	draw.Draw(canvas, contentArea, &image.Uniform{color.White}, image.Point{}, draw.Src)

	// 6. Offset rendering to content area
	viewport := layout.Dimensions{
		Content: layout.Rect{X: 0, Y: 100, Width: 800, Height: 0},
	}
	root.Layout(viewport)

	renderBox(canvas, root)
	return canvas
}

func renderBox(canvas *image.RGBA, box *layout.LayoutBox) {
	// Draw background color
	if bg := box.StyledNode.SpecifiedValues["background-color"]; bg != "" {
		c := parseColor(bg)
		rect := image.Rect(
			int(box.Dimensions.Content.X),
			int(box.Dimensions.Content.Y),
			int(box.Dimensions.Content.X+box.Dimensions.Content.Width),
			int(box.Dimensions.Content.Y+box.Dimensions.Content.Height),
		)
		draw.Draw(canvas, rect, &image.Uniform{c}, image.Point{}, draw.Src)
	}

	// Draw border (simple 1px black border for visibility)
	if box.StyledNode.Node.NodeType == dom.ElementNode {
		rect := image.Rect(
			int(box.Dimensions.Content.X),
			int(box.Dimensions.Content.Y),
			int(box.Dimensions.Content.X+box.Dimensions.Content.Width),
			int(box.Dimensions.Content.Y+box.Dimensions.Content.Height),
		)
		drawBorder(canvas, rect, color.Black)
	}

	for _, child := range box.Children {
		renderBox(canvas, child)
	}
}

func drawLogo(canvas *image.RGBA, x, y int) {
	// Stylized 'P'
	c := color.RGBA{100, 150, 255, 255}
	// Vertical bar
	for i := -10; i < 15; i++ {
		for j := -2; j < 2; j++ {
			canvas.Set(x+j, y+i, c)
		}
	}
	// Curve
	drawCircle(canvas, x+5, y-5, 7, c)
	drawCircle(canvas, x+5, y-5, 4, color.RGBA{33, 37, 43, 255})
}

func drawTextSimulation(canvas *image.RGBA, x, y int, text string, c color.Color) {
	// Draw dots for characters
	for i := 0; i < len(text); i++ {
		charX := x + i*8
		// Randomize height slightly for 'text-like' feel
		h := 2
		if text[i] == '/' || text[i] == '.' {
			h = 4
		}
		rect := image.Rect(charX, y-h, charX+5, y+1)
		draw.Draw(canvas, rect, &image.Uniform{c}, image.Point{}, draw.Src)
	}
}

func drawCircle(canvas *image.RGBA, x, y, r int, c color.Color) {
	for i := x - r; i <= x+r; i++ {
		for j := y - r; j <= y+r; j++ {
			if (i-x)*(i-x)+(j-y)*(j-y) <= r*r {
				canvas.Set(i, j, c)
			}
		}
	}
}

func drawBorder(canvas *image.RGBA, r image.Rectangle, c color.Color) {
	for x := r.Min.X; x < r.Max.X; x++ {
		canvas.Set(x, r.Min.Y, c)
		canvas.Set(x, r.Max.Y-1, c)
	}
	for y := r.Min.Y; y < r.Max.Y; y++ {
		canvas.Set(r.Min.X, y, c)
		canvas.Set(r.Max.X-1, y, c)
	}
}

func parseColor(s string) color.Color {
	switch s {
	case "white":
		return color.White
	case "red":
		return color.RGBA{255, 100, 100, 255}
	case "blue":
		return color.RGBA{100, 100, 255, 255}
	case "green":
		return color.RGBA{100, 255, 100, 255}
	case "black":
		return color.Black
	case "gray":
		return color.RGBA{200, 200, 200, 255}
	default:
		return color.White
	}
}
