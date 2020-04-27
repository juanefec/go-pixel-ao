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
	PlayerSpeed = 185.0
	Zoom        = 1.0
	ZoomSpeed   = 1.1
	fps         = 0
	second      = time.Tick(time.Second)
	MaxMana     = 2104
	MaxHealth   = 307
	ApocaDmg    = 190
)
var (
	Newline      = []byte{'\n'}
	TreeQuantity = 300
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
		"./images/bodydruida.png",
		"./images/heads.png",
		"./images/trees.png",
		"./images/dead.png",
		"./images/deadHead.png",
		"./images/desca.png",
		"./images/curaBody.png",
		"./images/curaHead.png",
		"./images/buda.png",
		"./images/sauce.png",
		"./images/grass.png",
		"./images/staff.png",
		"./images/hatpro.png",
		"./images/horizontalfence.png",
		"./images/verticalfence.png",
		"./images/bodyRedIcon.png",
		"./images/bodyBlueIcon.png",
		"./images/blueBody.png",
		"./images/redBody.png",
		"./images/explosion.png",
		"./images/fireball.png",
		"./images/creagod.png",
	)
	ld, err := SetNameWindow()
	if err != nil {
		log.Panic(err)
	}
	player := NewPlayer(ld.Name, ld.Side)
	spells := GameSpells{
		NewSpellData("apoca", &player),
		NewSpellData("desca", &player),
		NewSpellData("explo", &player),
		NewSpellData("fireball", &player),
	}
	forest := NewForest()
	buda := NewBuda(pixel.V(2000, 3400))
	otherPlayers := NewPlayersData()
	playerInfo := NewPlayerInfo(&player, &otherPlayers)
	resu := NewResu(pixel.V(2000, 2900))

	socket := socket.NewSocket("25.9.191.221", 3333)
	defer socket.Close()

	cfg := pixelgl.WindowConfig{
		Title:  "Creative AO",
		Bounds: pixel.R(0, 0, 850, 650),
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	cursor := NewCursor(win)
	go keyInputs(win, &player, cursor)
	go GameUpdate(socket, &otherPlayers, &player, spells...)

	for !win.Closed() {
		win.Clear(colornames.Black)
		cam := pixel.IM.Scaled(player.pos, Zoom).Moved(win.Bounds().Center().Sub(player.pos))
		player.cam = cam
		win.SetMatrix(cam)
		Zoom *= math.Pow(ZoomSpeed, win.MouseScroll().Y)

		forest.GrassBatch.Draw(win)
		forest.FenceBatchHTOP.Draw(win)
		resu.Draw(win, cam, &player)
		otherPlayers.Draw(win)
		player.Draw(win)
		buda.Draw(win)
		forest.Batch.Draw(win)
		forest.SauceBatch.Draw(win)
		forest.FenceBatchV.Draw(win)
		forest.FenceBatchHBOT.Draw(win)
		spells.Draw(win, cam, socket, &otherPlayers, cursor)
		playerInfo.Draw(win, cam)
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
	SpellCastExplo
	SpellCastFireball
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
	c.Mode = SpellCastFireball
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
	playersData                                *PlayersData
	player                                     *Player
	hdisplay, mdisplay, onsdisplay, posdisplay *text.Text
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
	pi.posdisplay = text.New(pixel.ZV, basicAtlas)
	pi.posdisplay.Color = colornames.Whitesmoke
	fmt.Fprintf(pi.posdisplay, "X: %v\nY: %v", player.pos.X, player.pos.Y)
	pi.player = player
	pi.playersData = pd
	return &pi
}

func (pi *PlayerInfo) Draw(win *pixelgl.Window, cam pixel.Matrix) {
	winSize := win.Bounds().Max
	infoPos := cam.Unproject(winSize.Add(pixel.V(-170, -10)))
	info := imdraw.New(nil)
	info.Color = colornames.Black
	info.EndShape = imdraw.SharpEndShape
	info.Push(
		infoPos.Add(pixel.V(-2, 2)),
		infoPos.Add(pixel.V(152, 2)),
		infoPos.Add(pixel.V(152, -22)),
		infoPos.Add(pixel.V(-2, -22)),
		infoPos.Add(pixel.V(-2, 2)),
	)
	info.Line(4)
	info.Push(
		infoPos.Add(pixel.V(-2, -28)),
		infoPos.Add(pixel.V(152, -28)),
		infoPos.Add(pixel.V(152, -52)),
		infoPos.Add(pixel.V(-2, -52)),
		infoPos.Add(pixel.V(-2, -28)),
	)
	info.Line(4)
	info.Color = pixel.RGB(1, 0, 0)
	hval := Map(float64(pi.player.hp), 0, float64(MaxHealth), 0, 150)
	info.Push(
		infoPos.Add(pixel.V(0, 0)),
		infoPos.Add(pixel.V(hval, 0)),
		infoPos.Add(pixel.V(0, -20)),
		infoPos.Add(pixel.V(hval, -20)),
	)
	info.Rectangle(0)
	info.Color = pixel.RGB(0, 0, 1)
	mval := Map(float64(pi.player.mp), 0, float64(MaxMana), 0, 150)
	info.Push(
		infoPos.Add(pixel.V(0, -30)),
		infoPos.Add(pixel.V(mval, -30)),
		infoPos.Add(pixel.V(0, -50)),
		infoPos.Add(pixel.V(mval, -50)),
	)
	info.Rectangle(0)
	info.Draw(win)
	pi.hdisplay.Clear()
	pi.mdisplay.Clear()
	pi.onsdisplay.Clear()
	pi.posdisplay.Clear()
	fmt.Fprintf(pi.hdisplay, "%v/%v", pi.player.hp, MaxHealth)
	fmt.Fprintf(pi.mdisplay, "%v/%v", pi.player.mp, MaxMana)
	fmt.Fprintf(pi.onsdisplay, "Online: %v", pi.playersData.Online+1)
	fmt.Fprintf(pi.posdisplay, "X: %v\nY: %v", int(pi.player.pos.X/10), int(pi.player.pos.Y/10))
	hmatrix := pixel.IM.Moved(infoPos.Add(pixel.V(46, -15)))
	mmatrix := pixel.IM.Moved(infoPos.Add(pixel.V(40, -45)))
	onspos := infoPos.Add(pixel.V(-winSize.X+190, -20))
	onsmatrix := pixel.IM.Moved(onspos).Scaled(onspos, 2)
	pospos := infoPos.Add(pixel.V(-winSize.X+190, -50))
	posmatrix := pixel.IM.Moved(pospos)

	pi.hdisplay.Draw(win, hmatrix)
	pi.mdisplay.Draw(win, mmatrix)
	pi.onsdisplay.Draw(win, onsmatrix)
	pi.posdisplay.Draw(win, posmatrix)
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

type GameSpells []*SpellData

func (gs GameSpells) Draw(win *pixelgl.Window, cam pixel.Matrix, s *socket.Socket, pd *PlayersData, cursor *Cursor) {
	for i := range gs {
		gs[i].Batch.Clear()
		gs[i].Update(win, cam, s, pd, cursor)
		gs[i].Batch.Draw(win)
	}
}

type SpellData struct {
	SpellName         string
	SpellMode         CursorMode
	ManaCost, Damage  int
	SpellSpeed        float64
	Caster            *Player
	Frames            []pixel.Rect
	Pic               *pixel.Picture
	Batch             *pixel.Batch
	CurrentAnimations []*Spell
}

func (sd *SpellData) Update(win *pixelgl.Window, cam pixel.Matrix, s *socket.Socket, pd *PlayersData, cursor *Cursor) {
	dt := time.Since(sd.Caster.lastCast).Seconds()
	dtproj := time.Since(sd.Caster.lastCastProj).Seconds()
	casted := false
	if win.JustPressed(pixelgl.MouseButtonLeft) && !sd.Caster.dead && sd.Caster.mp >= sd.ManaCost && cursor.Mode == sd.SpellMode {
		if dt >= ((time.Second.Seconds() / 10) * 9) {
			sd.Caster.lastCast = time.Now()
			for key := range pd.CurrentAnimations {
				mouse := cam.Unproject(win.MousePosition())
				if !pd.CurrentAnimations[key].dead && cursor.Mode != SpellCastFireball && pd.CurrentAnimations[key].OnMe(mouse) {
					spell := models.SpellMsg{
						ID:       s.ClientID,
						Type:     sd.SpellName,
						TargetID: key,
						Name:     sd.Caster.sname,
						X:        mouse.X,
						Y:        mouse.Y,
					}
					paylaod, _ := json.Marshal(spell)
					s.O <- models.NewMesg(models.Spell, paylaod)

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

					casted = true
					break
				}
			}
		}

		if !casted && cursor.Mode == sd.SpellMode && cursor.Mode == SpellCastFireball && dtproj >= ((time.Second.Seconds()/10)*6) {
			sd.Caster.lastCastProj = time.Now()
			mouse := cam.Unproject(win.MousePosition())
			spell := models.SpellMsg{
				ID:       s.ClientID,
				Type:     sd.SpellName,
				TargetID: ksuid.Nil,
				Name:     sd.Caster.sname,
				X:        mouse.X,
				Y:        mouse.Y,
			}
			paylaod, _ := json.Marshal(spell)
			s.O <- models.NewMesg(models.Spell, paylaod)

			vel := mouse.Sub(cam.Unproject(win.Bounds().Center()))
			centerMatrix := sd.Caster.bodyMatrix.Rotated(cam.Unproject(win.Bounds().Center()), vel.Angle()+(math.Pi/2)).Scaled(cam.Unproject(win.Bounds().Center()), 2)
			sd.Caster.mp -= sd.ManaCost
			newSpell := &Spell{
				caster:         s.ClientID,
				pos:            sd.Caster.pos,
				vel:            vel,
				spellName:      &sd.SpellName,
				step:           sd.Frames[0],
				frameNumber:    0.0,
				matrix:         &centerMatrix,
				last:           time.Now(),
				projectileLife: time.Now(),
			}
			newSpell.frame = pixel.NewSprite(*(sd.Pic), newSpell.step)
			sd.CurrentAnimations = append(sd.CurrentAnimations, newSpell)
			casted = true
		}
		if dt >= ((time.Second.Seconds() / 10) * 9) {
			cursor.SetNormalMode()
		}
	}
	if sd.SpellName == "fireball" {
	FBALLS:
		for i := 0; i <= len(sd.CurrentAnimations)-1; i++ {
			next, kill := sd.CurrentAnimations[i].NextFrameFireball(sd.Frames)
			if kill {
				if i < len(sd.CurrentAnimations)-1 {
					copy(sd.CurrentAnimations[i:], sd.CurrentAnimations[i+1:])
				}
				sd.CurrentAnimations[len(sd.CurrentAnimations)-1] = nil // or the zero sd.vCurrentAnimationslue of T
				sd.CurrentAnimations = sd.CurrentAnimations[:len(sd.CurrentAnimations)-1]
				continue
			}
			for key := range pd.CurrentAnimations {
				p := pd.CurrentAnimations[key]
				if sd.CurrentAnimations[i].caster != key && !p.dead && p.OnMe(sd.CurrentAnimations[i].pos) {
					if i < len(sd.CurrentAnimations)-1 {
						copy(sd.CurrentAnimations[i:], sd.CurrentAnimations[i+1:])
					}
					sd.CurrentAnimations[len(sd.CurrentAnimations)-1] = nil // or the zero sd.vCurrentAnimationslue of T
					sd.CurrentAnimations = sd.CurrentAnimations[:len(sd.CurrentAnimations)-1]

					p.hp -= sd.Damage
					if p.hp <= 0 {
						p.hp = 0
						p.dead = true
					}
					continue FBALLS
				}
			}
			if sd.CurrentAnimations[i].caster != s.ClientID && !sd.Caster.dead && sd.Caster.OnMe(sd.CurrentAnimations[i].pos) {
				sd.Caster.hp -= sd.Damage
				if sd.Caster.hp <= 0 {
					sd.Caster.hp = 0
					sd.Caster.dead = true
				}
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
	} else {
		for i := 0; i <= len(sd.CurrentAnimations)-1; i++ {
			next, kill := sd.CurrentAnimations[i].NextFrame(sd.Frames, sd.SpellSpeed)
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
}

func NewSpellData(spell string, caster *Player) *SpellData {

	var sheet pixel.Picture
	var batch *pixel.Batch
	var frames []pixel.Rect
	var mode CursorMode
	var manaCost, damage int
	var speed float64 = 21
	switch spell {
	case "apoca":
		sheet = Pictures["./images/apocas.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		unorderedFrames := getFrames(sheet, 145, 145, 4, 4)
		for i := range unorderedFrames {
			frames = append(frames, unorderedFrames[ApocaFrames[i]])
		}
		mode = SpellCastApoca
		manaCost = 1000
		damage = 180

		break
	case "desca":
		sheet = Pictures["./images/desca.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		frames = getFrames(sheet, 127, 127, 5, 3)
		mode = SpellCastDesca
		manaCost = 460
		damage = 130
	case "explo":
		sheet = Pictures["./images/explosion.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		frames = getFrames(sheet, 96, 96, 12, 0)
		mode = SpellCastExplo
		manaCost = 1550
		damage = 215
		speed = 13
	case "fireball":
		sheet = Pictures["./images/fireball.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		frames = getFrames(sheet, 24, 24, 7, 0)
		mode = SpellCastFireball
		manaCost = 300
		damage = 80
	}

	return &SpellData{
		SpellSpeed:        speed,
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
	caster         ksuid.KSUID
	vel, pos       pixel.Vec // para proyectiles
	projectileLife time.Time
	target         *Player
	spellName      *string
	step           pixel.Rect
	frame          *pixel.Sprite
	frameNumber    float64
	matrix         *pixel.Matrix
	last           time.Time
}

func (a *Spell) NextFrame(spellFrames []pixel.Rect, speed float64) (pixel.Rect, bool) {
	dt := time.Since(a.last).Seconds()
	a.last = time.Now()
	a.frameNumber += speed * dt
	i := int(a.frameNumber)
	if i <= len(spellFrames)-1 {
		return spellFrames[i], false
	}
	a.frameNumber = .0
	return spellFrames[0], true
}

func (a *Spell) NextFrameFireball(spellFrames []pixel.Rect) (pixel.Rect, bool) {
	dt := time.Since(a.last).Seconds()
	pdt := time.Since(a.projectileLife).Seconds()
	a.last = time.Now()
	a.frameNumber += 21 * dt
	i := int(a.frameNumber)
	if i <= len(spellFrames)-1 {
		vel := pixel.V(1, 1).Rotated(a.vel.Angle()).Rotated(-pixel.V(1, 1).Angle()).Scaled(1.4)
		a.pos = a.pos.Add(vel)
		(*a.matrix) = a.matrix.Moved(vel)
		return spellFrames[i], false
	}

	a.frameNumber = .0
	if pdt > time.Second.Seconds()*1.5 {
		return spellFrames[0], true
	}
	return spellFrames[0], false
}

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
	GodBody
	Head
	CoolHat
	Staff
	Phantom
	PhantomHead
)

type Skins []*Skin

func (s Skins) Load(imagPath string, t SkinType, w, h, qw, qh float64) {
	sheet := Pictures[imagPath]
	skin := &Skin{
		Pic:    &sheet,
		Batch:  pixel.NewBatch(&pixel.TrianglesData{}, sheet),
		Frames: getFrames(sheet, w, h, qw, qh),
	}
	s[t] = skin
}

func (s Skins) BatchClear() {
	for i := range s {
		s[i].Batch.Clear()
	}
}
func (s Skins) DrawToBatch(p *Player) {
	p.Update()
	if !p.dead {
		p.body.Draw(s[p.bodySkin].Batch, p.bodyMatrix)
		p.bacu.Draw(s[p.staffSkin].Batch, p.bodyMatrix)
		p.head.Draw(s[p.headSkin].Batch, p.headMatrix)
		p.hat.Draw(s[p.hatSkin].Batch, p.hatMatrix)
	} else {
		p.body.Draw(s[Phantom].Batch, p.bodyMatrix)
		p.head.Draw(s[PhantomHead].Batch, p.headMatrix)
	}
}

func (s Skins) Draw(win *pixelgl.Window) {
	for i := range s {
		s[i].Batch.Draw(win)
	}
}

type Skin struct {
	Frames []pixel.Rect
	Pic    *pixel.Picture
	Batch  *pixel.Batch
}

func NewPlayersData() PlayersData {
	pd := PlayersData{}
	pd.SkinsLen = 9
	pd.Skins = make([]*Skin, pd.SkinsLen)

	pd.Skins.Load("./images/bodydruida.png", TuniDruida, 25, 45, 6, 4)
	pd.Skins.Load("./images/redBody.png", RedBody, 25, 45, 6, 4)
	pd.Skins.Load("./images/blueBody.png", BlueBody, 25, 45, 6, 4)
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
							np := NewPlayer(p.Name, SkinType(p.Skin))
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

				newSpell := &Spell{
					spellName:      &spell.Type,
					frameNumber:    0.0,
					last:           time.Now(),
					projectileLife: time.Now(),
				}

				target := &Player{}
				if spell.Type != "fireball" {
					if s.ClientID == spell.TargetID {
						target = p
					} else {
						target = pd.CurrentAnimations[spell.TargetID]
					}
					newSpell.target = target
					newSpell.matrix = &target.headMatrix
				}

				for i := range ssd {
					sd := ssd[i]
					if spell.Type == sd.SpellName && spell.Type != "fireball" {
						newSpell.step = sd.Frames[0]
						newSpell.frame = pixel.NewSprite(*(sd.Pic), newSpell.step)
						sd.CurrentAnimations = append(sd.CurrentAnimations, newSpell)
						target.hp -= sd.Damage
						if target.hp <= 0 {
							target.hp = 0
							target.dead = true
						}
						break
					} else if spell.Type == sd.SpellName && spell.Type == "fireball" {
						caster := pd.CurrentAnimations[spell.ID]
						vel := pixel.V(spell.X, spell.Y).Sub(caster.pos)
						centerMatrix := caster.bodyMatrix.Rotated(caster.pos, vel.Angle()+(math.Pi/2)).Scaled(caster.pos, 2)
						newSpell.caster = spell.ID
						newSpell.vel = vel
						newSpell.pos = caster.pos
						newSpell.matrix = &centerMatrix
						newSpell.step = sd.Frames[0]
						newSpell.frame = pixel.NewSprite(*(sd.Pic), newSpell.step)
						sd.CurrentAnimations = append(sd.CurrentAnimations, newSpell)
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
	pd.Skins.BatchClear()
	pd.AnimationsMutex.RLock()
	for _, p := range pd.CurrentAnimations {
		pd.AnimationsMutex.RUnlock()
		pd.Skins.DrawToBatch(p)
		p.name.Draw(win, p.nameMatrix)
		pd.AnimationsMutex.RLock()
		//player.name.Draw(win, player.nameMatrix)
	}
	pd.AnimationsMutex.RUnlock()
	pd.Skins.Draw(win)

}

type Player struct {
	bodyFrames, headFrames, deadFrames, deadHeadFrames, bacuFrames, hatFrames []pixel.Rect
	headPic, bodyPic, deadPic, deadHeadPic, bacuPic, hatPic                   *pixel.Picture
	cam, headMatrix, bodyMatrix, nameMatrix, hatMatrix                        pixel.Matrix
	bodyFrame, headFrame, bacuFrame, hatFrame                                 pixel.Rect
	bodySkin, headSkin, hatSkin, staffSkin                                    SkinType
	head, body, bacu, hat                                                     *pixel.Sprite
	hp, mp                                                                    int // health/mana points

	pos                   pixel.Vec
	name                  *text.Text
	sname                 string
	playerUpdate          *models.PlayerMsg
	dir                   string
	moving                bool
	bodyStep              float64
	lastDeadFrame         time.Time
	lastBodyFrame         time.Time
	lastDrank             time.Time
	lastCast              time.Time
	lastCastProj          time.Time
	drinkingManaPotions   bool
	drinkingHealthPotions bool
	dead                  bool
}

func NewPlayer(name string, skin SkinType) Player {
	p := &Player{}
	p.sname = name
	basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	p.name = text.New(pixel.V(-28, 0), basicAtlas)

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

	p.bodySkin = skin
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
		p.name.Color = colornames.Blue

	case GodBody:
		bodySheet = Pictures["./images/creagod.png"]
		p.name.Color = colornames.Lightseagreen

	}

	fmt.Fprintln(p.name, name)
	bodyFrames = getFrames(bodySheet, 25, 45, 6, 4)

	p.playerUpdate = &models.PlayerMsg{}
	p.lastBodyFrame = time.Now()
	p.lastDeadFrame = time.Now()
	p.lastDrank = time.Now()
	p.lastCast = time.Now()

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
	p.mp = MaxMana
	p.hp = MaxHealth
	return *p
}

func (p *Player) OnMe(click pixel.Vec) bool {
	r := click.X < p.pos.X+14 && click.X > p.pos.X-14 && click.Y < p.pos.Y+30 && click.Y > p.pos.Y-20
	return r
}

func (p *Player) clientUpdate(s *socket.Socket) {
	p.playerUpdate = &models.PlayerMsg{
		ID:     s.ClientID,
		Name:   p.sname,
		Skin:   int(p.bodySkin),
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
		p.nameMatrix = pixel.IM.Moved(p.pos.Sub(p.name.Bounds().Center()).Add(pixel.V(0, -26)))
		p.head = pixel.NewSprite(*p.headPic, p.headFrame)
		p.body = pixel.NewSprite(*p.bodyPic, p.bodyFrame)
		p.bacu = pixel.NewSprite(*p.bacuPic, p.bacuFrame)
		p.hat = pixel.NewSprite(*p.hatPic, p.hatFrame)
		dt := time.Since(p.lastDrank).Seconds()
		second := time.Second.Seconds()
		if p.drinkingHealthPotions && !p.drinkingManaPotions {
			if dt > second/3.2 {
				p.hp += 30
				if p.hp > MaxHealth {
					p.hp = MaxHealth
				}
				p.lastDrank = time.Now()
			}
		}
		if p.drinkingManaPotions && !p.drinkingHealthPotions {
			if dt > second/4 {
				p.mp += 105
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
		p.headMatrix = pixel.IM.Moved(p.pos.Add(pixel.V(1, 20)))
		p.bodyMatrix = pixel.IM.Moved(p.pos.Add(pixel.V(0, 0)))
		p.nameMatrix = pixel.IM.Moved(p.pos.Sub(p.name.Bounds().Center()).Add(pixel.V(0, -26)))
		p.head = pixel.NewSprite(*p.deadHeadPic, p.headFrame)
		p.body = pixel.NewSprite(*p.deadPic, p.bodyFrame)
	}
}

func (p *Player) Draw(win *pixelgl.Window) {
	p.Update()
	p.body.Draw(win, p.bodyMatrix)
	p.head.Draw(win, p.headMatrix)
	p.name.Draw(win, p.nameMatrix)
	if !p.dead {
		p.bacu.Draw(win, p.bodyMatrix)
		p.hat.Draw(win, p.hatMatrix)
	}
}

func (p *Player) getNextBodyFrame(dirFrames []int, part []pixel.Rect) pixel.Rect {
	dt := time.Since(p.lastBodyFrame).Seconds()
	p.lastBodyFrame = time.Now()
	if p.moving {
		p.bodyStep += 13 * dt
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

type Resu struct {
	PosBody, PosHead       pixel.Vec
	BodyPic, HeadPic       pixel.Picture
	BodyFrame, HeadFrame   pixel.Rect
	BodySprite, HeadSprite *pixel.Sprite
}

func NewResu(pos pixel.Vec) *Resu {

	r := Resu{}
	r.PosBody = pos
	r.PosHead = pos.Add(pixel.V(1, 24))
	r.BodyPic, r.HeadPic = Pictures["./images/curaBody.png"], Pictures["./images/curaHead.png"]
	r.BodyFrame, r.HeadFrame = r.BodyPic.Bounds(), r.HeadPic.Bounds()
	r.HeadSprite = pixel.NewSprite(r.HeadPic, r.HeadFrame)
	r.BodySprite = pixel.NewSprite(r.BodyPic, r.BodyFrame)
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
	r.HeadSprite.Draw(win, pixel.IM.Moved(r.PosHead))
	r.BodySprite.Draw(win, pixel.IM.Moved(r.PosBody))
}

func (r *Resu) OnMe(click pixel.Vec) bool {
	b := click.X < r.PosBody.X+14 && click.X > r.PosBody.X-14 && click.Y < r.PosBody.Y+30 && click.Y > r.PosBody.Y-20
	return b
}

type Tree struct {
}
type Forest struct {
	Pic, SaucePic, GrassPic, FencePicH, FencePicV                              pixel.Picture
	Frames, GrassFrames                                                        []pixel.Rect
	SauceFrame, FenceFrameH, FenceFrameV                                       pixel.Rect
	Batch, SauceBatch, GrassBatch, FenceBatchHTOP, FenceBatchHBOT, FenceBatchV *pixel.Batch
}

func NewForest() *Forest {
	treeSheet := Pictures["./images/trees.png"]
	treeBatch := pixel.NewBatch(&pixel.TrianglesData{}, treeSheet)
	treeFrames := getFrames(treeSheet, 32, 32, 3, 3)

	sauceSheet := Pictures["./images/sauce.png"]
	sauceBatch := pixel.NewBatch(&pixel.TrianglesData{}, sauceSheet)
	sauceFrame := sauceSheet.Bounds()

	grassSheet := Pictures["./images/grass.png"]
	grassBatch := pixel.NewBatch(&pixel.TrianglesData{}, grassSheet)
	grassFrames := getFrames(grassSheet, 128, 128, 3, 3)

	vfenceSheet := Pictures["./images/verticalfence.png"]
	vfenceBatch := pixel.NewBatch(&pixel.TrianglesData{}, vfenceSheet)
	vfenceFrame := vfenceSheet.Bounds()
	hfenceSheet := Pictures["./images/horizontalfence.png"]
	hfenceBatchBot := pixel.NewBatch(&pixel.TrianglesData{}, hfenceSheet)
	hfenceBatchTop := pixel.NewBatch(&pixel.TrianglesData{}, hfenceSheet)
	hfenceFrame := hfenceSheet.Bounds()

	for x := 0; x <= 32; x++ {
		for y := 0; y <= 32; y++ {
			pos := pixel.V(float64(x*128)-64, float64(y*128)-64)
			bread := pixel.NewSprite(grassSheet, grassFrames[rand.Intn(len(grassFrames))])
			bread.Draw(grassBatch, pixel.IM.Moved(pos))
		}
	}

	for x := 0; x < 31; x++ {
		top := pixel.V(float64(x*128)+64, 4000)
		bottom := pixel.V(float64(x*128)+64, 0)
		fence := pixel.NewSprite(hfenceSheet, hfenceFrame)
		fence.Draw(hfenceBatchTop, pixel.IM.Moved(top))
		fence.Draw(hfenceBatchBot, pixel.IM.Moved(bottom))
	}

	for x := 0; x < 31; x++ {
		left := pixel.V(-9, float64(x*128)+64)
		rigth := pixel.V(4010, float64(x*128)+64)
		fence := pixel.NewSprite(vfenceSheet, vfenceFrame)
		fence.Draw(vfenceBatch, pixel.IM.Moved(left))
		fence.Draw(vfenceBatch, pixel.IM.Moved(rigth))
	}

	pathTreeLength := 22
	treeSeparation := 110
	pathTop := .0
	// Make path
	for i := 0; i <= pathTreeLength; i++ {
		h := float64(300 + i*treeSeparation)
		pos1 := pixel.V(1850, h)
		pos2 := pixel.V(2150, h)

		tree := pixel.NewSprite(treeSheet, treeFrames[rand.Intn(len(treeFrames))])
		tree.Draw(treeBatch, pixel.IM.Scaled(pixel.ZV, 3.5).Moved(pos1))
		tree.Draw(treeBatch, pixel.IM.Scaled(pixel.ZV, 3.5).Moved(pos2))
		pathTop = h
	}

	// Make "arena"
	for i := 0; i <= 18; i++ {
		w := float64(1000 + i*treeSeparation)
		top := pixel.V(w, 3850)
		bottom := pixel.V(w, pathTop)
		tree := pixel.NewSprite(treeSheet, treeFrames[rand.Intn(len(treeFrames))])
		tree.Draw(treeBatch, pixel.IM.Scaled(pixel.ZV, 3.5).Moved(top))
		if bottom.X < 1800 || bottom.X > 2200 {
			tree.Draw(treeBatch, pixel.IM.Scaled(pixel.ZV, 3.5).Moved(bottom))
		}
	}
	arenaTop := .0
	for i := 0; i <= 10; i++ {
		h := pathTop + float64(i*treeSeparation)
		left := pixel.V(1000, h)
		rigth := pixel.V(3000, h)
		tree := pixel.NewSprite(treeSheet, treeFrames[rand.Intn(len(treeFrames))])
		tree.Draw(treeBatch, pixel.IM.Scaled(pixel.ZV, 3.5).Moved(left))
		tree.Draw(treeBatch, pixel.IM.Scaled(pixel.ZV, 3.5).Moved(rigth))
		arenaTop = h
	}

	// Fill outside
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
		if pos.X < 1800 || pos.X > 2200 || pos.Y < 300 || pos.Y > float64(pathTreeLength*treeSeparation) {
			if pos.Y < pathTop || pos.Y > arenaTop || pos.X < 1000 || pos.X > 3000 {
				tree.Draw(treeBatch, pixel.IM.Scaled(pixel.ZV, 3.5).Moved(pos))

			}
		}
	}
	// Fill inside

	arenaTrees := []pixel.Vec{
		pixel.V(1100, 3010),
		pixel.V(1390, 2960),
		pixel.V(1700, 2980),
		pixel.V(2300, 3012),
		pixel.V(2550, 3060),
		pixel.V(2770, 2940),

		pixel.V(1130, 3322),
		pixel.V(1370, 3350),
		pixel.V(1740, 3260),
		pixel.V(2280, 3320),
		pixel.V(2550, 3250),
		pixel.V(2780, 3270),

		pixel.V(1100, 3610),
		pixel.V(1390, 3560),
		pixel.V(1700, 3580),
		pixel.V(1990, 3720),
		pixel.V(2300, 3612),
		pixel.V(2550, 3660),
		pixel.V(2770, 3540),
	}
	for i := 0; i <= len(arenaTrees)-1; i++ {

		tree := pixel.NewSprite(sauceSheet, sauceFrame)

		tree.Draw(sauceBatch, pixel.IM.Moved(arenaTrees[i]))
	}

	return &Forest{
		Pic:            treeSheet,
		Frames:         treeFrames,
		Batch:          treeBatch,
		SauceBatch:     sauceBatch,
		SauceFrame:     sauceFrame,
		SaucePic:       sauceSheet,
		GrassBatch:     grassBatch,
		GrassFrames:    grassFrames,
		GrassPic:       grassSheet,
		FenceBatchHTOP: hfenceBatchTop,
		FenceBatchHBOT: hfenceBatchBot,
		FenceFrameH:    hfenceFrame,
		FencePicH:      hfenceSheet,
		FenceBatchV:    vfenceBatch,
		FenceFrameV:    vfenceFrame,
		FencePicV:      vfenceSheet,
	}
}

type Buda struct {
	Pos        pixel.Vec
	Pic        pixel.Picture
	Frame      pixel.Rect
	BudaSprite *pixel.Sprite
}

func NewBuda(pos pixel.Vec) *Buda {
	b := Buda{}
	b.Pic = Pictures["./images/buda.png"]
	b.Frame = b.Pic.Bounds()
	b.BudaSprite = pixel.NewSprite(b.Pic, b.Frame)
	b.Pos = pos
	return &b
}

func (b *Buda) Draw(win *pixelgl.Window) {
	b.BudaSprite.Draw(win, pixel.IM.Moved(b.Pos))
}
