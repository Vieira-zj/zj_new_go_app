package main

import (
	"bytes"
	"image"
	_ "image/png"
	"log"
	"os"
	"strings"

	"github.com/hajimehoshi/bitmapfont/v3"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/keyboard/keyboard"
	rkeyboard "github.com/hajimehoshi/ebiten/v2/examples/resources/images/keyboard"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const (
	screenWidth  = 320
	screenHeight = 240
)

var keyboardImage *ebiten.Image

var fontFace = text.NewGoXFace(bitmapfont.Face)

func init() {
	img, _, err := image.Decode(bytes.NewReader(rkeyboard.Keyboard_png))
	if err != nil {
		log.Fatal(err)
	}

	imgSavePath := "/tmp/test/img/Keyboard.png"
	if err = os.WriteFile(imgSavePath, rkeyboard.Keyboard_png, 0644); err != nil {
		log.Fatal(err)
	}

	keyboardImage = ebiten.NewImageFromImage(img)
}

type Game struct {
	keys []ebiten.Key
}

func (g *Game) Update() error {
	// Get press keys from input.
	g.keys = inpututil.AppendPressedKeys(g.keys[:0])
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	const (
		offsetX = 24
		offsetY = 40
	)

	// Draw the base (grayed) keyboard image.
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(offsetX, offsetY)
	op.ColorScale.Scale(0.5, 0.5, 0.5, 1)
	screen.DrawImage(keyboardImage, op)

	// Draw the highlighted for press keys.
	op = &ebiten.DrawImageOptions{}
	for _, p := range g.keys {
		op.GeoM.Reset()
		r, ok := keyboard.KeyRect(p)
		if !ok {
			continue
		}
		op.GeoM.Translate(float64(r.Min.X), float64(r.Min.Y))
		op.GeoM.Translate(offsetX, offsetY)
		screen.DrawImage(keyboardImage.SubImage(r).(*ebiten.Image), op)
	}

	var (
		keyStrs  []string
		keyNames []string
	)
	for _, k := range g.keys {
		keyStrs = append(keyStrs, k.String())
		if name := ebiten.KeyName(k); name != "" {
			keyNames = append(keyNames, name)
		}
	}

	// Use bitmapfont.Face instead of ebitenutil.DebugPrint, since some key names might not be printed with DebugPrint.
	textOp := &text.DrawOptions{}
	textOp.LineSpacing = fontFace.Metrics().HLineGap + fontFace.Metrics().HAscent + fontFace.Metrics().HDescent
	text.Draw(screen, strings.Join(keyStrs, ", ")+"\n"+strings.Join(keyNames, ", "), fontFace, textOp)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Keyboard (Ebitengine Demo)")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
