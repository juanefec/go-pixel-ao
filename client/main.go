package main

import (
	"fmt"
	"image"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/juanefec/go-pixel-ao/client/socket"
	"github.com/segmentio/ksuid"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

var (
	camSpeed  = 200.0
	Zoom      = 1.0
	ZoomSpeed = 1.2
	frames    = 0
	second    = time.Tick(time.Second)
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
		Bounds: pixel.R(0, 0, 500, 500),
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

		frames++

		select {
		case <-second:

			win.SetTitle(fmt.Sprintf("%s | FPS: %d", cfg.Title, frames))
			frames = 0
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
	var apocaFrames []pixel.Rect
	for y := apocaSheet.Bounds().Min.Y; y < apocaSheet.Bounds().Max.Y; y += apocaSheet.Bounds().Max.Y / 4 {
		for x := apocaSheet.Bounds().Min.X; x < apocaSheet.Bounds().Max.X; x += apocaSheet.Bounds().Max.X / 4 {
			apocaFrames = append(apocaFrames, pixel.R(x, y, x+145, y+145))
		}
	}
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
	var bodyFrames []pixel.Rect
	bodyBatch := pixel.NewBatch(&pixel.TrianglesData{}, bodySheet)
	for y := bodySheet.Bounds().Min.Y; y < bodySheet.Bounds().Max.Y; y += bodySheet.Bounds().Max.Y / 4 {
		for x := bodySheet.Bounds().Min.X; x < bodySheet.Bounds().Max.X; x += bodySheet.Bounds().Max.X / 6 {
			bodyFrames = append(bodyFrames, pixel.R(x, y, x+19, y+38))
		}
	}

	headSheet := Pictures["./images/heads.png"]
	headBatch := pixel.NewBatch(&pixel.TrianglesData{}, headSheet)

	var headFrames []pixel.Rect
	for x := headSheet.Bounds().Min.X; x < headSheet.Bounds().Max.X; x += headSheet.Bounds().Max.X / 4 {
		headFrames = append(headFrames, pixel.R(x, 0, x+16, 16))
	}

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
			text := string(data)
			props := strings.Split(text, ";")
			if len(props) == 6 {
				id, _ := ksuid.Parse(props[0])
				x, _ := strconv.ParseFloat(props[2], 64)
				y, _ := strconv.ParseFloat(props[3], 64)
				pos := pixel.V(x, y)
				pd.AnimationsMutex.Lock()
				player, ok := pd.CurrentAnimations[id]
				if !ok {
					np := NewPlayer(&pos)
					np.bodyPic = pd.BodyPic
					np.headPic = pd.HeadPic
					np.bodyFrames = pd.BodyFrames
					np.headFrames = pd.HeadFrames
					pd.CurrentAnimations[id] = &np
					player, _ = pd.CurrentAnimations[id]
				}
				pd.AnimationsMutex.Unlock()
				player.updateOnline(props)

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
	var bodyFrames []pixel.Rect
	for y := bodySheet.Bounds().Min.Y; y < bodySheet.Bounds().Max.Y; y += bodySheet.Bounds().Max.Y / 4 {
		for x := bodySheet.Bounds().Min.X; x < bodySheet.Bounds().Max.X; x += bodySheet.Bounds().Max.X / 6 {
			bodyFrames = append(bodyFrames, pixel.R(x, y, x+19, y+38))
		}
	}

	headSheet := Pictures["./images/heads.png"]
	var headFrames []pixel.Rect
	for x := headSheet.Bounds().Min.X; x < headSheet.Bounds().Max.X; x += headSheet.Bounds().Max.X / 4 {
		headFrames = append(headFrames, pixel.R(x, 0, x+16, 16))
	}
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
	data := []string{}
	id := s.ClientID.String()
	data = []string{
		id,
		"name",
		fmt.Sprint(p.pos.X),
		fmt.Sprint(p.pos.Y),
		p.dir,
		fmt.Sprint(p.moving),
	}
	message := strings.Join(data, ";")
	s.O <- makeMessage([]byte(message))

}

func (p *Player) updateOnline(props []string) {
	posX, _ := strconv.ParseFloat(props[2], 64)
	posY, _ := strconv.ParseFloat(props[3], 64)
	p.pos = pixel.V(posX, posY)
	p.dir = props[4]
	p.moving, _ = strconv.ParseBool(props[5])
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
	var treeFrames []pixel.Rect
	for x := treeSheet.Bounds().Min.X; x < treeSheet.Bounds().Max.X; x += 32 {
		for y := treeSheet.Bounds().Min.Y; y < treeSheet.Bounds().Max.Y; y += 32 {
			treeFrames = append(treeFrames, pixel.R(x, y, x+32, y+32))
		}
	}
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

func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

func loadPictures(files ...string) map[string]pixel.Picture {
	var wg sync.WaitGroup
	var m sync.Mutex

	filesLength := len(files)
	contents := make(map[string]pixel.Picture, filesLength)
	wg.Add(filesLength)

	for _, file := range files {
		go func(file string) {
			content, err := loadPicture(file)

			if err != nil {
				log.Fatal(err)
			}

			m.Lock()
			contents[file] = content
			m.Unlock()
			wg.Done()
		}(file)
	}

	wg.Wait()

	return contents
}

func keyInputs(win *pixelgl.Window, player *Player) {
	last := time.Now()
	const (
		KeyUp    = pixelgl.KeyW
		KeyDown  = pixelgl.KeyS
		KeyLeft  = pixelgl.KeyA
		KeyRight = pixelgl.KeyD
	)

	timeMap := map[pixelgl.Button]int{
		KeyUp:    -1,
		KeyDown:  -1,
		KeyLeft:  -1,
		KeyRight: -1,
	}

	latestPressed := func(keyPressed pixelgl.Button, m map[pixelgl.Button]int) bool {
		var key pixelgl.Button
		min := 99999999999999
		for k, v := range m {
			if v < min && v > 0 {
				key = k
				min = v
			}
		}
		return key == keyPressed
	}

	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		if win.Pressed(KeyLeft) {
			if latestPressed(KeyLeft, timeMap) {
				player.moving = true
				player.dir = "left"
				player.pos.X -= camSpeed * dt
			}
			timeMap[KeyLeft]++
		} else {
			timeMap[KeyLeft] = -1
		}

		if win.Pressed(KeyRight) {
			if latestPressed(KeyRight, timeMap) {
				player.moving = true
				player.dir = "right"
				player.pos.X += camSpeed * dt
			}
			timeMap[KeyRight]++
		} else {
			timeMap[KeyRight] = -1
		}

		if win.Pressed(KeyDown) {
			if latestPressed(KeyDown, timeMap) {
				player.moving = true
				player.dir = "down"
				player.pos.Y -= camSpeed * dt
			}
			timeMap[KeyDown]++

		} else {
			timeMap[KeyDown] = -1
		}

		if win.Pressed(KeyUp) {
			if latestPressed(KeyUp, timeMap) {
				player.moving = true
				player.dir = "up"
				player.pos.Y += camSpeed * dt
			}
			timeMap[KeyUp]++
		} else {
			timeMap[KeyUp] = -1
		}

		if timeMap[KeyUp] == -1 && timeMap[KeyDown] == -1 && timeMap[KeyLeft] == -1 && timeMap[KeyRight] == -1 {
			player.moving = false
		}
	}
}
