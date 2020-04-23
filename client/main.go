package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	_ "image/png"

	"github.com/faiface/pixel"
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
)
var (
	Newline      = []byte{'\n'}
	TreeQuantity = 1000
	BodyUp       = []int{12, 13, 14, 15, 16, 17}
	BodyDown     = []int{18, 19, 20, 21, 22, 23}
	BodyLeft     = []int{6, 7, 8, 9, 10}
	BodyRight    = []int{0, 1, 2, 3, 4}

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
		"./images/grass.png",
		"./images/desca.png",
	)

	player := NewPlayer()
	apocaData := NewSpellData("apoca", &player)
	forest := NewForest()
	otherPlayers := NewPlayersData()

	socket := socket.NewSocket("127.0.0.1", 3333)
	defer socket.Close()

	cfg := pixelgl.WindowConfig{
		Title:  "Creative AO",
		Bounds: pixel.R(0, 0, 600, 600),
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	go keyInputs(win, &player)
	go GameUpdate(socket, &otherPlayers, apocaData)

	for !win.Closed() {
		apocaData.Batch.Clear()

		cam := pixel.IM.Scaled(player.pos, Zoom).Moved(win.Bounds().Center().Sub(player.pos))
		win.SetMatrix(cam)
		win.Clear(colornames.Forestgreen)

		apocaData.Update(win, cam, socket)
		player.Update()
		Zoom *= math.Pow(ZoomSpeed, win.MouseScroll().Y)
		//forest.GrassBatch.Draw(win)
		otherPlayers.Draw(win)
		player.body.Draw(win, player.bodyMatrix)
		player.head.Draw(win, player.headMatrix)
		player.name.Draw(win, player.nameMatrix)
		forest.Batch.Draw(win)
		apocaData.Batch.Draw(win)

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

type SpellData struct {
	SpellName         string
	Caster            *Player
	Frames            []pixel.Rect
	Pic               *pixel.Picture
	Batch             *pixel.Batch
	CurrentAnimations []*Spell
}

func (ad *SpellData) Update(win *pixelgl.Window, cam pixel.Matrix, s *socket.Socket) {
	if win.JustPressed(pixelgl.MouseButtonLeft) {
		mouse := cam.Unproject(win.MousePosition())
		newSpell := &Spell{
			spellName:   &ad.SpellName,
			step:        ad.Frames[0],
			frameNumber: 0.0,
			matrix:      pixel.IM.Scaled(pixel.ZV, .7).Moved(mouse),
			last:        time.Now(),
		}

		newSpell.frame = pixel.NewSprite(*(ad.Pic), newSpell.step)
		ad.CurrentAnimations = append(ad.CurrentAnimations, newSpell)
		spell := models.SpellMsg{
			ID:   s.ClientID,
			Name: "name",
			X:    mouse.X,
			Y:    mouse.Y,
		}
		paylaod, _ := json.Marshal(spell)
		s.O <- models.NewMesg(models.Spell, paylaod)

	}

	for i := 0; i <= len(ad.CurrentAnimations)-1; i++ {
		next, kill := ad.CurrentAnimations[i].NextFrame(ad.Frames)
		if kill {
			if i < len(ad.CurrentAnimations)-1 {
				copy(ad.CurrentAnimations[i:], ad.CurrentAnimations[i+1:])
			}
			ad.CurrentAnimations[len(ad.CurrentAnimations)-1] = nil // or the zero ad.vCurrentAnimationslue of T
			ad.CurrentAnimations = ad.CurrentAnimations[:len(ad.CurrentAnimations)-1]
			continue
		}
		ad.CurrentAnimations[i].step = next
		ad.CurrentAnimations[i].frame = pixel.NewSprite(*ad.Pic, ad.CurrentAnimations[i].step)
		ad.CurrentAnimations[i].frame.Draw(ad.Batch, ad.CurrentAnimations[i].matrix)
	}
}

func NewSpellData(spell string, caster *Player) *SpellData {

	var sheet pixel.Picture
	var batch *pixel.Batch
	var frames []pixel.Rect
	switch spell {
	case "apoca":
		sheet = Pictures["./images/apocas.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		unorderedFrames := getFrames(sheet, 145, 145, 4, 4)
		for i := range unorderedFrames {
			frames = append(frames, unorderedFrames[ApocaFrames[i]])
		}
		break
	case "desca":
		sheet = Pictures["./images/desca.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		frames = getFrames(sheet, 127, 127, 5, 3)
	}

	return &SpellData{
		Caster:            caster,
		SpellName:         spell,
		Frames:            frames,
		Pic:               &sheet,
		Batch:             batch,
		CurrentAnimations: make([]*Spell, 0),
	}
}

type Spell struct {
	spellName   *string
	step        pixel.Rect
	frame       *pixel.Sprite
	frameNumber float64
	matrix      pixel.Matrix
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
	BodyFrames, HeadFrames []pixel.Rect
	BodyPic, HeadPic       *pixel.Picture
	BodyBatch, HeadBatch   *pixel.Batch
	CurrentAnimations      map[ksuid.KSUID]*Player
	AnimationsMutex        *sync.RWMutex
}

func NewPlayersData() PlayersData {
	bodySheet := Pictures["./images/bodies.png"]
	bodyBatch := pixel.NewBatch(&pixel.TrianglesData{}, bodySheet)
	bodyFrames := getFrames(bodySheet, 19, 38, 6, 4)

	headSheet := Pictures["./images/heads.png"]
	headBatch := pixel.NewBatch(&pixel.TrianglesData{}, headSheet)
	headFrames := getFrames(headSheet, 16, 16, 4, 0)

	return PlayersData{
		BodyFrames:        bodyFrames,
		HeadFrames:        headFrames,
		BodyPic:           &bodySheet,
		HeadPic:           &headSheet,
		BodyBatch:         bodyBatch,
		HeadBatch:         headBatch,
		CurrentAnimations: map[ksuid.KSUID]*Player{},
		AnimationsMutex:   &sync.RWMutex{},
	}
}
func GameUpdate(s *socket.Socket, pd *PlayersData, sd *SpellData) {
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
							np := NewPlayer()
							pd.CurrentAnimations[p.ID] = &np
							player, _ = pd.CurrentAnimations[p.ID]
						}
						pd.AnimationsMutex.Unlock()
						player.pos = pixel.V(p.X, p.Y)
						player.dir = p.Dir
						player.moving = p.Moving
					}
				}
				break
			case models.Spell:

				spell := models.SpellMsg{}
				json.Unmarshal(msg.Payload, &spell)

				pos := pixel.V(spell.X, spell.Y)
				newSpell := &Spell{
					spellName:   &sd.SpellName,
					step:        sd.Frames[0],
					frameNumber: 0.0,
					matrix:      pixel.IM.Scaled(pixel.ZV, .7).Moved(pos),
					last:        time.Now(),
				}

				newSpell.frame = pixel.NewSprite(*(sd.Pic), newSpell.step)
				sd.CurrentAnimations = append(sd.CurrentAnimations, newSpell)
				break
			}

		}
	}
}

func (pd *PlayersData) Draw(win *pixelgl.Window) {
	pd.BodyBatch.Clear()
	pd.HeadBatch.Clear()
	pd.AnimationsMutex.RLock()
	for _, p := range pd.CurrentAnimations {
		pd.AnimationsMutex.RUnlock()
		p.Update()
		p.body.Draw(pd.BodyBatch, p.bodyMatrix)
		p.head.Draw(pd.HeadBatch, p.headMatrix)
		pd.AnimationsMutex.RLock()
		//player.name.Draw(win, player.nameMatrix)
	}
	pd.AnimationsMutex.RUnlock()

	pd.BodyBatch.Draw(win)
	pd.HeadBatch.Draw(win)
}

type Player struct {
	pos                                pixel.Vec
	headPic, bodyPic                   *pixel.Picture
	name                               *text.Text
	head, body                         *pixel.Sprite
	headMatrix, bodyMatrix, nameMatrix pixel.Matrix
	bodyFrames, headFrames             []pixel.Rect
	playerUpdate                       *models.PlayerMsg
	dir                                string
	moving                             bool
	bodyFrame                          pixel.Rect
	headFrame                          pixel.Rect
	bodyStep                           float64
	last                               time.Time
}

func NewPlayer() Player {
	p := &Player{}
	basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	p.name = text.New(pixel.V(-28, 0), basicAtlas)
	p.name.Color = colornames.Blue
	fmt.Fprintln(p.name, "creative")

	bodySheet := Pictures["./images/bodies.png"]
	bodyFrames := getFrames(bodySheet, 19, 38, 6, 4)

	headSheet := Pictures["./images/heads.png"]
	headFrames := getFrames(headSheet, 16, 16, 4, 0)
	p.playerUpdate = &models.PlayerMsg{}
	p.last = time.Now()
	p.bodyFrames = bodyFrames
	p.headFrames = headFrames
	p.bodyPic = &bodySheet
	p.headPic = &headSheet
	p.dir = "down"
	p.pos = pixel.ZV
	return *p
}

func (p *Player) clientUpdate(s *socket.Socket) {
	p.playerUpdate = &models.PlayerMsg{
		ID:     s.ClientID,
		Name:   "name",
		X:      p.pos.X,
		Y:      p.pos.Y,
		Dir:    p.dir,
		Moving: p.moving,
	}
	playerMsg, err := json.Marshal(p.playerUpdate)
	if err != nil {
		return
	}
	p.playerUpdate = &models.PlayerMsg{}
	s.O <- models.NewMesg(models.UpdateServer, playerMsg)

}

func (p *Player) Update() *Player {
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
	return p
}

func (p *Player) getNextBodyFrame(dirFrames []int) pixel.Rect {
	dt := time.Since(p.last).Seconds()
	p.last = time.Now()
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
