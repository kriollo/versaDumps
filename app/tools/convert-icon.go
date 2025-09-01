package main

import (
	"image/png"
	"os"

	ico "github.com/Kodeworks/golang-image-ico"
)

func main() {
	// Abrir la imagen PNG
	file, err := os.Open("../build/appicon.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Decodificar la imagen PNG
	img, err := png.Decode(file)
	if err != nil {
		panic(err)
	}

	// Crear el archivo ICO
	output, err := os.Create("../build/windows/icon.ico")
	if err != nil {
		panic(err)
	}
	defer output.Close()

	// Codificar como ICO con múltiples tamaños
	err = ico.Encode(output, img)
	if err != nil {
		panic(err)
	}

	println("✓ Icono ICO generado exitosamente")
}
