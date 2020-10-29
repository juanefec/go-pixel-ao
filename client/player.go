package main

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/juanefec/go-pixel-ao/client/socket"
	"github.com/juanefec/go-pixel-ao/models"
	"golang.org/x/image/colornames"
)

type Player struct {
	bodyFrames, headFrames, deadFrames, deadHeadFrames, bacuFrames, hatFrames []pixel.Rect
	headPic, bodyPic, deadPic, deadHeadPic, bacuPic, hatPic                   *pixel.Picture
	cam, headMatrix, bodyMatrix, nameMatrix, hatMatrix                        pixel.Matrix
	bodyFrame, headFrame, bacuFrame, hatFrame                                 pixel.Rect
	bodySkin, headSkin, hatSkin, staffSkin                                    SkinType
	head, body, bacu, hat                                                     *pixel.Sprite
	hp, mp, maxhp, maxmp                                                      float64 // health/mana points
	wizard                                                                    *Wizard
	chat                                                                      Chat
	pos                                                                       pixel.Vec
	name                                                                      *text.Text
	sname                                                                     string
	playerUpdate                                                              *models.PlayerMsg
	dir                                                                       string
	moving                                                                    bool
	bodyStep                                                                  float64
	lastDeadFrame                                                             time.Time
	lastBodyFrame                                                             time.Time
	lastDrank                                                                 time.Time
	lastCast                                                                  time.Time
	lastCastPrimary                                                           time.Time
	lastCastSecondary                                                         time.Time
	inviEffectOut                                                             time.Time
	lastRootedStart                                                           time.Time
	drinkingManaPotions                                                       bool
	drinkingHealthPotions                                                     bool
	dead                                                                      bool
	invisible                                                                 bool
	rooted                                                                    bool
	kills, deaths                                                             int
	playerMovementSpeed                                                       float64
	colliding                                                                 bool
	collitionDir                                                              string
}

func NewPlayer(name string, wizard *Wizard) Player {
	p := &Player{}
	p.sname = name

	p.name = text.New(pixel.V(-28, 0), basicAtlas)

	p.chat = Chat{
		p:          p,
		chatlog:    chatlog,
		msgTimeout: time.Now(),
		sent:       text.New(pixel.V(24, 0), basicAtlas),
		writing:    text.New(pixel.V(24, 0), basicAtlas),
		ssent:      "",
		swriting:   "",
	}
	p.chat.sent.Color = colornames.White
	p.chat.writing.Color = colornames.Burlywood

	headSheet := Pictures["./images/heads.png"]
	headFrames := getFrames(headSheet, 16, 16, 4, 0)

	bacuSheet := Pictures["./images/staff.png"]
	bacuFrames := getFrames(bacuSheet, 25, 45, 6, 4)

	hatSheet := Pictures["./images/hatpro.png"]
	hatFrames := getFrames(hatSheet, 25, 32, 4, 0)

	deadSheet := Pictures["./images/dead.png"]
	deadFrames := getFrames(deadSheet, 25, 29, 3, 4)

	deadHeadSheet := Pictures["./images/deadHead.png"]
	deadHeadFrames := getFrames(deadHeadSheet, 16, 16, 4, 0)

	p.bodySkin = wizard.Skin
	p.headSkin = Head
	p.hatSkin = CoolHat
	p.staffSkin = Staff

	var bodySheet pixel.Picture
	var bodyFrames []pixel.Rect
	switch p.bodySkin {
	case TuniDruida:
		bodySheet = Pictures["./images/bodydruida.png"]
		p.name.Color = colornames.Whitesmoke
	case RedBody:
		bodySheet = Pictures["./images/redBody.png"]
		p.name.Color = colornames.Red
	case BlueBody:
		bodySheet = Pictures["./images/blueBody.png"]
		p.name.Color = colornames.Cyan
	case DarkMasterBody:
		bodySheet = Pictures["./images/darkopshit.png"]
		p.name.Color = colornames.Black
	case BlueArmorBody:
		bodySheet = Pictures["./images/placaazul.png"]
		p.name.Color = colornames.Blue
	case TwilightBody:
		bodySheet = Pictures["./images/penumbras.png"]
		p.name.Color = colornames.Darkgoldenrod

	case GodBody:
		bodySheet = Pictures["./images/creagod.png"]
		p.name.Color = colornames.Turquoise

	}

	p.wizard = wizard

	fmt.Fprintln(p.name, name)
	bodyFrames = getFrames(bodySheet, 25, 45, 6, 4)

	p.playerUpdate = &models.PlayerMsg{}
	p.lastBodyFrame = time.Now()
	p.lastDeadFrame = time.Now()
	p.lastDrank = time.Now()
	p.lastCast = time.Now().Add(-time.Second * 2)
	p.lastCastPrimary = time.Now().Add(-time.Second * 20)
	p.lastCastSecondary = time.Now().Add(-time.Second * 20)
	p.inviEffectOut = time.Now().Add(-time.Second * 2)

	p.invisible = false
	p.headFrames = headFrames
	p.bacuFrames = bacuFrames
	p.hatFrames = hatFrames
	p.bodyFrames = bodyFrames
	p.bodyPic = &bodySheet
	p.headPic = &headSheet
	p.bacuPic = &bacuSheet
	p.hatPic = &hatSheet
	p.deadFrames = deadFrames
	p.deadHeadFrames = deadHeadFrames
	p.deadPic = &deadSheet
	p.deadHeadPic = &deadHeadSheet
	p.dir = "down"
	p.pos = pixel.V(2000, 2600)
	p.maxmp = MaxMana
	p.maxhp = MaxHealth
	if name == "   creagod   " {
		p.maxmp = MaxMana * 4
		p.maxhp = MaxHealth * 4
	}
	p.mp = p.maxmp
	p.hp = p.maxhp
	p.playerMovementSpeed = PlayerBaseSpeed
	return *p
}

func (p *Player) DrawHealthMana(win *pixelgl.Window) {
	infoPos := p.pos.Add(pixel.V(-16, -24))
	info := imdraw.New(nil)
	info.Color = colornames.Black
	info.EndShape = imdraw.SharpEndShape
	info.Push(
		infoPos.Add(pixel.V(0, 0)),
		infoPos.Add(pixel.V(32, 0)),
		infoPos.Add(pixel.V(0, -2)),
		infoPos.Add(pixel.V(32, -2)),
	)
	info.Rectangle(2)

	info.Color = pixel.RGB(1, 0, 0)
	hval := Map(float64(p.hp), 0, float64(p.maxhp), 0, 32)
	info.Push(
		infoPos.Add(pixel.V(0, 0)),
		infoPos.Add(pixel.V(hval, 0)),
		infoPos.Add(pixel.V(0, -2)),
		infoPos.Add(pixel.V(hval, -2)),
	)
	info.Rectangle(0)

	info.Draw(win)
}

func (p *Player) DrawIngameHud(win *pixelgl.Window) {
	arrowChargeBar := p.cam.Unproject(win.Bounds().Center()).Add(pixel.V(-16, 38))
	info := imdraw.New(nil)
	healthBar := arrowChargeBar.Add(pixel.V(0, -61))
	info.Color = colornames.Black
	info.EndShape = imdraw.SharpEndShape
	info.Push(
		healthBar.Add(pixel.V(0, 0)),
		healthBar.Add(pixel.V(32, 0)),
		healthBar.Add(pixel.V(0, -2)),
		healthBar.Add(pixel.V(32, -2)),
	)
	info.Rectangle(2)

	info.Color = pixel.RGB(1, 0, 0)
	hval := Map(float64(p.hp), 0, float64(p.maxhp), 0, 32)
	info.Push(
		healthBar.Add(pixel.V(0, 0)),
		healthBar.Add(pixel.V(hval, 0)),
		healthBar.Add(pixel.V(0, -2)),
		healthBar.Add(pixel.V(hval, -2)),
	)
	info.Rectangle(0)
	manaBar := healthBar.Add(pixel.V(0, -5))
	info.Color = colornames.Black
	info.EndShape = imdraw.SharpEndShape
	info.Push(
		manaBar.Add(pixel.V(0, 0)),
		manaBar.Add(pixel.V(32, 0)),
		manaBar.Add(pixel.V(0, -2)),
		manaBar.Add(pixel.V(32, -2)),
	)
	info.Rectangle(2)

	info.Color = pixel.RGB(0, 0, 1)
	mval := Map(float64(p.mp), 0, float64(p.maxmp), 0, 32)
	info.Push(
		manaBar.Add(pixel.V(0, 0)),
		manaBar.Add(pixel.V(mval, 0)),
		manaBar.Add(pixel.V(0, -2)),
		manaBar.Add(pixel.V(mval, -2)),
	)
	info.Rectangle(0)

	info.Draw(win)

}

func (p *Player) OnMe(click pixel.Vec) bool {
	r := click.X < p.pos.X+14 && click.X > p.pos.X-14 && click.Y < p.pos.Y+30 && click.Y > p.pos.Y-20
	return r
}

func (p *Player) CollidingCheck(pp pixel.Vec) {
	// tloffset := pixel.V(-14, 30)
	// tlp := p.pos.Add(tloffset) // rects ar 28x50
	// tlpp := pp.Add(tloffset)
	//r := (tlp.X+28) >= tlpp.X && tlp.X <= (tlpp.X+28) && (tlp.Y-50) >= tlpp.Y && tlp.Y <= (tlpp.Y-50)
	if p.moving {
		r := pp.X-12 < p.pos.X+12 && pp.X+12 > p.pos.X-12 && pp.Y-12 < p.pos.Y+12 && pp.Y+12 > p.pos.Y-12
		if r {
			switch true {
			case (pp.X+12 > p.pos.X-12) && p.dir == "left":
				maxDepth := (pp.X + 12) - (p.pos.X - 12)
				p.pos.X += maxDepth
				p.collitionDir = p.dir

			case (pp.X-12 < p.pos.X+12 && p.dir == "right"):
				maxDepth := (p.pos.X + 12) - (pp.X - 12)
				p.pos.X -= maxDepth
				p.collitionDir = p.dir

			case (pp.Y+12 > p.pos.Y-12) && p.dir == "down":
				maxDepth := (pp.Y + 12) - (p.pos.Y - 12)
				p.pos.Y += maxDepth
				p.collitionDir = p.dir

			case (pp.Y-12 < p.pos.Y+12) && p.dir == "up":
				maxDepth := (p.pos.Y + 12) - (pp.Y - 12)
				p.pos.Y -= maxDepth
				p.collitionDir = p.dir

			}
		} else {
			p.collitionDir = ""
		}
	}
}

func (p *Player) OnTrap(click pixel.Vec) bool {
	r := click.X < p.pos.X+12 && click.X > p.pos.X-12 && click.Y < p.pos.Y+5 && click.Y > p.pos.Y-20
	return r
}

func (p *Player) InsideRaduis(center pixel.Vec, r float64) bool {
	return math.Abs(Dist(center, p.pos)) <= r
}

func (p *Player) clientUpdate(s *socket.Socket) {
	p.playerUpdate = &models.PlayerMsg{
		ID:        s.ClientID,
		Name:      p.sname,
		Skin:      int(p.bodySkin),
		HP:        p.hp,
		X:         p.pos.X,
		Y:         p.pos.Y,
		Dir:       p.dir,
		Moving:    p.moving,
		Dead:      p.dead,
		Invisible: p.invisible,
	}
	playerMsg, err := json.Marshal(p.playerUpdate)
	if err != nil {
		return
	}
	p.playerUpdate = &models.PlayerMsg{}
	s.O <- models.NewMesg(models.UpdateServer, playerMsg)

}

func (p *Player) Update(pl *Player) {
	if !p.dead {
		if pl != nil {
			pl.CollidingCheck(p.pos)
		}
		switch p.dir {
		case "up":
			p.headFrame = p.headFrames[3]
			p.hatFrame = p.hatFrames[3]
			p.bodyFrame = p.getNextBodyFrame(BodyUp, p.bodyFrames)
			p.bacuFrame = p.getNextBodyFrame(BodyUp, p.bacuFrames)
		case "down":
			p.headFrame = p.headFrames[0]
			p.hatFrame = p.hatFrames[0]
			p.bodyFrame = p.getNextBodyFrame(BodyDown, p.bodyFrames)
			p.bacuFrame = p.getNextBodyFrame(BodyDown, p.bacuFrames)
		case "left":
			p.headFrame = p.headFrames[2]
			p.hatFrame = p.hatFrames[2]
			p.bodyFrame = p.getNextBodyFrame(BodyLeft, p.bodyFrames)
			p.bacuFrame = p.getNextBodyFrame(BodyLeft, p.bacuFrames)
		case "right":
			p.headFrame = p.headFrames[1]
			p.hatFrame = p.hatFrames[1]
			p.bodyFrame = p.getNextBodyFrame(BodyRight, p.bodyFrames)
			p.bacuFrame = p.getNextBodyFrame(BodyRight, p.bacuFrames)
		default:
			p.headFrame = p.headFrames[0]
			p.hatFrame = p.hatFrames[0]
			p.bodyFrame = p.getNextBodyFrame(BodyDown, p.bodyFrames)
			p.bacuFrame = p.getNextBodyFrame(BodyDown, p.bacuFrames)
		}
		p.headMatrix = pixel.IM.Moved(p.pos.Add(pixel.V(1, 22)))
		p.bodyMatrix = pixel.IM.Moved(p.pos.Add(pixel.V(0, 0)))
		p.hatMatrix = pixel.IM.Moved(p.pos.Add(pixel.V(1, 21)))
		p.nameMatrix = pixel.IM.Moved(p.pos.Sub(p.name.Bounds().Center().Floor()).Add(pixel.V(0, -37)))
		p.head = pixel.NewSprite(*p.headPic, p.headFrame)
		p.body = pixel.NewSprite(*p.bodyPic, p.bodyFrame)
		p.bacu = pixel.NewSprite(*p.bacuPic, p.bacuFrame)
		p.hat = pixel.NewSprite(*p.hatPic, p.hatFrame)
		dt := time.Since(p.lastDrank).Seconds()
		second := time.Second.Seconds()
		if p.drinkingHealthPotions && !p.drinkingManaPotions {
			if dt > second/3.3 {
				p.hp += 30
				if p.hp > p.maxhp {
					p.hp = p.maxhp
				}
				p.lastDrank = time.Now()
			}
		}
		if p.drinkingManaPotions && !p.drinkingHealthPotions {
			if dt > second/4 {
				p.mp += p.maxmp * 0.05
				if p.mp > p.maxmp {
					p.mp = p.maxmp
				}
				p.lastDrank = time.Now()
			}
		}
	} else if p.dead {
		switch p.dir {
		case "up":
			p.headFrame = p.headFrames[3]
			p.bodyFrame = p.getNextDeadFrame(DeadUp)
		case "down":
			p.headFrame = p.headFrames[0]
			p.bodyFrame = p.getNextDeadFrame(DeadDown)
		case "left":
			p.headFrame = p.headFrames[2]
			p.bodyFrame = p.getNextDeadFrame(DeadLeft)
		case "right":
			p.headFrame = p.headFrames[1]
			p.bodyFrame = p.getNextDeadFrame(DeadRight)
		default:
			p.headFrame = p.headFrames[0]
			p.bodyFrame = p.getNextDeadFrame(DeadDown)
		}
		p.headMatrix = pixel.IM.Moved(p.pos.Add(pixel.V(1, 20)))
		p.bodyMatrix = pixel.IM.Moved(p.pos.Add(pixel.V(0, 0)))
		p.nameMatrix = pixel.IM.Moved(p.pos.Sub(p.name.Bounds().Center().Floor()).Add(pixel.V(0, -37)))
		p.head = pixel.NewSprite(*p.deadHeadPic, p.headFrame)
		p.body = pixel.NewSprite(*p.deadPic, p.bodyFrame)
		p.rooted = false
	}
	if time.Since(p.lastRootedStart).Seconds() > 0 {
		p.rooted = false
	}
}

func (p *Player) CheckColitions(pp *Player) {

}

func (p *Player) Draw(win *pixelgl.Window, s *socket.Socket) {
	p.Update(nil)
	if win.JustPressed(pixelgl.KeyEnter) {
		p.chat.chatting = !p.chat.chatting
		if !p.chat.chatting {
			p.chat.Send(s)
		}
	}
	if p.chat.chatting {
		p.chat.Write(win)
	}
	p.chat.Draw(win, p.pos)
	p.name.Draw(win, p.nameMatrix)
	if !p.invisible {
		p.body.Draw(win, p.bodyMatrix)
		p.head.Draw(win, p.headMatrix)

		if !p.dead {
			p.bacu.Draw(win, p.bodyMatrix)
			p.hat.Draw(win, p.hatMatrix)
		}
	}
	dt := time.Since(p.inviEffectOut).Seconds()
	if dt >= time.Second.Seconds()*2 {
		p.invisible = false
	}
}

func (p *Player) getNextBodyFrame(dirFrames []int, part []pixel.Rect) pixel.Rect {
	dt := time.Since(p.lastBodyFrame).Seconds()
	p.lastBodyFrame = time.Now()
	if p.moving {
		p.bodyStep += (p.playerMovementSpeed / 15) * dt
		newFrame := int(p.bodyStep)
		if (newFrame <= 5 && (p.dir == "up" || p.dir == "down")) || (newFrame <= 4 && (p.dir == "right" || p.dir == "left")) {
			return part[dirFrames[newFrame]]
		}
	}
	p.bodyStep = 0
	return part[dirFrames[0]]
}

func (p *Player) getNextDeadFrame(dirFrames []int) pixel.Rect {
	dt := time.Since(p.lastDeadFrame).Seconds()
	p.lastDeadFrame = time.Now()
	if p.moving {
		p.bodyStep += 7 * dt
		newFrame := int(p.bodyStep)
		if (newFrame <= 2 && (p.dir == "up" || p.dir == "down")) || (newFrame <= 2 && (p.dir == "right" || p.dir == "left")) {
			return p.deadFrames[dirFrames[newFrame]]
		}
	}
	p.bodyStep = 0
	return p.deadFrames[dirFrames[0]]
}
