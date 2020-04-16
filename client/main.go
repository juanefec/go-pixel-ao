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

func makeMessage(d []byte) []byte {
	d = append(d, Newline...)
	return d
}

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
	)

	player := NewPlayer(nil)
	apocaData := NewApocaData()
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
	go otherPlayers.playersUpdate(socket)

	for !win.Closed() {
		apocaData.Batch.Clear()

		cam := pixel.IM.Scaled(player.pos, Zoom).Moved(win.Bounds().Center().Sub(player.pos))
		win.SetMatrix(cam)
		win.Clear(colornames.Forestgreen)

		apocaData.Update(win, cam)
		player.Update()
		Zoom *= math.Pow(ZoomSpeed, win.MouseScroll().Y)
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

type ApocaData struct {
	Frames            []pixel.Rect
	Pic               *pixel.Picture
	Batch             *pixel.Batch
	CurrentAnimations []*Apoca
}

func (ad *ApocaData) Update(win *pixelgl.Window, cam pixel.Matrix) {
	if win.JustPressed(pixelgl.MouseButtonLeft) {
		mouse := cam.Unproject(win.MousePosition())
		newApoca := &Apoca{
			step:        ad.Frames[ApocaFrames[0]],
			frameNumber: 0.0,
			matrix:      pixel.IM.Scaled(pixel.ZV, .7).Moved(mouse),
			last:        time.Now(),
		}

		newApoca.frame = pixel.NewSprite(*(ad.Pic), newApoca.step)
		ad.CurrentAnimations = append(ad.CurrentAnimations, newApoca)

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

func NewApocaData() *ApocaData {

	apocaSheet := Pictures["./images/apocas.png"]
	batch := pixel.NewBatch(&pixel.TrianglesData{}, apocaSheet)
	apocaFrames := getFrames(apocaSheet, 145, 145, 4, 4)
	return &ApocaData{
		Frames:            apocaFrames,
		Pic:               &apocaSheet,
		Batch:             batch,
		CurrentAnimations: make([]*Apoca, 0),
	}
}

type Apoca struct {
	step        pixel.Rect
	frame       *pixel.Sprite
	frameNumber float64
	matrix      pixel.Matrix
	last        time.Time
}

func (a *Apoca) NextFrame(apocaFrames []pixel.Rect) (pixel.Rect, bool) {
	dt := time.Since(a.last).Seconds()
	a.last = time.Now()
	a.frameNumber += 21 * dt
	i := int(a.frameNumber)
	if i <= len(ApocaFrames)-1 {
		return apocaFrames[ApocaFrames[i]], false
	}
	a.frameNumber = .0
	return apocaFrames[ApocaFrames[0]], true
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
func (pd *PlayersData) playersUpdate(s *socket.Socket) {
	for {
		select {
		case data := <-s.I:
			var wupdate models.UpdateMessage
			err := json.Unmarshal(data, &wupdate)
			if err != nil {
				continue
			}
			for i := 0; i <= len(wupdate.Players)-1; i++ {
				pd.AnimationsMutex.Lock()
				p := wupdate.Players[i]
				if p.ID != s.ClientID {
					player, ok := pd.CurrentAnimations[p.ID]
					if !ok {
						pos := pixel.V(p.X, p.Y)
						np := NewPlayer(&pos)
						np.bodyPic = pd.BodyPic
						np.headPic = pd.HeadPic
						np.bodyFrames = pd.BodyFrames
						np.headFrames = pd.HeadFrames
						pd.CurrentAnimations[p.ID] = &np
						player, _ = pd.CurrentAnimations[p.ID]
					}
					player.pos = pixel.V(p.X, p.Y)
					player.dir = p.Dir
					player.moving = p.Moving
				}
				pd.AnimationsMutex.Unlock()
			}

		}
	}
}

func (pd *PlayersData) Draw(win *pixelgl.Window) {
	pd.BodyBatch.Clear()
	pd.HeadBatch.Clear()
	pd.AnimationsMutex.Lock()
	for key := range pd.CurrentAnimations {
		player := pd.CurrentAnimations[key]
		player.Update()
		player.body.Draw(pd.BodyBatch, player.bodyMatrix)
		player.head.Draw(pd.HeadBatch, player.headMatrix)
		//player.name.Draw(win, player.nameMatrix)
	}
	pd.AnimationsMutex.Unlock()

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
	dir                                string
	moving                             bool
	bodyFrame                          pixel.Rect
	headFrame                          pixel.Rect
	bodyStep                           float64
	last                               time.Time
}

func NewPlayer(pos *pixel.Vec) Player {
	p := &Player{}
	basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	p.name = text.New(pixel.V(-28, 0), basicAtlas)
	p.name.Color = colornames.Blue
	fmt.Fprintln(p.name, "creative")

	bodySheet := Pictures["./images/bodies.png"]
	bodyFrames := getFrames(bodySheet, 19, 38, 6, 4)

	headSheet := Pictures["./images/heads.png"]
	headFrames := getFrames(headSheet, 16, 16, 4, 0)

	p.last = time.Now()
	p.bodyFrames = bodyFrames
	p.headFrames = headFrames
	p.bodyPic = &bodySheet
	p.headPic = &headSheet
	p.dir = "down"
	p.pos = pixel.ZV
	if pos != nil {
		p.pos = *pos
	}
	return *p
}

func (p *Player) clientUpdate(s *socket.Socket) {
	player := models.PlayerMsg{
		ID:     s.ClientID,
		Name:   "name",
		X:      p.pos.X,
		Y:      p.pos.Y,
		Dir:    p.dir,
		Moving: p.moving,
	}
	data := models.PlayerMessage{
		Player: player,
	}
	msg, err := json.Marshal(data)
	if err != nil {
		return
	}
	s.O <- makeMessage(msg)

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
		p.bodyStep = 0
		return p.bodyFrames[dirFrames[0]]
	}
	p.bodyStep = 0
	return p.bodyFrames[dirFrames[0]]
}

type Forest struct {
	Pic    pixel.Picture
	Frames []pixel.Rect
	Batch  *pixel.Batch
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
	return &Forest{
		Pic:    treeSheet,
		Frames: treeFrames,
		Batch:  treeBatch,
	}
}
