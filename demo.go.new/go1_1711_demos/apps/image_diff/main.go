package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
)

func main() {
	output := flag.String("output", "", "Output image file path.")
	help := flag.Bool("help", false, "Help.")
	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	args := flag.Args()
	if len(args) != 2 {
		fmt.Println("usage: imagediff [<option>...] <image1> <image2>")
		os.Exit(1)
	}

	img1 := mustLoadImage(args[0])
	img2 := mustLoadImage(args[1])

	diffImg, passRate := DiffImage(img1, img2)
	fmt.Println("Pass rate:", passRate)
	mustSaveImage(diffImg, *output)
}

func mustLoadImage(fPath string) image.Image {
	f, err := os.Open(fPath)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		log.Fatalln(err)
	}
	return img
}

func mustSaveImage(img image.Image, output string) {
	f, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	if err = png.Encode(f, img); err != nil {
		log.Fatalln(err)
	}
}
