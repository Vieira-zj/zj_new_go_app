package main

import (
	"bytes"
	"image"
	"image/color"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/images"
)

// Refer: https://ebitengine.org/en/examples/flood.html

const screenWidth, screenHeight = 320, 240

var (
	ebitenImage *ebiten.Image

	colors = []color.RGBA{
		{0xff, 0xff, 0xff, 0xff},
		{0xff, 0xff, 0x0, 0xff},
		{0xff, 0x0, 0xff, 0xff},
		{0xff, 0x0, 0x0, 0xff},
		{0x0, 0xff, 0xff, 0xff},
		{0x0, 0xff, 0x0, 0xff},
		{0x0, 0x0, 0xff, 0xff},
		{0x0, 0x0, 0x0, 0xff},
	}
)

func init() {
	img, _, err := image.Decode(bytes.NewReader(images.Ebiten_png))
	if err != nil {
		log.Fatal(err)
	}

	ebitenImage = ebiten.NewImageFromImage(img)
}

type Game struct {
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	const ox, oy, dx, dy = 10, 10, 60, 50

	screen.Fill(color.NRGBA{0x00, 0x40, 0x80, 0xff})

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(ox, oy)
	screen.DrawImage(ebitenImage, op)

	// Fill with solid colors
	for i, c := range colors {
		x := i % 4
		y := i/4 + 1
		op := &colorm.DrawImageOptions{}
		op.GeoM.Translate(ox+float64(dx*x), oy+float64(dy*y))

		// Reset RGB (not Alpha) 0 forcibly
		var cm colorm.ColorM
		cm.Scale(0, 0, 0, 1)

		// Set color
		r := float64(c.R) / 0xff
		g := float64(c.G) / 0xff
		b := float64(c.B) / 0xff
		cm.Translate(r, g, b, 0)
		colorm.DrawImage(screen, ebitenImage, cm, op)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Flood fill with solid colors (Ebitengine Demo)")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
