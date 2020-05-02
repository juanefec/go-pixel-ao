package main

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

type CursorMode int

const (
	Normal CursorMode = iota
	SpellCastDesca
	SpellCastApoca
	SpellCastExplo
	SpellCastPrimarySkill
	SpellCastSecondarySkill
)

type Cursor struct {
	win  *pixelgl.Window
	Mode CursorMode
}

func NewCursor(win *pixelgl.Window) *Cursor {
	return &Cursor{
		win:  win,
		Mode: Normal,
	}
}

func (c *Cursor) SetSpellFireballMode() {
	c.Mode = SpellCastPrimarySkill
}
func (c *Cursor) SetSpellExploMode() {
	c.Mode = SpellCastExplo
}
func (c *Cursor) SetSpellApocaMode() {
	c.Mode = SpellCastApoca
}
func (c *Cursor) SetSpellDescaMode() {
	c.Mode = SpellCastDesca
}
func (c *Cursor) SetNormalMode() {
	c.Mode = Normal
}

func (c *Cursor) Draw(cam pixel.Matrix) {
	if c.Mode != Normal {
		mouse := cam.Unproject(c.win.MousePosition())
		center := cam.Unproject(c.win.Bounds().Center())
		c.win.SetCursorVisible(false)
		cross := imdraw.New(nil)
		if Dist(mouse, center) <= OnTargetSpellRange {
			cross.Color = colornames.Black
		} else {
			cross.Color = colornames.Red
		}
		cross.EndShape = imdraw.SharpEndShape
		cross.Push(
			mouse.Add(pixel.V(-7, 0)),
			mouse.Add(pixel.V(7, 0)),
		)
		cross.Line(1.65)
		cross.Push(
			mouse.Add(pixel.V(0, 7)),
			mouse.Add(pixel.V(0, -7)),
		)
		cross.Line(1.65)
		cross.Draw(c.win)
	} else {
		c.win.SetCursorVisible(true)
	}
}

type HudComponent int

const (
	HealthNumber HudComponent = iota
	ManaNumber
	OnlineCount
	PosXY
	TypingMark
	FPSCount
	ZoomINButton
	ZoomOUTButton
	KDCount
	RankingTitle
	Ranking1
	Ranking2
	Ranking3
	Ranking4
	Ranking5
	Ranking6
	Ranking7
	Ranking8
	Ranking9
	Ranking10
)

type IconType int

const (
	ApocaIcon IconType = iota
	ExploIcon
	DescaIcon
	FireballIcon
	LavaSpotIcon
	IcesnipeIcon
	SmokeSpotIcon
	HealingballIcon
	HealSpotIcon
	ArrowshotIcon
	BearTrapIcon
	IgniterFireballIcon
	ImplodeIcon
	ManaShotIcon
	ManaSpotIcon
)

type Icons []*Icon
type Icon struct {
	Type   IconType
	Frame  pixel.Rect
	Sprite *pixel.Sprite
	Pic    *pixel.Picture
}

func (is Icons) Load(kind IconType, pos int, imagPath string) {
	sheet := Pictures[imagPath]
	icon := &Icon{
		Type:   kind,
		Pic:    &sheet,
		Sprite: pixel.NewSprite(sheet, sheet.Bounds()),
		Frame:  sheet.Bounds(),
	}
	is[pos] = icon
}

type PlayerInfo struct {
	playersData *PlayersData
	player      *Player
	hudText     []*TextProp
	nfps        int
	skillIcons  Icons
	//ranking     []*models.RankingPosMsg
}

func NewPlayerInfo(player *Player, pd *PlayersData, reload ...Skin) *PlayerInfo {
	pi := PlayerInfo{}
	icons := make(Icons, 5)
	icons.Load(ApocaIcon, 0, "./images/apocaIcon.png")
	icons.Load(ExploIcon, 1, "./images/exploIcon.png")
	icons.Load(DescaIcon, 2, "./images/descaIcon.png")

	switch player.wizard.Type {
	case DarkWizard:
		icons.Load(FireballIcon, 3, "./images/fireballIcon.png")
		icons.Load(LavaSpotIcon, 4, "./images/lavaSpotIcon.png")
		break
	case Sniper:
		icons.Load(IcesnipeIcon, 3, "./images/icesnipeIcon.png")
		icons.Load(SmokeSpotIcon, 4, "./images/smokeSpotIcon.png")
		break
	case Hunter:
		icons.Load(ArrowshotIcon, 3, "./images/xxxxxxxxxx.png")
		icons.Load(BearTrapIcon, 4, "./images/xxxxxxxxxxxx.png")
		break
	case Timewreker:
		icons.Load(IgniterFireballIcon, 3, "./images/rockShotIcon.png")
		icons.Load(ImplodeIcon, 4, "./images/flashEffectIcon.png")
		break
	case Monk:
		icons.Load(HealingballIcon, 3, "./images/healingShotIcon.png")
		icons.Load(HealSpotIcon, 4, "./images/healingSpotIcon.png")
	case Shaman:
		icons.Load(ManaShotIcon, 3, "./images/manaShotIcon.png")
		icons.Load(ManaSpotIcon, 4, "./images/manaSpotIcon.png")
	}

	hudProps := make([]*TextProp, 20)
	basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	hudProps[HealthNumber] = NewTextProp(basicAtlas, "%v/%v", player.hp, MaxHealth)
	hudProps[ManaNumber] = NewTextProp(basicAtlas, "%v/%v", player.mp, MaxMana)
	hudProps[OnlineCount] = NewTextProp(basicAtlas, "Typing...")
	hudProps[PosXY] = NewTextProp(basicAtlas, "Online: %v", pd.Online+1)
	hudProps[TypingMark] = NewTextProp(basicAtlas, "X: %v\nY: %v", player.pos.X, player.pos.Y)
	hudProps[FPSCount] = NewTextProp(basicAtlas, "FPS: %v", 0)
	hudProps[ZoomINButton] = NewTextProp(basicAtlas, "in")
	hudProps[ZoomOUTButton] = NewTextProp(basicAtlas, "out")
	hudProps[KDCount] = NewTextProp(basicAtlas, "K/D: %v/%v", player.kills, player)
	hudProps[RankingTitle] = NewTextProp(basicAtlas, "Top 10 K/D        K    D")
	hudProps[Ranking1] = NewTextProp(basicAtlas, "1: %v 	| %v | %v", "-", 0, 0)
	hudProps[Ranking2] = NewTextProp(basicAtlas, "2: %v 	| %v | %v", "-", 0, 0)
	hudProps[Ranking3] = NewTextProp(basicAtlas, "3: %v 	| %v | %v", "-", 0, 0)
	hudProps[Ranking4] = NewTextProp(basicAtlas, "4: %v 	| %v | %v", "-", 0, 0)
	hudProps[Ranking5] = NewTextProp(basicAtlas, "5: %v 	| %v | %v", "-", 0, 0)
	hudProps[Ranking6] = NewTextProp(basicAtlas, "6: %v 	| %v | %v", "-", 0, 0)
	hudProps[Ranking7] = NewTextProp(basicAtlas, "7: %v 	| %v | %v", "-", 0, 0)
	hudProps[Ranking8] = NewTextProp(basicAtlas, "8: %v 	| %v | %v", "-", 0, 0)
	hudProps[Ranking9] = NewTextProp(basicAtlas, "9: %v 	| %v | %v", "-", 0, 0)
	hudProps[Ranking10] = NewTextProp(basicAtlas, "10: %v 	| %v | %v", "-", 0, 0)

	pi.player = player
	pi.playersData = pd
	pi.hudText = hudProps
	pi.skillIcons = icons
	return &pi
}

type TextProp struct {
	Text  *text.Text
	SText string
}

func NewTextProp(a *text.Atlas, s string, ss ...interface{}) *TextProp {
	tp := &TextProp{
		SText: fmt.Sprintf(s, ss...),
		Text:  text.New(pixel.ZV, a),
	}
	tp.Text.Color = colornames.Whitesmoke
	fmt.Fprint(tp.Text, tp.SText)
	return tp
}

func (tp *TextProp) Draw(win *pixelgl.Window, m pixel.Matrix, s string, ss ...interface{}) {
	tp.Text.Clear()
	tp.SText = fmt.Sprintf(s, ss...)
	fmt.Fprint(tp.Text, tp.SText)
	tp.Text.Draw(win, m)
}

func getRectangleVecs(pos, size pixel.Vec) []pixel.Vec {
	return []pixel.Vec{
		pos,
		pixel.V(pos.X+size.X, pos.Y),
		pixel.V(pos.X, pos.Y+size.Y),
		pixel.V(pos.X+size.X, pos.Y+size.Y),
	}
}

func (pi *PlayerInfo) Draw(win *pixelgl.Window, cam pixel.Matrix, cursor *Cursor, wizard *Wizard) {
	winSize := win.Bounds().Max
	topRigthInfoPos := cam.Unproject(winSize.Add(pixel.V(-330, -70)))
	info := imdraw.New(nil)
	info.Color = colornames.Black
	info.EndShape = imdraw.SharpEndShape
	// Heath Mana info
	info.Push(
		getRectangleVecs(topRigthInfoPos.Add(pixel.V(-2, -2)), pixel.V(154, 24))...,
	)
	info.Rectangle(4)
	info.Push(
		getRectangleVecs(topRigthInfoPos.Add(pixel.V(-2, -32)), pixel.V(154, 24))...,
	)
	info.Rectangle(4)
	info.Color = pixel.RGB(1, 0, 0)
	hval := Map(float64(pi.player.hp), 0, float64(MaxHealth), 0, 150)
	info.Push(
		getRectangleVecs(topRigthInfoPos.Add(pixel.V(0, 0)), pixel.V(hval, 20))...,
	)
	info.Rectangle(0)
	info.Color = pixel.RGB(0, 0, 1)
	mval := Map(float64(pi.player.mp), 0, float64(MaxMana), 0, 150)
	info.Push(
		getRectangleVecs(topRigthInfoPos.Add(pixel.V(0, -30)), pixel.V(mval, 20))...,
	)
	info.Rectangle(0)

	//zoom toggle
	zoomTogglePos := topRigthInfoPos.Add(pixel.V(140, -80))
	info.Color = color.RGBA{0, 10, 0, 170}
	info.Push(
		getRectangleVecs(zoomTogglePos.Add(pixel.V(0, 0)), pixel.V(18, 34))...,
	)
	info.Rectangle(0)
	info.Push(
		getRectangleVecs(zoomTogglePos.Add(pixel.V(2, 2)), pixel.V(14, 14))...,
	)
	info.Rectangle(0)
	info.Push(
		getRectangleVecs(zoomTogglePos.Add(pixel.V(2, 18)), pixel.V(14, 14))...,
	)
	info.Rectangle(0)

	pi.hudText[ZoomINButton].Text.Color = colornames.White
	pi.hudText[ZoomOUTButton].Text.Color = colornames.White
	if Zoom == 2 {
		info.Color = color.RGBA{233, 233, 233, 133}
		info.Push(
			getRectangleVecs(zoomTogglePos.Add(pixel.V(2, 2)), pixel.V(14, 14))...,
		)
		info.Rectangle(0)
		pi.hudText[ZoomINButton].Text.Color = colornames.Black
	}
	if Zoom == 1 {
		info.Color = color.RGBA{233, 233, 233, 133}
		info.Push(
			getRectangleVecs(zoomTogglePos.Add(pixel.V(2, 18)), pixel.V(14, 14))...,
		)
		info.Rectangle(0)
		pi.hudText[ZoomOUTButton].Text.Color = colornames.Black
	}

	// Habilities info
	colorTransparent := color.RGBA{200, 10, 25, 90}
	info.Color = colorTransparent
	rectSize := pixel.V(30, 30)
	separation := 40.0
	leftBottomInfoPos := cam.Unproject(pixel.ZV)
	icon1pos := leftBottomInfoPos.Add(pixel.V(20, 30))
	info.Push(
		getRectangleVecs(icon1pos, rectSize)...,
	)
	info.Rectangle(0)
	icon2pos := icon1pos.Add(pixel.V(separation, 0))
	info.Push(
		getRectangleVecs(icon2pos, rectSize)...,
	)
	info.Rectangle(0)
	icon3pos := icon2pos.Add(pixel.V(separation, 0))
	info.Push(
		getRectangleVecs(icon3pos, rectSize)...,
	)
	info.Rectangle(0)
	icon4pos := icon3pos.Add(pixel.V(separation, 0))
	info.Push(
		getRectangleVecs(icon4pos, rectSize)...,
	)
	info.Rectangle(0)
	icon5pos := icon4pos.Add(pixel.V(separation, 0))
	info.Push(
		getRectangleVecs(icon5pos, rectSize)...,
	)
	info.Rectangle(0)
	info.Color = color.RGBA{40, 210, 88, 120}
	mainSpellsCooldown := pixel.V(30, Map(time.Since(pi.player.lastCast).Seconds(), 0, wizard.Intervals[0], 0, 30))
	info.Push(
		getRectangleVecs(icon1pos, mainSpellsCooldown)...,
	)
	info.Rectangle(0)
	info.Push(
		getRectangleVecs(icon2pos, mainSpellsCooldown)...,
	)
	info.Rectangle(0)
	info.Push(
		getRectangleVecs(icon3pos, mainSpellsCooldown)...,
	)
	info.Rectangle(0)

	classPrimarySpellsCooldown := pixel.V(30, Map(time.Since(pi.player.lastCastPrimary).Seconds(), 0, wizard.Intervals[1], 0, 30))
	classSecondarySpellsCooldown := pixel.V(30, Map(time.Since(pi.player.lastCastSecondary).Seconds(), 0, wizard.Intervals[2], 0, 30))
	info.Push(
		getRectangleVecs(icon4pos, classPrimarySpellsCooldown)...,
	)
	info.Rectangle(0)
	info.Push(
		getRectangleVecs(icon5pos, classSecondarySpellsCooldown)...,
	)
	info.Rectangle(0)

	info.Color = color.RGBA{20, 20, 20, 160}
	posoffset := pixel.V(-2, -2)
	sizeoffset := pixel.V(4, 4)
	switch cursor.Mode {
	case SpellCastExplo:
		info.Push(
			getRectangleVecs(icon1pos.Add(posoffset), rectSize.Add(sizeoffset))...,
		)
		info.Rectangle(0)
	case SpellCastApoca:
		info.Push(
			getRectangleVecs(icon2pos.Add(posoffset), rectSize.Add(sizeoffset))...,
		)
		info.Rectangle(0)
	case SpellCastDesca:
		info.Push(
			getRectangleVecs(icon3pos.Add(posoffset), rectSize.Add(sizeoffset))...,
		)
		info.Rectangle(0)
	}

	info.Draw(win)

	pi.skillIcons[ExploIcon].Sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 0.3).Moved(icon1pos.Add(pixel.V(14, 14))))
	pi.skillIcons[ApocaIcon].Sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 0.2).Moved(icon2pos.Add(pixel.V(15, 12))))
	pi.skillIcons[DescaIcon].Sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 0.2).Moved(icon3pos.Add(pixel.V(15, 14))))
	switch wizard.Type {
	case DarkWizard:
		pi.skillIcons[3].Sprite.Draw(win, pixel.IM.Moved(icon4pos.Add(pixel.V(14, 15))))
		pi.skillIcons[4].Sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 0.25).Moved(icon5pos.Add(pixel.V(15, 14))))
	case Sniper:
		pi.skillIcons[3].Sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 0.45).Moved(icon4pos.Add(pixel.V(14, 15))))
		pi.skillIcons[4].Sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 0.25).Moved(icon5pos.Add(pixel.V(15, 14))))
	case Hunter:
		pi.skillIcons[3].Sprite.Draw(win, pixel.IM.Moved(icon4pos.Add(pixel.V(14, 15))))
		pi.skillIcons[4].Sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 0.45).Moved(icon5pos.Add(pixel.V(15, 14))))
	case Timewreker:
		pi.skillIcons[3].Sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 0.45).Moved(icon4pos.Add(pixel.V(14, 15))))
		pi.skillIcons[4].Sprite.Draw(win, pixel.IM.Moved(icon5pos.Add(pixel.V(15, 15))))
	case Monk:
		pi.skillIcons[3].Sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 0.4).Moved(icon4pos.Add(pixel.V(14, 15))))
		pi.skillIcons[4].Sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 0.25).Moved(icon5pos.Add(pixel.V(15, 14))))
	case Shaman:
		pi.skillIcons[3].Sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 0.4).Moved(icon4pos.Add(pixel.V(14, 15))))
		pi.skillIcons[4].Sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 0.25).Moved(icon5pos.Add(pixel.V(15, 14))))
	}

	pi.hudText[HealthNumber].Draw(win, pixel.IM.Moved(topRigthInfoPos.Add(pixel.V(46, 6))), "%v/%v", int(pi.player.hp), int(MaxHealth))
	pi.hudText[ManaNumber].Draw(win, pixel.IM.Moved(topRigthInfoPos.Add(pixel.V(40, -25))), "%v/%v", int(pi.player.mp), int(MaxMana))
	topLeftInfoPos := cam.Unproject(pixel.V(30, winSize.Y-50))
	pi.hudText[OnlineCount].Draw(win, pixel.IM.Moved(topLeftInfoPos).Scaled(topLeftInfoPos, 2), "Online: %v", pi.playersData.Online+1)
	pi.hudText[FPSCount].Draw(win, pixel.IM.Moved(topLeftInfoPos.Add(pixel.V(0, -20))), "FPS: %v", pi.nfps)
	pi.hudText[PosXY].Draw(win, pixel.IM.Moved(topLeftInfoPos.Add(pixel.V(0, -40))), "X: %v\nY: %v", int(pi.player.pos.X/10), int(pi.player.pos.Y/10))
	pi.hudText[ZoomINButton].Draw(win, pixel.IM.Moved(zoomTogglePos.Add(pixel.V(3, 5))), "zi")
	pi.hudText[ZoomOUTButton].Draw(win, pixel.IM.Moved(zoomTogglePos.Add(pixel.V(3, 22))), "zo")

	pi.hudText[KDCount].Draw(win, pixel.IM.Moved(topRigthInfoPos.Add(pixel.V(-80, 10))), "K/D: %v/%v", pi.player.kills, pi.player.deaths)

	pi.hudText[TypingMark].Text.Clear()
	if pi.player.chat.chatting {
		pi.hudText[TypingMark].Draw(win, pixel.IM.Moved(topRigthInfoPos.Add(pixel.V(-80, -10))), "Typing...")
	}

	// Draw tab ranking status
	if win.Pressed(pixelgl.KeyTab) {
		rankingInfo := imdraw.New(nil)
		rankingInfo.Color = color.RGBA{5, 10, 30, 70}
		rankingInfo.EndShape = imdraw.SharpEndShape
		centerBasedPos := cam.Unproject(win.Bounds().Center())
		rankingInfo.Push(
			getRectangleVecs(centerBasedPos.Add(pixel.V(-150, -150)), pixel.V(300, 300))...,
		)
		rankingInfo.Rectangle(0)
		rankingInfo.Draw(win)
		rankLen := len(Ranking)
		myTop := Ranking10
		if rankLen < 10 {
			myTop = HudComponent(rankLen - 1 + int(Ranking1))
		}
		c := 1.0
		topLeftRankingPos := centerBasedPos.Add(pixel.V(-125, 140))
		pi.hudText[RankingTitle].Draw(win, pixel.IM.Moved(topLeftRankingPos.Add(pixel.V(80, -5))), pi.hudText[RankingTitle].SText)
		for i := Ranking1; i <= myTop; i++ {
			pi.hudText[i].Draw(win, pixel.IM.Moved(topLeftRankingPos.Add(pixel.V(0, -c*25))), "%v   | %v| %v|", PadRight(fmt.Sprintf("%v: %v", i-Ranking1+1, strings.TrimSpace(Ranking[i-Ranking1].Name)), " ", 23), PadRight(fmt.Sprint(Ranking[i-Ranking1].K), " ", 4), PadRight(fmt.Sprint(Ranking[i-Ranking1].D), " ", 4))
			c++
		}

	}

	// Zoom Button
	if win.JustPressed(pixelgl.MouseButtonLeft) {
		ix, iy := zoomTogglePos.Add(pixel.V(9, 9)).XY()
		ox, oy := zoomTogglePos.Add(pixel.V(9, 24)).XY()
		mx, my := cam.Unproject(win.MousePosition()).XY()
		if mx < ix+7 && mx > ix-7 && my < iy+7 && my > iy-7 {
			Zoom = 2
		}
		if mx < ox+7 && mx > ox-7 && my < oy+7 && my > oy-7 {
			Zoom = 1
		}
	}
}

func PadRight(str, pad string, lenght int) string {
	for {
		str += pad
		if len(str) > lenght {
			return str[0:lenght]
		}
	}
}
func PadLeft(str, pad string, lenght int) string {
	for {
		str = pad + str
		if len(str) > lenght {
			return str[0:lenght]
		}
	}
}
