package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
)

func DiffImage(img1, img2 image.Image) (image.Image, string) {
	size1 := img1.Bounds().Size()
	size2 := img2.Bounds().Size()
	if size1.X != size2.X {
		log.Fatalln("images X size is not match")
	}
	if size1.Y != size2.Y {
		log.Fatalln("images Y size is not match")
	}

	diffsCount := 0
	red := color.RGBA64{0xffff, 0, 0, 0x4000}
	diffImg := image.NewRGBA(image.Rect(0, 0, size1.X, size1.Y))
	for y := 0; y < size1.Y; y++ {
		for x := 0; x < size1.X; x++ {
			r1, g1, b1, a1 := img1.At(x, y).RGBA()
			r2, g2, b2, a2 := img2.At(x, y).RGBA()
			dst := color.RGBA64{uint16(r2), uint16(g2), uint16(b2), uint16(a2)}
			if r1+g1+b1+a1 != r2+g2+b2+a2 {
				diffsCount++
				diffImg.Set(x, y, blend(dst, red))
			} else {
				diffImg.Set(x, y, dst)
			}
		}
	}

	diffRate := float32(diffsCount) / (float32(size1.X) * float32(size1.Y))
	passRate := 100 - diffRate*100
	return diffImg, fmt.Sprintf("%.2f", passRate) + "%"
}

func blend(col1, col2 color.Color) color.Color {
	col1R, col1G, col1B, _ := col1.RGBA()
	col2R, col2G, col2B, col2A := col2.RGBA()
	col2Ap := float64(col2A) / 0xffff

	outR := float64(col2R)*col2Ap + float64(col1R)*(1-col2Ap)
	outG := float64(col2G)*col2Ap + float64(col1G)*(1-col2Ap)
	outB := float64(col2B)*col2Ap + float64(col1B)*(1-col2Ap)

	return color.NRGBA64{
		uint16(outR),
		uint16(outG),
		uint16(outB),
		0xffff,
	}
}
