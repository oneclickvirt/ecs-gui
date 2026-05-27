package main

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed assets/icons/logo.png
var appIconPNG []byte

func appIconResource() fyne.Resource {
	return fyne.NewStaticResource("logo.png", appIconPNG)
}
