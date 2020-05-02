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

type LoginStep int

const (
	Name LoginStep = iota
	ChooseWizard
)

func LoginWindow() (Wizard, error) {

	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	nickname := text.New(pixel.V(50, 100), atlas)
	nickname.Color = colornames.Lightgrey

	txt := text.New(pixel.V(0, 0), atlas)
	txt.Color = colornames.Lightgray
	txt.WriteString("Enter nickname:\n")

	choose := text.New(pixel.V(0, 0), atlas)
	choose.Color = colornames.Darkgray
	choose.WriteString("Choose wizard:\n")

	redSkin := Pictures["./images/bodyRedIcon.png"]
	redIcon := pixel.NewSprite(redSkin, redSkin.Bounds())

	blueSkin := Pictures["./images/bodyBlueIcon.png"]
	blueIcon := pixel.NewSprite(blueSkin, blueSkin.Bounds())

	darkSkin := Pictures["./images/darkopshitIcon.png"]
	darkIcon := pixel.NewSprite(darkSkin, darkSkin.Bounds())

	armbluSkin := Pictures["./images/placaazulIcon.png"]
	armbluIcon := pixel.NewSprite(armbluSkin, armbluSkin.Bounds())

	druidSkin := Pictures["./images/bodydruidaIcon.png"]
	druidIcon := pixel.NewSprite(druidSkin, druidSkin.Bounds())

	twilightSkin := Pictures["./images/penumbrasIcon.png"]
	twilightIcon := pixel.NewSprite(twilightSkin, twilightSkin.Bounds())

	cfg := pixelgl.WindowConfig{
		Title:  "Creative AO | Login",
		Bounds: pixel.R(0, 0, 600, 500),
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
				if nn == "   creagod   " {
					return Wizard{
						Skin:          GodBody,
						Name:          nn,
						Type:          Sniper,
						SpecialSpells: []string{"icesnipe", "smoke-spot"},
						Intervals:     []float64{BasicSpellInterval, IcesnipeSpellInterval / 4, LavaSpellInterval / 8},
					}, nil
				}
				loginStep = ChooseWizard
			}
		}

		if loginStep == ChooseWizard {
			dist := 83.0
			bPos := pixel.V(dist, 160)
			rPos := pixel.V(dist*2, 160)
			armbluPos := pixel.V(dist*3, 160)
			darkPos := pixel.V(dist*4, 160)
			druidPos := pixel.V(dist*5, 160)
			twilightPos := pixel.V(dist*6, 160)

			blueIcon.Draw(win, pixel.IM.Moved(bPos).Scaled(bPos, 2))
			redIcon.Draw(win, pixel.IM.Moved(rPos).Scaled(rPos, 2))
			armbluIcon.Draw(win, pixel.IM.Moved(armbluPos).Scaled(armbluPos, 2))
			darkIcon.Draw(win, pixel.IM.Moved(darkPos).Scaled(darkPos, 2))
			druidIcon.Draw(win, pixel.IM.Moved(druidPos).Scaled(druidPos, 2))
			twilightIcon.Draw(win, pixel.IM.Moved(twilightPos).Scaled(twilightPos, 2))

			choose.Draw(win, pixel.IM.Moved(win.Bounds().Center().Sub(choose.Bounds().Center()).Add(pixel.V(0, 0))).Scaled(win.Bounds().Center(), 2))
			if win.Pressed(pixelgl.MouseButtonLeft) {
				x, y := win.MousePosition().XY()
				halfdist := dist / 2
				if x < dist+halfdist && x > dist-halfdist && y > 110 && y < 210 {
					return Wizard{
						Name:          nn,
						Skin:          BlueBody,
						Type:          Monk,
						SpecialSpells: []string{"healshot", "heal-spot"},
						Intervals:     []float64{BasicSpellInterval, FireballSpellInterval, LavaSpellInterval},
					}, nil

				}
				// if x < (dist*2)+halfdist && x > (dist*2)-halfdist && y > 110 && y < 210 {
				// 	return Wizard{
				// 		Skin:          RedBody,
				// 		Name:          nn,
				// 		Type:          Hunter,
				// 		SpecialSpells: []string{"arrowshot", "bear-trap"},
				// 		Intervals:     []float64{BasicSpellInterval, FireballSpellInterval, 0},
				// 	}, nil
				// }
				if x < (dist*3)+halfdist && x > (dist*3)-halfdist && y > 110 && y < 210 {
					return Wizard{
						Skin:          BlueArmorBody,
						Name:          nn,
						Type:          Sniper,
						SpecialSpells: []string{"icesnipe", "smoke-spot"},
						Intervals:     []float64{BasicSpellInterval, IcesnipeSpellInterval, LavaSpellInterval},
					}, nil
				}
				if x < (dist*4)+halfdist && x > (dist*4)-halfdist && y > 110 && y < 210 {
					return Wizard{
						Name:          nn,
						Skin:          DarkMasterBody,
						Type:          DarkWizard,
						SpecialSpells: []string{"fireball", "lava-spot"},
						Intervals:     []float64{BasicSpellInterval, FireballSpellInterval, LavaSpellInterval},
					}, nil
				}
				if x < (dist*5)+halfdist && x > (dist*5)-halfdist && y > 110 && y < 210 {
					return Wizard{
						Skin:          TuniDruida,
						Name:          nn,
						Type:          Shaman,
						SpecialSpells: []string{"manashot", "mana-spot"},
						Intervals:     []float64{BasicSpellInterval, FireballSpellInterval, LavaSpellInterval},
					}, nil
				}
				if x < (dist*6)+halfdist && x > (dist*6)-halfdist && y > 110 && y < 210 {
					return Wizard{
						Skin:          TwilightBody,
						Name:          nn,
						Type:          Timewreker,
						SpecialSpells: []string{"rockshot", "flash"},
						Intervals:     []float64{BasicSpellInterval, RockSpellInterval, FlashSpellInterval},
					}, nil

				}
			}
		}

		txt.Draw(win, pixel.IM.Moved(win.Bounds().Center().Sub(txt.Bounds().Center()).Add(pixel.V(0, 100))).Scaled(win.Bounds().Center(), 2))
		nickname.Draw(win, pixel.IM.Moved(win.Bounds().Center().Sub(nickname.Bounds().Center()).Add(pixel.V(0, 70))).Scaled(win.Bounds().Center(), 2))
		win.Update()
		<-fps
	}
	return Wizard{}, fmt.Errorf("No se ingreso el nombre correctamente")
}

func inBody(skin SkinType) {}
