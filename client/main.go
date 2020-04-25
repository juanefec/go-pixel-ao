package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/juanefec/go-pixel-ao/client/socket"
	"github.com/juanefec/go-pixel-ao/models"
	"github.com/segmentio/ksuid"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

var (
	PlayerSpeed = 200.0
	Zoom        = 1.0
	ZoomSpeed   = 1.2
	fps         = 0
	second      = time.Tick(time.Second)
	MaxMana     = 3444
	MaxHealth   = 366
	ApocaDmg    = 190
)
var (
	Newline      = []byte{'\n'}
	TreeQuantity = 1000
	BodyUp       = []int{12, 13, 14, 15, 16, 17}
	BodyDown     = []int{18, 19, 20, 21, 22, 23}
	BodyLeft     = []int{6, 7, 8, 9, 10}
	BodyRight    = []int{0, 1, 2, 3, 4}
	DeadUp       = []int{6, 7, 8}
	DeadDown     = []int{9, 10, 11}
	DeadLeft     = []int{3, 4, 5}
	DeadRight    = []int{0, 1, 2}

	ApocaFrames = []int{12, 13, 14, 15, 8, 9, 10, 11, 4, 5, 6, 7, 0, 1, 2, 3}

	Pictures map[string]pixel.Picture
)

const (
	message       = "Ping"
	StopCharacter = "\r\n\r\n"
)

//message order [id;name;playerX;playerY;dir;moving]

func run() {
	Pictures = loadPictures(
		"./images/apocas.png",
		"./images/bodies.png",
		"./images/heads.png",
		"./images/trees.png",
		"./images/dead.png",
		"./images/deadHead.png",
		"./images/desca.png",
		"./images/curaBody.png",
		"./images/curaHead.png",
	)
	name, err := SetNameWindow()
	if err != nil {
		log.Panic(err)
	}
	player := NewPlayer(name)
	apocaData := NewSpellData("apoca", &player)
	descaData := NewSpellData("desca", &player)
	forest := NewForest()
	otherPlayers := NewPlayersData()
	playerInfo := NewPlayerInfo(&player, &otherPlayers)
	resu := NewResu()

	socket := socket.NewSocket("25.9.191.221", 3333)
	defer socket.Close()

	cfg := pixelgl.WindowConfig{
		Title:  "Creative AO",
		Bounds: pixel.R(0, 0, 600, 600),
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	cursor := NewCursor(win)
	go keyInputs(win, &player, cursor)
	go GameUpdate(socket, &otherPlayers, &player, apocaData, descaData)

	for !win.Closed() {
		apocaData.Batch.Clear()
		descaData.Batch.Clear()

		cam := pixel.IM.Scaled(player.pos, Zoom).Moved(win.Bounds().Center().Sub(player.pos))
		win.SetMatrix(cam)
		win.Clear(colornames.Forestgreen)

		apocaData.Update(win, cam, socket, &otherPlayers, cursor)
		descaData.Update(win, cam, socket, &otherPlayers, cursor)
		player.Update()
		Zoom *= math.Pow(ZoomSpeed, win.MouseScroll().Y)
		//forest.GrassBatch.Draw(win)
		resu.Draw(win, cam, &player)
		otherPlayers.Draw(win)
		player.body.Draw(win, player.bodyMatrix)
		player.head.Draw(win, player.headMatrix)
		player.name.Draw(win, player.nameMatrix)
		forest.Batch.Draw(win)
		apocaData.Batch.Draw(win)
		descaData.Batch.Draw(win)
		playerInfo.DrawPlayerInfo(win)
		cursor.Draw(cam)
		fps++

		select {
		case <-second:

			win.SetTitle(fmt.Sprintf("%s | FPS: %d", cfg.Title, fps))
			fps = 0
		default:
		}
		win.Update()
		player.clientUpdate(socket)
	}
}

func main() {
	pixelgl.Run(run)
}

type CursorMode int

const (
	Normal CursorMode = iota
	SpellCastDesca
	SpellCastApoca
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
		c.win.SetCursorVisible(false)
		cross := imdraw.New(nil)
		cross.Color = colornames.Black
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

type PlayerInfo struct {
	playersData                    *PlayersData
	player                         *Player
	hdisplay, mdisplay, onsdisplay *text.Text
}

func NewPlayerInfo(player *Player, pd *PlayersData) *PlayerInfo {
	pi := PlayerInfo{}
	basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	pi.hdisplay = text.New(pixel.ZV, basicAtlas)
	pi.hdisplay.Color = colornames.Whitesmoke
	fmt.Fprintf(pi.hdisplay, "%v/%v", player.hp, MaxHealth)
	pi.mdisplay = text.New(pixel.ZV, basicAtlas)
	pi.mdisplay.Color = colornames.Whitesmoke
	fmt.Fprintf(pi.mdisplay, "%v/%v", player.mp, MaxMana)
	pi.onsdisplay = text.New(pixel.ZV, basicAtlas)
	pi.onsdisplay.Color = colornames.Whitesmoke
	fmt.Fprintf(pi.onsdisplay, "Online: %v", pd.Online+1)
	pi.player = player
	pi.playersData = pd
	return &pi
}

func (pi *PlayerInfo) DrawPlayerInfo(win *pixelgl.Window) {
	info := imdraw.New(nil)
	info.Color = colornames.Black
	info.EndShape = imdraw.SharpEndShape
	info.Push(
		pi.player.pos.Add(pixel.V(138, 292)),
		pi.player.pos.Add(pixel.V(292, 292)),
		pi.player.pos.Add(pixel.V(292, 268)),
		pi.player.pos.Add(pixel.V(138, 268)),
		pi.player.pos.Add(pixel.V(138, 292)),
	)
	info.Line(4)
	info.Push(
		pi.player.pos.Add(pixel.V(138, 252)),
		pi.player.pos.Add(pixel.V(292, 252)),
		pi.player.pos.Add(pixel.V(292, 228)),
		pi.player.pos.Add(pixel.V(138, 228)),
		pi.player.pos.Add(pixel.V(138, 252)),
	)
	info.Line(4)
	info.Color = pixel.RGB(1, 0, 0).Mul(pixel.Alpha(20))
	hval := Map(float64(pi.player.hp), 0, 366, 140, 290)
	info.Push(
		pi.player.pos.Add(pixel.V(140, 290)),
		pi.player.pos.Add(pixel.V(hval, 290)),
		pi.player.pos.Add(pixel.V(140, 270)),
		pi.player.pos.Add(pixel.V(hval, 270)),
	)
	info.Rectangle(0)
	info.Color = pixel.RGB(0, 0, 1).Mul(pixel.Alpha(20))
	mval := Map(float64(pi.player.mp), 0, 3444, 140, 290)
	info.Push(
		pi.player.pos.Add(pixel.V(140, 250)),
		pi.player.pos.Add(pixel.V(mval, 250)),
		pi.player.pos.Add(pixel.V(140, 230)),
		pi.player.pos.Add(pixel.V(mval, 230)),
	)
	info.Rectangle(0)
	info.Draw(win)
	pi.hdisplay.Clear()
	pi.mdisplay.Clear()
	pi.onsdisplay.Clear()
	fmt.Fprintf(pi.hdisplay, "%v/%v", pi.player.hp, MaxHealth)
	fmt.Fprintf(pi.mdisplay, "%v/%v", pi.player.mp, MaxMana)
	fmt.Fprintf(pi.onsdisplay, "Online: %v", pi.playersData.Online+1)
	hmatrix := pixel.IM.Moved(pi.player.bodyMatrix.Project(pixel.V(195, 276)))
	mmatrix := pixel.IM.Moved(pi.player.bodyMatrix.Project(pixel.V(195, 236)))
	onsmatrix := pixel.IM.Moved(pi.player.bodyMatrix.Project(pixel.V(-290, 282)))
	pi.hdisplay.Draw(win, hmatrix)
	pi.mdisplay.Draw(win, mmatrix)
	pi.onsdisplay.Draw(win, onsmatrix)
}

func Map(v, s1, st1, s2, st2 float64) float64 {
	newval := (v-s1)/(st1-s1)*(st2-s2) + s2
	if newval < s2 {
		return s2
	}
	if newval > st2 {
		return st2
	}
	return newval
}

type SpellData struct {
	SpellName         string
	SpellMode         CursorMode
	ManaCost, Damage  int
	Caster            *Player
	Frames            []pixel.Rect
	Pic               *pixel.Picture
	Batch             *pixel.Batch
	CurrentAnimations []*Spell
}

func (sd *SpellData) Update(win *pixelgl.Window, cam pixel.Matrix, s *socket.Socket, pd *PlayersData, cursor *Cursor) {
	dt := time.Since(sd.Caster.lastCast).Seconds()
	if win.JustPressed(pixelgl.MouseButtonLeft) && sd.Caster.mp >= sd.ManaCost && cursor.Mode == sd.SpellMode && dt >= ((time.Second.Seconds()/3)*2) {
		sd.Caster.lastCast = time.Now()
		for key, _ := range pd.CurrentAnimations {
			mouse := cam.Unproject(win.MousePosition())
			if pd.CurrentAnimations[key].OnMe(mouse) && !pd.CurrentAnimations[key].dead {
				sd.Caster.mp -= sd.ManaCost
				newSpell := &Spell{
					spellName:   &sd.SpellName,
					step:        sd.Frames[0],
					frameNumber: 0.0,
					matrix:      &pd.CurrentAnimations[key].headMatrix,
					last:        time.Now(),
				}

				newSpell.frame = pixel.NewSprite(*(sd.Pic), newSpell.step)
				sd.CurrentAnimations = append(sd.CurrentAnimations, newSpell)
				spell := models.SpellMsg{
					ID:       s.ClientID,
					Type:     sd.SpellName,
					TargetID: key,
					Name:     "name",
					X:        mouse.X,
					Y:        mouse.Y,
				}
				paylaod, _ := json.Marshal(spell)
				s.O <- models.NewMesg(models.Spell, paylaod)

				break
			}
		}
		cursor.SetNormalMode()
	}

	for i := 0; i <= len(sd.CurrentAnimations)-1; i++ {
		next, kill := sd.CurrentAnimations[i].NextFrame(sd.Frames)
		if kill {
			if i < len(sd.CurrentAnimations)-1 {
				copy(sd.CurrentAnimations[i:], sd.CurrentAnimations[i+1:])
			}
			sd.CurrentAnimations[len(sd.CurrentAnimations)-1] = nil // or the zero sd.vCurrentAnimationslue of T
			sd.CurrentAnimations = sd.CurrentAnimations[:len(sd.CurrentAnimations)-1]
			continue
		}
		sd.CurrentAnimations[i].step = next
		sd.CurrentAnimations[i].frame = pixel.NewSprite(*sd.Pic, sd.CurrentAnimations[i].step)
		sd.CurrentAnimations[i].frame.Draw(sd.Batch, (*sd.CurrentAnimations[i].matrix))
	}
}

func NewSpellData(spell string, caster *Player) *SpellData {

	var sheet pixel.Picture
	var batch *pixel.Batch
	var frames []pixel.Rect
	var mode CursorMode
	var manaCost, damage int
	switch spell {
	case "apoca":
		sheet = Pictures["./images/apocas.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		unorderedFrames := getFrames(sheet, 145, 145, 4, 4)
		for i := range unorderedFrames {
			frames = append(frames, unorderedFrames[ApocaFrames[i]])
		}
		mode = SpellCastApoca
		manaCost = 1800
		damage = 190
		break
	case "desca":
		sheet = Pictures["./images/desca.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		frames = getFrames(sheet, 127, 127, 5, 3)
		mode = SpellCastDesca
		manaCost = 990
		damage = 90
	}

	return &SpellData{
		Caster:            caster,
		SpellName:         spell,
		Frames:            frames,
		Pic:               &sheet,
		Batch:             batch,
		SpellMode:         mode,
		ManaCost:          manaCost,
		Damage:            damage,
		CurrentAnimations: make([]*Spell, 0),
	}
}

type Spell struct {
	target      *Player
	spellName   *string
	step        pixel.Rect
	frame       *pixel.Sprite
	frameNumber float64
	matrix      *pixel.Matrix
	last        time.Time
}

func (a *Spell) NextFrame(spellFrames []pixel.Rect) (pixel.Rect, bool) {
	dt := time.Since(a.last).Seconds()
	a.last = time.Now()
	a.frameNumber += 21 * dt
	i := int(a.frameNumber)
	if i <= len(spellFrames)-1 {
		return spellFrames[i], false
	}
	a.frameNumber = .0
	return spellFrames[0], true
}

type PlayersData struct {
	Online                                             int
	BodyFrames, HeadFrames, DeadFrames, DeadHeadFrames []pixel.Rect
	BodyPic, HeadPic, DeadPic, DeadHeadPic             *pixel.Picture
	BodyBatch, HeadBatch, DeadBatch, DeadHeadBatch     *pixel.Batch
	CurrentAnimations                                  map[ksuid.KSUID]*Player
	AnimationsMutex                                    *sync.RWMutex
}

func NewPlayersData() PlayersData {
	bodySheet := Pictures["./images/bodies.png"]
	bodyBatch := pixel.NewBatch(&pixel.TrianglesData{}, bodySheet)
	bodyFrames := getFrames(bodySheet, 19, 38, 6, 4)

	headSheet := Pictures["./images/heads.png"]
	headBatch := pixel.NewBatch(&pixel.TrianglesData{}, headSheet)
	headFrames := getFrames(headSheet, 16, 16, 4, 0)

	deadSheet := Pictures["./images/dead.png"]
	deadBatch := pixel.NewBatch(&pixel.TrianglesData{}, deadSheet)
	deadFrames := getFrames(deadSheet, 25, 29, 3, 4)

	deadHeadSheet := Pictures["./images/deadHead.png"]
	deadHeadBatch := pixel.NewBatch(&pixel.TrianglesData{}, deadHeadSheet)
	deadHeadFrames := getFrames(deadHeadSheet, 16, 16, 4, 0)

	return PlayersData{
		BodyFrames:        bodyFrames,
		HeadFrames:        headFrames,
		BodyPic:           &bodySheet,
		HeadPic:           &headSheet,
		BodyBatch:         bodyBatch,
		HeadBatch:         headBatch,
		DeadFrames:        deadFrames,
		DeadHeadFrames:    deadHeadFrames,
		DeadPic:           &deadSheet,
		DeadHeadPic:       &deadHeadSheet,
		DeadBatch:         deadBatch,
		DeadHeadBatch:     deadHeadBatch,
		CurrentAnimations: map[ksuid.KSUID]*Player{},
		AnimationsMutex:   &sync.RWMutex{},
		Online:            0,
	}
}
func GameUpdate(s *socket.Socket, pd *PlayersData, p *Player, ssd ...*SpellData) {
	for {
		select {
		case data := <-s.I:
			msg := models.UnmarshallMesg(data)
			switch msg.Type {
			case models.UpdateClient:

				players := []*models.PlayerMsg{}
				json.Unmarshal(msg.Payload, &players)
				for i := 0; i <= len(players)-1; i++ {
					p := players[i]
					if p.ID != s.ClientID {
						pd.AnimationsMutex.Lock()
						player, ok := pd.CurrentAnimations[p.ID]
						if !ok {
							pd.Online++
							np := NewPlayer(p.Name)
							pd.CurrentAnimations[p.ID] = &np
							player, _ = pd.CurrentAnimations[p.ID]
						}
						pd.AnimationsMutex.Unlock()
						player.pos = pixel.V(p.X, p.Y)
						player.dir = p.Dir
						player.moving = p.Moving
						player.dead = p.Dead
					}
				}
				break
			case models.Spell:

				spell := models.SpellMsg{}
				json.Unmarshal(msg.Payload, &spell)

				target := &Player{}
				if s.ClientID == spell.TargetID {
					target = p
				} else {
					target = pd.CurrentAnimations[spell.TargetID]
				}

				newSpell := &Spell{
					spellName:   &spell.Type,
					target:      target,
					frameNumber: 0.0,
					matrix:      &target.headMatrix,
					last:        time.Now(),
				}
				for i := range ssd {
					sd := ssd[i]
					if spell.Type == sd.SpellName {
						newSpell.step = sd.Frames[0]
						newSpell.frame = pixel.NewSprite(*(sd.Pic), newSpell.step)
						sd.CurrentAnimations = append(sd.CurrentAnimations, newSpell)
						target.hp -= sd.Damage
						if target.hp < 0 {
							target.hp = 0
							target.dead = true
						}
						break
					}
				}
			case models.Disconect:
				pd.Online--
				m := models.DisconectMsg{}
				json.Unmarshal(msg.Payload, &m)
				pd.AnimationsMutex.Lock()
				delete(pd.CurrentAnimations, m.ID)
				pd.AnimationsMutex.Unlock()
			}

		}
	}
}

func (pd *PlayersData) Draw(win *pixelgl.Window) {
	pd.BodyBatch.Clear()
	pd.HeadBatch.Clear()
	pd.DeadBatch.Clear()
	pd.DeadHeadBatch.Clear()
	pd.AnimationsMutex.RLock()
	for _, p := range pd.CurrentAnimations {
		pd.AnimationsMutex.RUnlock()
		p.Update()
		if !p.dead {
			p.body.Draw(pd.BodyBatch, p.bodyMatrix)
			p.head.Draw(pd.HeadBatch, p.headMatrix)
		} else {
			p.body.Draw(pd.DeadBatch, p.bodyMatrix)
			p.head.Draw(pd.DeadHeadBatch, p.headMatrix)
		}
		p.name.Draw(win, p.nameMatrix)
		pd.AnimationsMutex.RLock()
		//player.name.Draw(win, player.nameMatrix)
	}
	pd.AnimationsMutex.RUnlock()
	pd.BodyBatch.Draw(win)
	pd.HeadBatch.Draw(win)
	pd.DeadBatch.Draw(win)
	pd.DeadHeadBatch.Draw(win)
}

type Player struct {
	pos                                                pixel.Vec
	headPic, bodyPic, deadPic, deadHeadPic             *pixel.Picture
	name                                               *text.Text
	sname                                              string
	head, body                                         *pixel.Sprite
	headMatrix, bodyMatrix, nameMatrix                 pixel.Matrix
	bodyFrames, headFrames, deadFrames, deadHeadFrames []pixel.Rect
	playerUpdate                                       *models.PlayerMsg
	dir                                                string
	moving                                             bool
	bodyFrame                                          pixel.Rect
	headFrame                                          pixel.Rect
	bodyStep                                           float64
	lastDeadFrame                                      time.Time
	lastBodyFrame                                      time.Time
	lastDrank                                          time.Time
	lastCast                                           time.Time
	hp, mp                                             int // health/mana points
	drinkingManaPotions                                bool
	drinkingHealthPotions                              bool
	dead                                               bool
}

func NewPlayer(name string) Player {
	p := &Player{}
	p.sname = name
	basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	p.name = text.New(pixel.V(-28, 0), basicAtlas)
	p.name.Color = colornames.Blue
	fmt.Fprintln(p.name, name)

	bodySheet := Pictures["./images/bodies.png"]
	bodyFrames := getFrames(bodySheet, 19, 38, 6, 4)

	headSheet := Pictures["./images/heads.png"]
	headFrames := getFrames(headSheet, 16, 16, 4, 0)

	deadSheet := Pictures["./images/dead.png"]
	deadFrames := getFrames(deadSheet, 25, 29, 3, 4)

	deadHeadSheet := Pictures["./images/deadHead.png"]
	deadHeadFrames := getFrames(deadHeadSheet, 16, 16, 4, 0)

	p.playerUpdate = &models.PlayerMsg{}
	p.lastBodyFrame = time.Now()
	p.lastDeadFrame = time.Now()
	p.lastDrank = time.Now()
	p.lastCast = time.Now()
	p.bodyFrames = bodyFrames
	p.headFrames = headFrames
	p.bodyPic = &bodySheet
	p.headPic = &headSheet
	p.deadFrames = deadFrames
	p.deadHeadFrames = deadHeadFrames
	p.deadPic = &deadSheet
	p.deadHeadPic = &deadHeadSheet
	p.dir = "down"
	p.pos = pixel.ZV
	p.mp = MaxMana
	p.hp = MaxHealth
	return *p
}

func (p *Player) OnMe(click pixel.Vec) bool {
	r := click.X < p.pos.X+18 && click.X > p.pos.X-18 && click.Y < p.pos.Y+25 && click.Y > p.pos.Y-25
	return r
}

func (p *Player) clientUpdate(s *socket.Socket) {
	p.playerUpdate = &models.PlayerMsg{
		ID:     s.ClientID,
		Name:   p.sname,
		X:      p.pos.X,
		Y:      p.pos.Y,
		Dir:    p.dir,
		Moving: p.moving,
		Dead:   p.dead,
	}
	playerMsg, err := json.Marshal(p.playerUpdate)
	if err != nil {
		return
	}
	p.playerUpdate = &models.PlayerMsg{}
	s.O <- models.NewMesg(models.UpdateServer, playerMsg)

}

func (p *Player) Update() {
	if !p.dead {
		switch p.dir {
		case "up":
			p.headFrame = p.headFrames[3]
			p.bodyFrame = p.getNextBodyFrame(BodyUp)
		case "down":
			p.headFrame = p.headFrames[0]
			p.bodyFrame = p.getNextBodyFrame(BodyDown)
		case "left":
			p.headFrame = p.headFrames[2]
			p.bodyFrame = p.getNextBodyFrame(BodyLeft)
		case "right":
			p.headFrame = p.headFrames[1]
			p.bodyFrame = p.getNextBodyFrame(BodyRight)
		default:
			p.headFrame = p.headFrames[0]
			p.bodyFrame = p.getNextBodyFrame(BodyDown)
		}
		p.headMatrix = pixel.IM.Moved(p.pos.Add(pixel.V(1, 25)))
		p.bodyMatrix = pixel.IM.Moved(p.pos.Add(pixel.V(0, 0)))
		p.nameMatrix = pixel.IM.Moved(p.pos.Add(pixel.V(0, -26)))
		p.head = pixel.NewSprite(*p.headPic, p.headFrame)
		p.body = pixel.NewSprite(*p.bodyPic, p.bodyFrame)

		dt := time.Since(p.lastDrank).Seconds()
		if p.drinkingHealthPotions && !p.drinkingManaPotions {
			if dt > time.Second.Seconds()/4 {
				p.hp += 30
				if p.hp > MaxHealth {
					p.hp = MaxHealth
				}
				p.lastDrank = time.Now()
			}
		}
		if p.drinkingManaPotions && !p.drinkingHealthPotions {
			if dt > time.Second.Seconds()/4 {
				p.mp += 172
				if p.mp > MaxMana {
					p.mp = MaxMana
				}
				p.lastDrank = time.Now()
			}
		}
	} else {
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
		p.headMatrix = pixel.IM.Moved(p.pos.Add(pixel.V(1, 25)))
		p.bodyMatrix = pixel.IM.Moved(p.pos.Add(pixel.V(0, 0)))
		p.nameMatrix = pixel.IM.Moved(p.pos.Add(pixel.V(0, -26)))
		p.head = pixel.NewSprite(*p.deadHeadPic, p.headFrame)
		p.body = pixel.NewSprite(*p.deadPic, p.bodyFrame)
	}
}

func (p *Player) getNextBodyFrame(dirFrames []int) pixel.Rect {
	dt := time.Since(p.lastBodyFrame).Seconds()
	p.lastBodyFrame = time.Now()
	if p.moving {
		p.bodyStep += 11 * dt
		newFrame := int(p.bodyStep)
		if (newFrame <= 5 && (p.dir == "up" || p.dir == "down")) || (newFrame <= 4 && (p.dir == "right" || p.dir == "left")) {
			return p.bodyFrames[dirFrames[newFrame]]
		}
	}
	p.bodyStep = 0
	return p.bodyFrames[dirFrames[0]]
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

type Resu struct {
	PosBody, PosHead     pixel.Vec
	BodyPic, HeadPic     pixel.Picture
	BodyFrame, HeadFrame pixel.Rect
}

func NewResu() *Resu {

	r := Resu{}
	r.PosBody = pixel.V(500, 500)
	r.PosHead = pixel.V(501, 524)
	r.BodyPic, r.HeadPic = Pictures["./images/curaBody.png"], Pictures["./images/curaHead.png"]
	r.BodyFrame, r.HeadFrame = r.BodyPic.Bounds(), r.HeadPic.Bounds()
	return &r
}

func (r *Resu) Draw(win *pixelgl.Window, cam pixel.Matrix, p *Player) {
	if win.JustPressed(pixelgl.MouseButtonRight) {
		mouse := cam.Unproject(win.MousePosition())
		if r.OnMe(mouse) && p.dead {
			p.dead = false
			p.hp = MaxHealth
			p.mp = MaxMana
		}
	}
	head := pixel.NewSprite(r.HeadPic, r.HeadFrame)
	body := pixel.NewSprite(r.BodyPic, r.BodyFrame)
	head.Draw(win, pixel.IM.Moved(r.PosHead))
	body.Draw(win, pixel.IM.Moved(r.PosBody))
}

func (r *Resu) OnMe(click pixel.Vec) bool {
	b := click.X < r.PosBody.X+18 && click.X > r.PosBody.X-18 && click.Y < r.PosBody.Y+25 && click.Y > r.PosBody.Y-25
	return b
}

type Forest struct {
	Pic, GrassPic     pixel.Picture
	Frames            []pixel.Rect
	GrassFrame        pixel.Rect
	Batch, GrassBatch *pixel.Batch
}

func NewForest() *Forest {
	treeSheet := Pictures["./images/trees.png"]
	treeBatch := pixel.NewBatch(&pixel.TrianglesData{}, treeSheet)
	treeFrames := getFrames(treeSheet, 32, 32, 3, 3)

	for i := 0; i <= TreeQuantity; i++ {
		pos := pixel.ZV
		dirX := rand.Float64()
		dirY := rand.Float64()
		if dirX < .5 {
			pos = pos.Add(pixel.V(dirX*4000, 0))
		} else {
			pos = pos.Sub(pixel.V(-dirX*4000, 0))
		}
		if dirY < .5 {
			pos = pos.Add(pixel.V(0, dirY*4000))
		} else {
			pos = pos.Sub(pixel.V(0, -dirY*4000))
		}
		tree := pixel.NewSprite(treeSheet, treeFrames[rand.Intn(len(treeFrames))])
		tree.Draw(treeBatch, pixel.IM.Scaled(pixel.ZV, 3.5).Moved(pos))
	}

	// grassSheet := Pictures["./images/grass.png"]
	// grassBatch := pixel.NewBatch(&pixel.TrianglesData{}, grassSheet)
	// grassFrame := grassSheet.Bounds().Resized(grassSheet.Bounds().Center(), pixel.V(40, 40))
	// grass := pixel.NewSprite(grassSheet, grassFrame)
	// for x := 0; x <= 40; x++ {
	// 	for y := 0; y <= 40; y++ {
	// 		pos := pixel.V(float64(x)*40, float64(y)*40)
	// 		log.Println(pos.String())
	// 		grass.Draw(grassBatch, pixel.IM.Scaled(pixel.ZV, 1).Moved(pos))
	// 	}
	// }
	return &Forest{
		Pic:    treeSheet,
		Frames: treeFrames,
		Batch:  treeBatch,
		// GrassPic:   grassSheet,
		// GrassFrame: grassFrame,
		// GrassBatch: grassBatch,
	}
}
