package main

import (
	"fmt"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/golang/image/colornames"
	"golang.org/x/image/font/basicfont"
)

type LoginStep int

const (
	Name LoginStep = iota
	Color
)

type LoginData struct {
	Name string
	Side SkinType
}

func SetNameWindow() (LoginData, error) {

	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	nickname := text.New(pixel.V(50, 100), atlas)
	nickname.Color = colornames.Lightgrey

	txt := text.New(pixel.V(0, 0), atlas)
	txt.Color = colornames.Lightgray
	txt.WriteString("Enter nickname:\n")

	choose := text.New(pixel.V(0, 0), atlas)
	choose.Color = colornames.Lightgray
	choose.WriteString("Choose side:\n")

	redSide := Pictures["./images/bodyRedIcon.png"]
	rIcon := pixel.NewSprite(redSide, redSide.Bounds())

	blueSide := Pictures["./images/bodyBlueIcon.png"]
	bIcon := pixel.NewSprite(blueSide, blueSide.Bounds())

	cfg := pixelgl.WindowConfig{
		Title:  "Creative AO | Login",
		Bounds: pixel.R(0, 0, 300, 500),
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	defer win.Destroy()

	fps := time.Tick(time.Second / 120)

	nn := ""
	loginStep := Name
	for !win.Closed() {
		win.Clear(colornames.Black)

		if loginStep == Name {
			nickname.WriteString(win.Typed())
			if win.Typed() != "" {
				nn = fmt.Sprint(nn, win.Typed())
			}
			if win.JustPressed(pixelgl.KeyBackspace) || win.Repeated(pixelgl.KeyBackspace) {
				if nn != "" {
					nn = nn[:len(nn)-1]
					nickname.Clear()
					nickname.WriteString(nn)
				}
			}
			if win.JustPressed(pixelgl.KeyEnter) || win.Repeated(pixelgl.KeyEnter) {
				loginStep = Color
			}
		}

		if loginStep == Color {
			bframe := imdraw.New(nil)
			bframe.Color = colornames.Grey
			bframe.EndShape = imdraw.SharpEndShape
			bframe.Push(
				pixel.V(120, 110),
				pixel.V(120, 210),
				pixel.V(20, 110),
				pixel.V(20, 210),
			)
			bframe.Rectangle(0)

			bframe.Push(
				pixel.V(280, 110),
				pixel.V(280, 210),
				pixel.V(180, 110),
				pixel.V(180, 210),
			)
			bframe.Rectangle(0)
			bframe.Draw(win)
			bPos := pixel.V(70, 160)
			bIcon.Draw(win, pixel.IM.Moved(bPos).Scaled(bPos, 2))
			rPos := pixel.V(230, 160)
			rIcon.Draw(win, pixel.IM.Moved(rPos).Scaled(rPos, 2))
			choose.Draw(win, pixel.IM.Moved(win.Bounds().Center().Sub(choose.Bounds().Center()).Add(pixel.V(0, 0))).Scaled(win.Bounds().Center(), 2))
			if win.Pressed(pixelgl.MouseButtonLeft) {
				x, y := win.MousePosition().XY()
				if x < 120 && x > 20 && y > 110 && y < 210 {
					return LoginData{
						Name: nn,
						Side: BlueBody,
					}, nil
				}
				if x < 280 && x > 180 && y > 110 && y < 210 {
					return LoginData{
						Name: nn,
						Side: RedBody,
					}, nil
				}
			}
		}

		txt.Draw(win, pixel.IM.Moved(win.Bounds().Center().Sub(txt.Bounds().Center()).Add(pixel.V(0, 100))).Scaled(win.Bounds().Center(), 2))
		nickname.Draw(win, pixel.IM.Moved(win.Bounds().Center().Sub(nickname.Bounds().Center()).Add(pixel.V(0, 70))).Scaled(win.Bounds().Center(), 2))
		win.Update()
		<-fps
	}
	return LoginData{}, fmt.Errorf("No se ingreso el nombre correctamente")
}
