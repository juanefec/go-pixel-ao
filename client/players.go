package main

import (
	"sync"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/segmentio/ksuid"
)

type PlayersData struct {
	Online            int
	SkinsLen          int
	DeadSkinIndex     int
	Skins             Skins
	CurrentAnimations map[ksuid.KSUID]*Player
	AnimationsMutex   *sync.RWMutex
}

type SkinType int

const (
	TuniDruida = iota
	RedBody
	BlueBody
	DarkMasterBody
	BlueArmorBody
	TwilightBody
	GodBody
	Head
	CoolHat
	Staff
	Phantom
	PhantomHead
)

type Skins []*Skin

func (s Skins) BatchClear() {
	for i := range s {
		s[i].Batch.Clear()
	}
}
func (s Skins) DrawToBatch(p *Player, pl *Player) {
	p.Update(pl)
	if !p.dead && !p.invisible {
		p.body.Draw(s[p.bodySkin].Batch, p.bodyMatrix)
		p.bacu.Draw(s[p.staffSkin].Batch, p.bodyMatrix)
		p.head.Draw(s[p.headSkin].Batch, p.headMatrix)
		p.hat.Draw(s[p.hatSkin].Batch, p.hatMatrix)
	} else if p.dead {
		p.body.Draw(s[Phantom].Batch, p.bodyMatrix)
		p.head.Draw(s[PhantomHead].Batch, p.headMatrix)
		p.invisible = false
	}
}

func (s Skins) Draw(win *pixelgl.Window) {
	for i := range s {
		s[i].Batch.Draw(win)
	}
}

func (s Skins) Load(imagPath string, t SkinType, w, h, qw, qh float64) {
	sheet := Pictures[imagPath]
	skin := &Skin{
		Pic:    &sheet,
		Batch:  pixel.NewBatch(&pixel.TrianglesData{}, sheet),
		Frames: getFrames(sheet, w, h, qw, qh),
	}
	s[t] = skin
}

type Skin struct {
	Frames []pixel.Rect
	Pic    *pixel.Picture
	Batch  *pixel.Batch
}

func NewPlayersData() PlayersData {
	pd := PlayersData{}
	pd.SkinsLen = 12
	pd.Skins = make([]*Skin, pd.SkinsLen)

	pd.Skins.Load("./images/bodydruida.png", TuniDruida, 25, 45, 6, 4)
	pd.Skins.Load("./images/redBody.png", RedBody, 25, 45, 6, 4)
	pd.Skins.Load("./images/blueBody.png", BlueBody, 25, 45, 6, 4)
	pd.Skins.Load("./images/darkopshit.png", DarkMasterBody, 25, 45, 6, 4)
	pd.Skins.Load("./images/placaazul.png", BlueArmorBody, 25, 45, 6, 4)
	pd.Skins.Load("./images/penumbras.png", TwilightBody, 25, 45, 6, 4)
	pd.Skins.Load("./images/creagod.png", GodBody, 25, 45, 6, 4)
	pd.Skins.Load("./images/heads.png", Head, 16, 16, 4, 0)
	pd.Skins.Load("./images/hatpro.png", CoolHat, 25, 32, 4, 0)
	pd.Skins.Load("./images/staff.png", Staff, 25, 45, 6, 4)
	pd.Skins.Load("./images/dead.png", Phantom, 25, 29, 3, 4)
	pd.Skins.Load("./images/deadHead.png", PhantomHead, 16, 16, 4, 0)

	pd.CurrentAnimations = map[ksuid.KSUID]*Player{}
	pd.AnimationsMutex = &sync.RWMutex{}
	pd.Online = 0
	return pd
}

func (pd *PlayersData) Draw(win *pixelgl.Window, pl *Player) {
	pd.Skins.BatchClear()
	pd.AnimationsMutex.RLock()
	for _, p := range pd.CurrentAnimations {
		pd.AnimationsMutex.RUnlock()
		pd.Skins.DrawToBatch(p, pl)
		if !p.invisible {
			p.name.Draw(win, p.nameMatrix.Moved(pixel.V(0, 4)))
			p.chat.Draw(win, p.bounds.Pos)
			p.DrawHealthMana(win)
		}
		pd.AnimationsMutex.RLock()
		//player.name.Draw(win, player.nameMatrix)
	}
	pd.AnimationsMutex.RUnlock()
	pd.Skins.Draw(win)

}
