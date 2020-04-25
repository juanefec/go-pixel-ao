package main

import (
	"fmt"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/golang/image/colornames"
	"golang.org/x/image/font/basicfont"
)

func SetNameWindow() (string, error) {
	cfg := pixelgl.WindowConfig{
		Title:  "Creative AO | Login",
		Bounds: pixel.R(0, 0, 300, 300),
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	defer win.Destroy()

	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	nickname := text.New(pixel.V(50, 100), atlas)
	nickname.Color = colornames.Lightgrey

	txt := text.New(pixel.V(0, 0), atlas)
	txt.Color = colornames.Lightgray
	txt.WriteString("Enter nickname:\n")

	fps := time.Tick(time.Second / 120)

	nn := ""

	for !win.Closed() {

		nickname.WriteString(win.Typed())
		if win.Typed() != "" {
			nn = fmt.Sprint(nn, win.Typed())
		}
		if win.JustPressed(pixelgl.KeyEnter) || win.Repeated(pixelgl.KeyEnter) {
			return nn, nil
		}

		win.Clear(colornames.Black)
		txt.Draw(win, pixel.IM.Moved(win.Bounds().Center().Sub(txt.Bounds().Center()).Add(pixel.V(0, 30))).Scaled(win.Bounds().Center(), 2))

		nickname.Draw(win, pixel.IM.Moved(win.Bounds().Center().Sub(nickname.Bounds().Center())).Scaled(win.Bounds().Center(), 2))
		win.Update()

		<-fps
	}
	return "", fmt.Errorf("No se ingreso el nombre correctamente")
}
