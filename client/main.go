package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"time"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/juanefec/go-pixel-ao/client/socket"
	"github.com/juanefec/go-pixel-ao/models"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

const (
	PlayerBaseSpeed = 185.0
)

var (
	FireballSpeed      = 290.0
	Zoom               = 2.0
	ZoomSpeed          = 1.1
	second             = time.Tick(time.Second)
	MaxMana            = 2324.0
	MaxHealth          = 347.0
	OnTargetSpellRange = 450.0
	AOESpellRange      = 700.0
	TrapSpellRange     = 100.0
	FlashSpellRange    = 200.0
	//Spell intervals
	BasicSpellInterval    = (time.Second.Seconds() / 10) * 9
	FireballSpellInterval = (time.Second.Seconds() / 10) * 7
	IcesnipeSpellInterval = (time.Second.Seconds() / 10) * 8
	RockSpellInterval     = time.Second.Seconds() * 8
	LavaSpellInterval     = time.Second.Seconds() * 14
	ManaSpotSpellInterval = time.Second.Seconds() * 16
	FlashSpellInterval    = time.Second.Seconds() * 10
	TrapsChargeInterval   = time.Second.Seconds()
	FlashChargeInterval   = time.Second.Seconds() * 6

	ArrowMaxCharge = time.Second.Seconds() * 2.5
	// Ranking
	Ranking = []models.RankingPosMsg{}
)
var (
	Newline   = []byte{'\n'}
	BodyUp    = []int{12, 13, 14, 15, 16, 17}
	BodyDown  = []int{18, 19, 20, 21, 22, 23}
	BodyLeft  = []int{6, 7, 8, 9, 10}
	BodyRight = []int{0, 1, 2, 3, 4}
	DeadUp    = []int{6, 7, 8}
	DeadDown  = []int{9, 10, 11}
	DeadLeft  = []int{3, 4, 5}
	DeadRight = []int{0, 1, 2}

	ApocaFrames       = []int{12, 13, 14, 15, 8, 9, 10, 11, 4, 5, 6, 7, 0, 1, 2, 3}
	BloodFrames       = []int{18, 19, 20, 21, 22, 23, 12, 13, 14, 15, 16, 17, 6, 7, 8, 9, 10, 11, 0, 1, 2, 3, 4, 5}
	HealingShotFrames = []int{48, 49, 50, 51, 52, 53, 54, 55, 40, 41, 42, 43, 44, 45, 46, 47, 32, 33, 34, 35, 36, 37, 38, 39, 24, 25, 26, 27, 28, 29, 30, 31, 16, 17, 18, 19, 20, 21, 22, 23, 8, 9, 10, 11, 12, 13, 14, 15, 0, 1, 2, 3, 4, 5, 6, 7}
	RockFrames     = []int{56, 57, 58, 59, 60, 61, 62, 63, 48, 49, 50, 51, 52, 53, 54, 55, 40, 41, 42, 43, 44, 45, 46, 47, 32, 33, 34, 35, 36, 37, 38, 39, 24, 25, 26, 27, 28, 29, 30, 31, 16, 17, 18, 19, 20, 21, 22, 23, 8, 9, 10, 11, 12, 13, 14, 15, 0, 1, 2, 3, 4, 5, 6, 7}
	ArrowHitFrames = []int{12, 13, 14, 15, 8, 9, 10, 11, 4, 5, 6, 7, 0, 1, 2, 3}
	
	Pictures       map[string]pixel.Picture
	Key            KeyConfig
)
var basicAtlas = text.NewAtlas(basicfont.Face7x13, text.ASCII)
var chatlog = NewChatlog()

func main() {
	pixelgl.Run(run)
}

func run() {
	Pictures = loadPictures(
		"./images/apocas.png",
		"./images/bodydruida.png",
		"./images/bodydruidaIcon.png",
		"./images/heads.png",
		"./images/trees.png",
		"./images/dead.png",
		"./images/deadHead.png",
		"./images/desca.png",
		"./images/curaBody.png",
		"./images/curaHead.png",
		"./images/arbolmuerto.png",
		"./images/newGrass.png",
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
		"./images/smallExplosion.png",
		"./images/talltree.png",
		"./images/tallnoleafstree.png",
		"./images/darkopshit.png",
		"./images/darkopshitIcon.png",
		"./images/placaazul.png",
		"./images/placaazulIcon.png",
		"./images/penumbras.png",
		"./images/penumbrasIcon.png",
		"./images/gameIcon.png",
		"./images/icesnipe.png",
		"./images/blood.png",
		"./images/apocaIcon.png",
		"./images/descaIcon.png",
		"./images/exploIcon.png",
		"./images/fireballIcon.png",
		"./images/icesnipeIcon.png",
		"./images/lavaSpot.png",
		"./images/lavaSpotIcon.png",
		"./images/healingSpot.png",
		"./images/healingSpotIcon.png",
		"./images/healingShot.png",
		"./images/healingShotIcon.png",
		"./images/manaSpot.png",
		"./images/manaSpotIcon.png",
		"./images/manaShot.png",
		"./images/manaShotIcon.png",
		"./images/smokeSpot.png",
		"./images/smokeSpotIcon.png",
		"./images/flashEffect.png",
		"./images/flashEffectIcon.png",
		"./images/rockShot.png",
		"./images/rockShotIcon.png",
		"./images/arrowShot.png",
		"./images/arrowShotIcon.png",
		"./images/arrowExplo.png",
		"./images/hunterTrap.png",
		"./images/hunterTrapIcon.png",
	)
	
	rawConfig, err := ioutil.ReadFile("./key-config.json")
	if err != nil {
		panic(err)
	}

	Key = KeyConfig{}
	err = json.Unmarshal(rawConfig, &Key)
	if err != nil {
		panic(err)
	}

	ld, err := LoginWindow()
	if err != nil {
		log.Fatal(err)
	}
	player := NewPlayer(ld.Name, &ld)

	allSpells, allEffects := InitSpells(&player)

	forest := NewForest()
	//buda := NewBuda(pixel.V(2000, 3400))
	otherPlayers := NewPlayersData()
	playerInfo := NewPlayerInfo(&player, &otherPlayers, allSpells)
	resu := NewResu(pixel.V(2000, 2900))

	socket := socket.NewSocket("localhost", 33333)
	defer socket.Close()

	cfg := pixelgl.WindowConfig{
		Title: "Creative AO",
		//Monitor: pixelgl.PrimaryMonitor(),
		Bounds: pixel.R(0, 0, 1360, 840),
		Icon:   []pixel.Picture{Pictures["./images/gameIcon.png"]},
		//VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	cursor := NewCursor(win)
	go keyInputs(win, &player, cursor)
	go GameUpdate(socket, &otherPlayers, &player, allSpells)
	fps := 0
	unatachedCam := false
	offset := pixel.ZV
	newCenter := pixel.ZV
	for !win.Closed() {
		win.Clear(colornames.Forestgreen)
		cam := pixel.IM.Scaled(player.pos, Zoom).Moved(win.Bounds().Center().Sub(player.pos))

		player.cam = cam

		if win.Pressed(pixelgl.KeyLeftShift) {

			if win.Pressed(pixelgl.MouseButtonRight) {

				if !unatachedCam {
					unatachedCam = true
					offset = cam.Unproject(win.MousePosition()).Sub(player.pos)
					newCenter = cam.Unproject(win.MousePosition()).Sub(newCenter).Sub(offset)

				}
				newCenter = cam.Unproject(win.MousePosition()).Sub(player.pos).Sub(offset)
			} else {
				unatachedCam = false
			}
			cam = cam.Moved(newCenter)
		} else {
			unatachedCam = false
		}
		win.SetMatrix(cam)
		forest.GrassBatch.Draw(win)
		forest.FenceBatchHTOP.Draw(win)
		resu.Draw(win, cam, &player)
		otherPlayers.Draw(win, &player)
		player.Draw(win, socket)
		player.DrawIngameHud(win)
		//buda.Draw(win)
		forest.Batch.Draw(win)
		forest.Trees.Draw(win)
		forest.FenceBatchV.Draw(win)
		forest.FenceBatchHBOT.Draw(win)
		allSpells.Draw(win, cam, socket, &otherPlayers, cursor, allEffects)
		allEffects.Draw(win, cam, socket, &otherPlayers, cursor)
		playerInfo.Draw(win, cam, cursor, &ld)
		chatlog.Draw(win, cam)
		cursor.Draw(cam, player.pos)

		fps++
		if !player.chat.chatting && win.JustPressed(pixelgl.KeyZ) {
			if Zoom == 2 {
				Zoom = 1
			} else {
				Zoom = 2
			}
		}
		select {
		case <-second:
			playerInfo.nfps = fps
			fps = 0
		default:
		}
		win.Update()
		player.clientUpdate(socket)
	}
}

type WizardType int

const (
	DarkWizard WizardType = iota //tuni negra
	Monk                         // la celeste
	Shaman                       // el druida
	Sniper                       //
	Timewreker                   //
	Hunter                       // el rojo
	Igniter                      // penumbras
)

type Wizard struct {
	Name          string
	Skin          SkinType
	Type          WizardType
	SpecialSpells []string
}

func SendDeathEvent(s *socket.Socket, d models.DeathMsg) {
	dmm, _ := json.Marshal(d)
	s.O <- models.NewMesg(models.Death, dmm)
}

func GameUpdate(s *socket.Socket, pd *PlayersData, p *Player, spells GameSpells) {
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
							wiz := Wizard{
								Skin: SkinType(p.Skin),
							}
							np := NewPlayer(p.Name, &wiz)
							pd.CurrentAnimations[p.ID] = &np
							player, _ = pd.CurrentAnimations[p.ID]
						}
						pd.AnimationsMutex.Unlock()
						player.pos = pixel.V(p.X, p.Y)
						player.dir = p.Dir
						player.moving = p.Moving
						player.dead = p.Dead
						player.hp = p.HP
						player.invisible = p.Invisible
					}
				}
				break
			case models.Spell:

				spell := models.SpellMsg{}
				json.Unmarshal(msg.Payload, &spell)
				now := time.Now()
				newSpell := &Spell{
					spellName:      &spell.SpellName,
					frameNumber:    0.0,
					last:           now,
					projectileLife: now,
					damageInterval: now,
				}

				target := &Player{}
				if spell.SpellType == "on-target" {
					if s.ClientID == spell.TargetID {
						target = p
					} else {
						target = pd.CurrentAnimations[spell.TargetID]
					}
					newSpell.target = target
					newSpell.matrix = &target.headMatrix
				}
				for i := range spells {
					if spells[i].Name() == spell.SpellName {
						spells[i].UpdateFromServer(newSpell, pd, spell, s)
					}
				}

			case models.Chat:
				chatMsg := models.ChatMsg{}
				json.Unmarshal(msg.Payload, &chatMsg)
				pd.CurrentAnimations[chatMsg.ID].chat.WriteSent(chatMsg.ID, chatMsg.Name, chatMsg.Message)
			case models.UpdateRanking:
				rankingMsg := []models.RankingPosMsg{}
				json.Unmarshal(msg.Payload, &rankingMsg)
				Ranking = rankingMsg
				for i := range Ranking {
					if Ranking[i].ID == s.ClientID {
						p.kills = Ranking[i].K
						p.deaths = Ranking[i].D
					}
				}
			case models.Disconect:
				m := models.DisconectMsg{}
				json.Unmarshal(msg.Payload, &m)
				if _, exist := pd.CurrentAnimations[m.ID]; exist {
					pd.Online--
					pd.AnimationsMutex.Lock()
					delete(pd.CurrentAnimations, m.ID)
					pd.AnimationsMutex.Unlock()
				}
			}

		}
	}
}
