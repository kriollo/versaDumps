package main

import (
	"fmt"
	"image/png"
	"os"

	ico "github.com/Kodeworks/golang-image-ico"
)

func iconMain() {
	in := "build/appicon.png"
	out := "build/windows/icon.ico"
	if len(os.Args) >= 2 {
		in = os.Args[1]
	}
	if len(os.Args) >= 3 {
		out = os.Args[2]
	}

	file, err := os.Open(in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening %s: %v\n", in, err)
		return
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error decoding png: %v\n", err)
		return
	}

	output, err := os.Create(out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating %s: %v\n", out, err)
		return
	}
	defer output.Close()

	if err := ico.Encode(output, img); err != nil {
		fmt.Fprintf(os.Stderr, "error encoding ico: %v\n", err)
		return
	}

	fmt.Println("✓ Icono ICO generado exitosamente ->", out)
}
