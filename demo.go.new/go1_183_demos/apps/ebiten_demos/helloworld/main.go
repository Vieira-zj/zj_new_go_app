package main

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const screenWidth, screenHeight = 640, 480

var img *ebiten.Image

func init() {
	var err error
	imgPath := "/tmp/test/img/gopher.png"
	img, _, err = ebitenutil.NewImageFromFile(imgPath)
	if err != nil {
		log.Fatal(err)
	}
}

type Game struct{}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0xff, 0xe0, 0xe6, 0xff})

	op := &ebiten.DrawImageOptions{}
	// the image is moved by 50 pixels rightward, and by 50 pixels downward
	op.GeoM.Translate(50, 50)
	op.GeoM.Scale(1.5, 1)
	screen.DrawImage(img, op)

	ebitenutil.DebugPrint(screen, "Hello, World!")
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Hello, World!")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
