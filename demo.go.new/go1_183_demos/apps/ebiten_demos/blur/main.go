package main

import (
	"bytes"
	"image"
	_ "image/jpeg"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/images"
)

// Refer: https://ebitengine.org/en/examples/blur.html

const screenWidth, screenHeight = 640, 480

var gophersImage *ebiten.Image

func init() {
	img, _, err := image.Decode(bytes.NewReader(images.FiveYears_jpg))
	if err != nil {
		log.Fatal(err)
	}

	imgSavePath := "/tmp/test/img/FiveYears.jpg"
	if err = os.WriteFile(imgSavePath, images.FiveYears_jpg, 0644); err != nil {
		log.Fatal(err)
	}

	gophersImage = ebiten.NewImageFromImage(img)
}

type Game struct {
	isDebug bool
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, 0)
	screen.DrawImage(gophersImage, op)

	layers := 0
	for j := -3; j <= 3; j++ {
		for i := -3; i <= 3; i++ {
			layers++
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(i), 244+float64(j))
			op.ColorScale.ScaleAlpha(1 / float32(layers))
			if g.isDebug {
				log.Printf("draw image: j=%d, i=%d, layers=%d", j, i, layers)
			}
			screen.DrawImage(gophersImage, op)
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Blur (Ebitengine Demo)")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
