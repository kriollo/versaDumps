package main

import (
	"fmt"
	"image/png"
	"os"

	ico "github.com/Kodeworks/golang-image-ico"
)

func main() {
	in := "../build/appicon.png"
	out := "../build/windows/icon.ico"
	if len(os.Args) >= 2 {
		in = os.Args[1]
	}
	if len(os.Args) >= 3 {
		out = os.Args[2]
	}

	file, err := os.Open(in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening %s: %v\n", in, err)
		os.Exit(1)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error decoding png: %v\n", err)
		os.Exit(1)
	}

	output, err := os.Create(out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating %s: %v\n", out, err)
		os.Exit(1)
	}
	defer output.Close()

	if err := ico.Encode(output, img); err != nil {
		fmt.Fprintf(os.Stderr, "error encoding ico: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Icono ICO generado exitosamente ->", out)
}
