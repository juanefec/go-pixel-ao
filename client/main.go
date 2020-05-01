package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"image/color"
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

const (
	PlayerBaseSpeed = 185.0
)

var (
	PlayerSpeed        = 185.0
	FireballSpeed      = 290.0
	Zoom               = 2.0
	ZoomSpeed          = 1.1
	second             = time.Tick(time.Second)
	MaxMana            = 2324
	MaxHealth          = 347
	OnTargetSpellRange = 400.0
	FlashSpellRange    = 200.0
	//Spell intervals
	BasicSpellInterval    = (time.Second.Seconds() / 10) * 9
	FireballSpellInterval = (time.Second.Seconds() / 10) * 7
	IcesnipeSpellInterval = (time.Second.Seconds() / 10) * 8
	LavaSpellInterval     = time.Second.Seconds() * 14
	FlashSpellInterval    = time.Second.Seconds() * 6
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
	HealingShotFrames = []int{
		48, 49, 50, 51, 52, 53, 54, 55,
		40, 41, 42, 43, 44, 45, 46, 47,
		32, 33, 34, 35, 36, 37, 38, 39,
		24, 25, 26, 27, 28, 29, 30, 31,
		16, 17, 18, 19, 20, 21, 22, 23,
		8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7}
	RockFrames = []int{56, 57, 58, 59, 60, 61, 62, 63, 48, 49, 50, 51, 52, 53, 54, 55, 40, 41, 42, 43, 44, 45, 46, 47, 32, 33, 34, 35, 36, 37, 38, 39, 24, 25, 26, 27, 28, 29, 30, 31, 16, 17, 18, 19, 20, 21, 22, 23, 8, 9, 10, 11, 12, 13, 14, 15, 0, 1, 2, 3, 4, 5, 6, 7}

	Pictures map[string]pixel.Picture
	Key      KeyConfig
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
	allSpells := SpellKinds{
		OnTarget: GameSpells{
			NewSpellData("apoca", &player),
			NewSpellData("desca", &player),
			NewSpellData("explo", &player),
		},
		Projectile: GameSpells{
			NewSpellData("fireball", &player),
			NewSpellData("icesnipe", &player),
			NewSpellData("healshot", &player),
			NewSpellData("manashot", &player),
			NewSpellData("rockshot", &player),
		},
		AOE: GameSpells{
			NewSpellData("lava-spot", &player),
			NewSpellData("heal-spot", &player),
			NewSpellData("smoke-spot", &player),
			NewSpellData("mana-spot", &player),
		},
		Movement: GameSpells{
			NewSpellData("flash", &player),
		},
		Effects: GameSpells{
			NewSpellData("mini-explo", &player),
			NewSpellData("blood-explo", &player),
		},
	}

	forest := NewForest()
	//buda := NewBuda(pixel.V(2000, 3400))
	otherPlayers := NewPlayersData()
	playerInfo := NewPlayerInfo(&player, &otherPlayers)
	resu := NewResu(pixel.V(2000, 2900))

	socket := socket.NewSocket("190.247.147.18", 33333)
	defer socket.Close()

	cfg := pixelgl.WindowConfig{
		Title: "Creative AO",
		//Monitor: pixelgl.PrimaryMonitor(),
		Bounds: pixel.R(0, 0, 1360, 840),
		Icon:   []pixel.Picture{Pictures["./images/gameIcon.png"]},
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	cursor := NewCursor(win)
	go keyInputs(win, &player, cursor)
	go GameUpdate(socket, &otherPlayers, &player, allSpells)
	fps := 0
	for !win.Closed() {
		win.Clear(colornames.Black)
		cam := pixel.IM.Scaled(player.pos, Zoom).Moved(win.Bounds().Center().Sub(player.pos))
		player.cam = cam
		win.SetMatrix(cam)

		forest.GrassBatch.Draw(win)
		forest.FenceBatchHTOP.Draw(win)
		resu.Draw(win, cam, &player)
		otherPlayers.Draw(win)
		player.Draw(win, socket)
		//buda.Draw(win)
		forest.Batch.Draw(win)
		forest.Trees.Draw(win)
		forest.FenceBatchV.Draw(win)
		forest.FenceBatchHBOT.Draw(win)
		allSpells.Draw(win, cam, socket, &otherPlayers, cursor)
		playerInfo.Draw(win, cam, cursor, &ld)
		cursor.Draw(cam)
		fps++

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

func main() {
	pixelgl.Run(run)
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
	Intervals     []float64
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

type SpellKinds struct {
	OnTarget,
	Projectile,
	AOE,
	Movement,
	Effects GameSpells
}

func (sk *SpellKinds) Draw(win *pixelgl.Window, cam pixel.Matrix, s *socket.Socket, pd *PlayersData, cursor *Cursor) {
	sk.OnTarget.Draw(win, cam, s, pd, cursor)
	sk.Projectile.Draw(win, cam, s, pd, cursor, sk.Effects...)
	sk.Effects.Draw(win, cam, s, pd, cursor)
	sk.AOE.Draw(win, cam, s, pd, cursor)
	sk.Movement.Draw(win, cam, s, pd, cursor)
}

type GameSpells []*SpellData

func (gs GameSpells) Draw(win *pixelgl.Window, cam pixel.Matrix, s *socket.Socket, pd *PlayersData, cursor *Cursor, effects ...*SpellData) {
	switch gs[0].SpellType {
	case "on-target":
		for i := range gs {
			gs[i].Batch.Clear()
			gs[i].UpdateOnTarget(win, cam, s, pd, cursor)
			gs[i].Batch.Draw(win)
		}
	case "projectile":
		for i := range gs {
			gs[i].Batch.Clear()
			gs[i].UpdateProjectiles(win, cam, s, pd, cursor, effects...)
			gs[i].Batch.Draw(win)
		}
	case "aoe":
		for i := range gs {
			gs[i].Batch.Clear()
			gs[i].UpdateAOE(win, cam, s, pd, cursor)
			gs[i].Batch.Draw(win)
		}
	case "movement":
		for i := range gs {
			gs[i].Batch.Clear()
			gs[i].UpdateMovement(win, cam, s, pd, cursor)
			gs[i].Batch.Draw(win)
		}

	}

}

type SpellData struct {
	EffectRadius      float64
	SpellLifespawn    float64
	Interval          float64
	SpellType         string
	SpellName         string
	SpellMode         CursorMode
	WizardCaster      WizardType
	ManaCost, Damage  int
	SpellSpeed        float64
	ScaleF            float64
	ProjSpeed         float64
	Caster            *Player
	Frames            []pixel.Rect
	Pic               *pixel.Picture
	Batch             *pixel.Batch
	CurrentAnimations []*Spell
}

func (sd *SpellData) UpdateOnTarget(win *pixelgl.Window, cam pixel.Matrix, s *socket.Socket, pd *PlayersData, cursor *Cursor) {
	dt := time.Since(sd.Caster.lastCast).Seconds()
	if !sd.Caster.chat.chatting && win.JustPressed(pixelgl.MouseButtonLeft) && !sd.Caster.dead && sd.Caster.mp >= sd.ManaCost && cursor.Mode == sd.SpellMode && Normal != sd.SpellMode {
		if dt >= sd.Interval {
			sd.Caster.lastCast = time.Now()
			mouse := cam.Unproject(win.MousePosition())
			if Dist(mouse, cam.Unproject(win.Bounds().Center())) <= OnTargetSpellRange {
				for key := range pd.CurrentAnimations {

					if !pd.CurrentAnimations[key].dead && cursor.Mode != SpellCastPrimarySkill && pd.CurrentAnimations[key].OnMe(mouse) {
						spell := models.SpellMsg{
							ID:        s.ClientID,
							SpellType: sd.SpellType,
							SpellName: sd.SpellName,
							TargetID:  key,
							Name:      sd.Caster.sname,
							X:         mouse.X,
							Y:         mouse.Y,
						}
						paylaod, _ := json.Marshal(spell)
						s.O <- models.NewMesg(models.Spell, paylaod)

						sd.Caster.mp -= sd.ManaCost
						newSpell := &Spell{
							target:      pd.CurrentAnimations[key],
							spellName:   &sd.SpellName,
							step:        sd.Frames[0],
							frameNumber: 0.0,
							matrix:      &pd.CurrentAnimations[key].headMatrix,
							last:        time.Now(),
						}

						newSpell.frame = pixel.NewSprite(*(sd.Pic), newSpell.step)
						sd.CurrentAnimations = append(sd.CurrentAnimations, newSpell)

						break
					}
				}
			}
			cursor.SetNormalMode()
		}
	}

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
		sd.CurrentAnimations[i].frame.Draw(sd.Batch, (*sd.CurrentAnimations[i].matrix).Scaled(sd.CurrentAnimations[i].target.pos, sd.ScaleF))
	}

}

func (sd *SpellData) UpdateProjectiles(win *pixelgl.Window, cam pixel.Matrix, s *socket.Socket, pd *PlayersData, cursor *Cursor, effects ...*SpellData) {
	dtproj := time.Since(sd.Caster.lastCastPrimary).Seconds()

	if !sd.Caster.chat.chatting && (win.JustPressed(pixelgl.Button(Key.FireB)) && sd.Caster.wizard.Type == sd.WizardCaster) && !sd.Caster.dead && sd.Caster.mp >= sd.ManaCost {
		if dtproj >= sd.Interval {
			sd.Caster.lastCastPrimary = time.Now()
			mouse := cam.Unproject(win.MousePosition())
			spell := models.SpellMsg{
				ID:        s.ClientID,
				SpellType: sd.SpellType,
				SpellName: sd.SpellName,
				TargetID:  ksuid.Nil,
				Name:      sd.Caster.sname,
				X:         mouse.X,
				Y:         mouse.Y,
			}
			paylaod, _ := json.Marshal(spell)
			s.O <- models.NewMesg(models.Spell, paylaod)

			projectedCenter := cam.Unproject(win.Bounds().Center())
			vel := mouse.Sub(projectedCenter)
			centerMatrix := pixel.IM
			switch sd.SpellName {
			case "fireball":
				centerMatrix = sd.Caster.bodyMatrix.Rotated(projectedCenter, vel.Angle()+(math.Pi/2)).Scaled(projectedCenter, 2)
			case "icesnipe":
				centerMatrix = sd.Caster.bodyMatrix.Rotated(projectedCenter, vel.Angle()).Scaled(projectedCenter, .6)
			case "healshot", "manashot":
				centerMatrix = sd.Caster.bodyMatrix.Rotated(projectedCenter, vel.Angle()+(math.Pi/2)).Scaled(projectedCenter, .6)
			case "rockshot":
				centerMatrix = sd.Caster.bodyMatrix.Rotated(projectedCenter, vel.Angle())
			}
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
		}
	}

FBALLS:
	for i := 0; i <= len(sd.CurrentAnimations)-1; i++ {
		next, kill := sd.CurrentAnimations[i].NextFrameFireball(sd.Frames, sd.ProjSpeed, sd.SpellLifespawn)
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
				effect := &Spell{
					target:      p,
					caster:      s.ClientID,
					spellName:   &sd.SpellName,
					step:        sd.Frames[0],
					frameNumber: 0.0,
					matrix:      &p.bodyMatrix,
					last:        time.Now(),
				}
				for i := range effects {
					if ("mini-explo" == effects[i].SpellName && sd.SpellName == "fireball") || ("blood-explo" == effects[i].SpellName && sd.SpellName == "icesnipe") {
						effects[i].CurrentAnimations = append(effects[i].CurrentAnimations, effect)
					}
				}

				p.hp -= sd.Damage
				if p.hp <= 0 {
					p.hp = 0
					p.dead = true
				}
				if p.hp > MaxHealth {
					p.hp = MaxHealth
				}
				continue FBALLS
			}
		}
		if sd.CurrentAnimations[i].caster != s.ClientID && !sd.Caster.dead && sd.Caster.OnMe(sd.CurrentAnimations[i].pos) {
			casterID := sd.CurrentAnimations[i].caster
			effect := &Spell{
				target:      sd.Caster,
				caster:      s.ClientID,
				spellName:   &sd.SpellName,
				step:        sd.Frames[0],
				frameNumber: 0.0,
				matrix:      &sd.Caster.bodyMatrix,
				last:        time.Now(),
			}
			for i := range effects {
				if "mini-explo" == effects[i].SpellName && sd.SpellName == "fireball" {
					sd.Caster.hp -= sd.Damage
					effects[i].CurrentAnimations = append(effects[i].CurrentAnimations, effect)
				}
				if "blood-explo" == effects[i].SpellName && sd.SpellName == "icesnipe" {
					sd.Caster.hp -= int(Map(Dist(sd.Caster.pos, pd.CurrentAnimations[casterID].pos), 0, 500, 15, float64(sd.Damage)))
					effects[i].CurrentAnimations = append(effects[i].CurrentAnimations, effect)
				}

			}
			if sd.SpellName == "rockshot" {
				sd.Caster.lastRootedStart = time.Now()
				sd.Caster.rooted = true
				sd.Caster.hp -= sd.Damage
			}
			if sd.SpellName == "healshot" {
				sd.Caster.hp -= sd.Damage
			}
			if sd.SpellName == "manashot" {
				sd.Caster.mp -= sd.Damage
			}
			if sd.Caster.hp <= 0 {
				sd.Caster.hp = 0
				sd.Caster.dead = true
			}
			if sd.Caster.hp > MaxHealth {
				sd.Caster.hp = MaxHealth
			}
			if sd.Caster.mp < 0 {
				sd.Caster.mp = 0
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
		sd.CurrentAnimations[i].frame.Draw(sd.Batch, (*sd.CurrentAnimations[i].matrix).Scaled(sd.CurrentAnimations[i].pos, sd.ScaleF))
	}

}

func (sd *SpellData) UpdateAOE(win *pixelgl.Window, cam pixel.Matrix, s *socket.Socket, pd *PlayersData, cursor *Cursor) {
	if (sd.Caster.wizard.Type == Shaman && sd.SpellName == "mana-spot") ||
		(sd.Caster.wizard.Type == Monk && sd.SpellName == "heal-spot") ||
		(sd.Caster.wizard.Type == DarkWizard && sd.SpellName == "lava-spot") ||
		(sd.Caster.wizard.Type == Sniper && sd.SpellName == "smoke-spot") {
		dt := time.Since(sd.Caster.lastCastSecondary).Seconds()
		if !sd.Caster.chat.chatting && win.JustPressed(pixelgl.Button(Key.IceSnipe)) && !sd.Caster.dead && sd.Caster.mp >= sd.ManaCost {
			if dt >= sd.Interval {
				sd.Caster.lastCastSecondary = time.Now()
				mouse := cam.Unproject(win.MousePosition())
				if Dist(mouse, cam.Unproject(win.Bounds().Center())) <= OnTargetSpellRange {

					spell := models.SpellMsg{
						ID:        s.ClientID,
						SpellType: sd.SpellType,
						SpellName: sd.SpellName,
						TargetID:  ksuid.Nil,
						Name:      sd.Caster.sname,
						X:         mouse.X,
						Y:         mouse.Y,
					}
					paylaod, _ := json.Marshal(spell)
					s.O <- models.NewMesg(models.Spell, paylaod)

					spellMatrix := pixel.IM.Moved(mouse)
					sd.Caster.mp -= sd.ManaCost
					newSpell := &Spell{
						projectileLife: time.Now(),
						pos:            mouse,
						spellName:      &sd.SpellName,
						step:           sd.Frames[0],
						frameNumber:    0.0,
						matrix:         &spellMatrix,
						last:           time.Now(),
						caster:         s.ClientID,
						damageInterval: time.Now(),
					}

					newSpell.frame = pixel.NewSprite(*(sd.Pic), newSpell.step)
					sd.CurrentAnimations = append(sd.CurrentAnimations, newSpell)

				}
			}
		}
	}
	for i := 0; i <= len(sd.CurrentAnimations)-1; i++ {
		next, kill := sd.CurrentAnimations[i].NextFrameFireball(sd.Frames, sd.ProjSpeed, sd.SpellLifespawn)
		if kill {
			if i < len(sd.CurrentAnimations)-1 {
				copy(sd.CurrentAnimations[i:], sd.CurrentAnimations[i+1:])
			}
			sd.CurrentAnimations[len(sd.CurrentAnimations)-1] = nil // or the zero sd.vCurrentAnimationslue of T
			sd.CurrentAnimations = sd.CurrentAnimations[:len(sd.CurrentAnimations)-1]
			continue
		}
		if !sd.Caster.dead && sd.Caster.InsideRaduis(sd.CurrentAnimations[i].pos, sd.EffectRadius) {
			switch sd.SpellName {
			case "lava-spot", "heal-spot", "mana-spot":
				dt := time.Since(sd.CurrentAnimations[i].damageInterval).Seconds()
				sd.CurrentAnimations[i].damageInterval = time.Now()
				if sd.WizardCaster == Shaman {
					println(int(float64(sd.Damage) * dt))
					sd.Caster.mp -= int(float64(sd.Damage) * dt)
					if sd.Caster.mp > MaxMana {
						sd.Caster.mp = MaxMana
					}
				} else if sd.WizardCaster == Monk {
					println(int(float64(sd.Damage) * dt))
					sd.Caster.hp -= int(float64(sd.Damage) * dt)
					if sd.Caster.hp > MaxHealth {
						sd.Caster.hp = MaxHealth
					}
				} else {
					if sd.CurrentAnimations[i].caster != s.ClientID {
						sd.Caster.hp -= int(float64(sd.Damage) * dt)
						if sd.Caster.hp <= 0 {
							sd.Caster.hp = 0
							sd.Caster.dead = true
						}
					}
				}

			case "smoke-spot":
				sd.Caster.invisible = true
				sd.Caster.inviEffectOut = time.Now()
			}
		} else {
			sd.CurrentAnimations[i].damageInterval = time.Now()
		}
		sd.CurrentAnimations[i].step = next
		sd.CurrentAnimations[i].frame = pixel.NewSprite(*sd.Pic, sd.CurrentAnimations[i].step)
		sd.CurrentAnimations[i].frame.Draw(sd.Batch, (*sd.CurrentAnimations[i].matrix).Scaled(sd.CurrentAnimations[i].pos, sd.ScaleF))
	}

}

func (sd *SpellData) UpdateMovement(win *pixelgl.Window, cam pixel.Matrix, s *socket.Socket, pd *PlayersData, cursor *Cursor) {
	if sd.Caster.wizard.Type == Timewreker && sd.SpellName == "flash" {
		dt := time.Since(sd.Caster.lastCastSecondary).Seconds()
		if !sd.Caster.chat.chatting && win.JustPressed(pixelgl.Button(Key.IceSnipe)) && !sd.Caster.dead && sd.Caster.mp >= sd.ManaCost {
			if dt >= sd.Interval {
				sd.Caster.lastCastSecondary = time.Now()
				mouse := cam.Unproject(win.MousePosition())
				spell := models.SpellMsg{
					ID:        s.ClientID,
					SpellType: sd.SpellType,
					SpellName: sd.SpellName,
					TargetID:  ksuid.Nil,
					Name:      sd.Caster.sname,
					X:         mouse.X,
					Y:         mouse.Y,
				}
				paylaod, _ := json.Marshal(spell)
				s.O <- models.NewMesg(models.Spell, paylaod)

				spellMatrix := pixel.IM.Moved(sd.Caster.pos)
				sd.Caster.mp -= sd.ManaCost
				newSpell := &Spell{
					pos:         sd.Caster.pos,
					spellName:   &sd.SpellName,
					step:        sd.Frames[0],
					frameNumber: 0.0,
					matrix:      &spellMatrix,
					last:        time.Now(),
					caster:      s.ClientID,
				}

				newSpell.frame = pixel.NewSprite(*(sd.Pic), newSpell.step)
				sd.CurrentAnimations = append(sd.CurrentAnimations, newSpell)
				dist := Dist(mouse, cam.Unproject(win.Bounds().Center()))
				if dist <= FlashSpellRange {
					sd.Caster.pos = mouse
				} else {
					nm := VectorNormalize(mouse.Sub(cam.Unproject(win.Bounds().Center())))
					sd.Caster.pos = nm.Scaled(FlashSpellRange).Add(cam.Unproject(win.Bounds().Center()))
				}

			}
		}
	}
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
		sd.CurrentAnimations[i].frame.Draw(sd.Batch, (*sd.CurrentAnimations[i].matrix).Scaled(sd.CurrentAnimations[i].pos, sd.ScaleF))
	}

}

func Dist(v1, v2 pixel.Vec) float64 {
	w, h := math.Abs(v1.X-v2.X), math.Abs(v1.Y-v2.Y)
	return math.Sqrt(math.Pow(w, 2) + math.Pow(h, 2))
}

func NewSpellData(spell string, caster *Player) *SpellData {

	var sheet pixel.Picture
	var batch *pixel.Batch
	var frames []pixel.Rect
	var mode CursorMode
	var manaCost, damage int
	var speed float64 = 21
	var scalef = .8
	var spellspeed = .0
	var spellType = "on-target"
	var interval = BasicSpellInterval
	var lifespawn = 1.0
	var efectRaduis = 70.0
	var casterType WizardType
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
		damage = 190
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
		manaCost = 1600
		damage = 220
		speed = 17
		scalef = 1.2
	case "fireball":
		casterType = DarkWizard
		sheet = Pictures["./images/fireball.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		frames = getFrames(sheet, 24, 24, 7, 0)
		mode = SpellCastPrimarySkill
		manaCost = 200
		damage = 80
		scalef = .9
		spellspeed = 280.0
		spellType = "projectile"
		interval = FireballSpellInterval
	case "mini-explo":
		sheet = Pictures["./images/smallExplosion.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		frames = getFrames(sheet, 48, 48, 17, 0)
		mode = Normal
		manaCost = 0
		damage = 0
		speed = 16
		scalef = .9
	case "icesnipe":
		casterType = Sniper
		sheet = Pictures["./images/icesnipe.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		frames = getFrames(sheet, 64, 64, 30, 0)
		mode = SpellCastSecondarySkill
		manaCost = 800
		damage = 210
		speed = 12
		spellspeed = 500
		spellType = "projectile"
		interval = IcesnipeSpellInterval
	case "blood-explo":
		sheet = Pictures["./images/blood.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		unorderedFrames := getFrames(sheet, 100, 100, 6, 4)
		for i := range unorderedFrames {
			frames = append(frames, unorderedFrames[BloodFrames[i]])
		}
		mode = Normal
		manaCost = 0
		damage = 0
		speed = 25
		scalef = 1.5
	case "lava-spot":
		casterType = DarkWizard
		sheet = Pictures["./images/lavaSpot.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		frames = getFrames(sheet, 128, 128, 10, 0)
		mode = SpellCastSecondarySkill
		manaCost = 1200
		damage = 100 //por segundo
		speed = 12
		spellspeed = 0
		scalef = 1.5
		spellType = "aoe"
		lifespawn = time.Second.Seconds() * 5
		interval = LavaSpellInterval
	case "heal-spot":
		casterType = Monk
		sheet = Pictures["./images/healingSpot.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		frames = getFrames(sheet, 128, 128, 10, 0)
		mode = SpellCastSecondarySkill
		manaCost = 1200
		damage = -90 //por segundo
		speed = 12
		spellspeed = 0
		scalef = 1.5
		spellType = "aoe"
		lifespawn = time.Second.Seconds() * 4
		interval = LavaSpellInterval
	case "healshot":
		casterType = Monk
		sheet = Pictures["./images/healingShot.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		unorderedFrames := getFrames(sheet, 64, 75, 8, 7)
		for i := range unorderedFrames {
			frames = append(frames, unorderedFrames[HealingShotFrames[i]])
		}
		mode = SpellCastSecondarySkill
		manaCost = 350
		damage = -60
		speed = 15
		spellspeed = 240
		scalef = 1
		spellType = "projectile"
		lifespawn = time.Second.Seconds()
		interval = IcesnipeSpellInterval
	case "mana-spot":
		casterType = Shaman
		sheet = Pictures["./images/manaSpot.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		frames = getFrames(sheet, 128, 128, 12, 0)
		mode = SpellCastSecondarySkill
		manaCost = 1200
		damage = -300 //por segundo
		speed = 12
		spellspeed = 0
		scalef = 1.5
		spellType = "aoe"
		lifespawn = time.Second.Seconds() * 6
		interval = LavaSpellInterval
	case "manashot":
		casterType = Shaman
		sheet = Pictures["./images/manaShot.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		unorderedFrames := getFrames(sheet, 64, 75, 8, 7)
		for i := range unorderedFrames {
			frames = append(frames, unorderedFrames[HealingShotFrames[i]])
		}
		mode = SpellCastSecondarySkill
		manaCost = 300
		damage = 400
		speed = 15
		spellspeed = 240
		scalef = 1
		spellType = "projectile"
		lifespawn = time.Second.Seconds()
		interval = IcesnipeSpellInterval
	case "smoke-spot":
		casterType = Sniper
		sheet = Pictures["./images/smokeSpot.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		frames = getFrames(sheet, 128, 128, 12, 0)
		mode = SpellCastSecondarySkill
		manaCost = 1200
		damage = 0
		speed = 12
		spellspeed = 0
		scalef = 1.5
		spellType = "aoe"
		lifespawn = time.Second.Seconds() * 3
		interval = LavaSpellInterval
	case "rockshot":
		casterType = Timewreker
		sheet = Pictures["./images/rockShot.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		unorderedFrames := getFrames(sheet, 64, 64, 8, 8)
		for i := range unorderedFrames {
			frames = append(frames, unorderedFrames[RockFrames[i]])
		}
		mode = SpellCastSecondarySkill
		manaCost = 700
		damage = 130
		speed = 40
		spellspeed = 230
		scalef = .5
		spellType = "projectile"
		lifespawn = time.Second.Seconds()
		interval = IcesnipeSpellInterval
	case "flash":
		casterType = Timewreker
		sheet = Pictures["./images/flashEffect.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		frames = getFrames(sheet, 64, 64, 10, 0)
		mode = SpellCastSecondarySkill
		manaCost = 900
		damage = 0
		speed = 25
		spellspeed = 0
		scalef = 1.5
		spellType = "movement"
		lifespawn = time.Second.Seconds() * 3
		interval = FlashSpellInterval

	}

	return &SpellData{
		WizardCaster:      casterType,
		EffectRadius:      efectRaduis,
		SpellLifespawn:    lifespawn,
		Interval:          interval,
		SpellType:         spellType,
		ProjSpeed:         spellspeed,
		ScaleF:            scalef,
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
	damageInterval time.Time
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

func (a *Spell) NextFrameFireball(spellFrames []pixel.Rect, speed, lifespawn float64) (pixel.Rect, bool) {
	dt := time.Since(a.last).Seconds()
	pdt := time.Since(a.projectileLife).Seconds()
	a.last = time.Now()
	a.frameNumber += 21 * dt
	i := int(a.frameNumber)
	if i <= len(spellFrames)-1 {
		if speed != 0 {
			vel := pixel.V(1, 1).Rotated(a.vel.Angle()).Rotated(-pixel.V(1, 1).Angle()).Scaled(dt * speed)
			a.pos = a.pos.Add(vel)
			(*a.matrix) = a.matrix.Moved(vel)
		}
		return spellFrames[i], false
	}

	a.frameNumber = .0
	if pdt > time.Second.Seconds()*lifespawn {
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
func (s Skins) DrawToBatch(p *Player) {
	p.Update()
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

func GameUpdate(s *socket.Socket, pd *PlayersData, p *Player, spells SpellKinds) {
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
				if spell.SpellType != "projectile" && spell.SpellType != "aoe" && spell.SpellType != "movement" {
					if s.ClientID == spell.TargetID {
						target = p
					} else {
						target = pd.CurrentAnimations[spell.TargetID]
					}
					newSpell.target = target
					newSpell.matrix = &target.headMatrix
				}

				for i := range spells.OnTarget {
					sd := spells.OnTarget[i]
					if spell.SpellName == sd.SpellName {
						newSpell.step = sd.Frames[0]
						newSpell.frame = pixel.NewSprite(*(sd.Pic), newSpell.step)
						sd.CurrentAnimations = append(sd.CurrentAnimations, newSpell)
						target.hp -= sd.Damage
						if target.hp <= 0 {
							target.hp = 0
							target.dead = true
						}
						break
					}
				}
				for i := range spells.Projectile {
					sd := spells.Projectile[i]
					if spell.SpellName == sd.SpellName {
						caster := pd.CurrentAnimations[spell.ID]
						vel := pixel.V(spell.X, spell.Y).Sub(caster.pos)
						centerMatrix := pixel.IM
						switch spell.SpellName {
						case "fireball":
							centerMatrix = caster.bodyMatrix.Rotated(caster.pos, vel.Angle()+(math.Pi/2)).Scaled(caster.pos, 2)
						case "icesnipe":
							centerMatrix = caster.bodyMatrix.Rotated(caster.pos, vel.Angle()).Scaled(caster.pos, .6)
						case "healshot", "manashot":
							centerMatrix = caster.bodyMatrix.Rotated(caster.pos, vel.Angle()+(math.Pi/2)).Scaled(caster.pos, .6)
						case "rockshot":
							centerMatrix = caster.bodyMatrix.Rotated(caster.pos, vel.Angle())
						}
						newSpell.caster = spell.ID
						newSpell.vel = vel
						newSpell.pos = caster.pos
						newSpell.matrix = &centerMatrix
						newSpell.step = sd.Frames[0]
						newSpell.frame = pixel.NewSprite(*(sd.Pic), newSpell.step)
						sd.CurrentAnimations = append(sd.CurrentAnimations, newSpell)
					}
				}
				for i := range spells.AOE {
					sd := spells.AOE[i]
					if spell.SpellName == sd.SpellName {
						newSpell.pos = pixel.V(spell.X, spell.Y)
						centerMatrix := pixel.IM.Moved(newSpell.pos)
						newSpell.caster = spell.ID
						newSpell.matrix = &centerMatrix
						newSpell.step = sd.Frames[0]
						newSpell.frame = pixel.NewSprite(*(sd.Pic), newSpell.step)
						sd.CurrentAnimations = append(sd.CurrentAnimations, newSpell)
					}
				}
				for i := range spells.Movement {
					sd := spells.Movement[i]
					if spell.SpellName == sd.SpellName {
						caster := pd.CurrentAnimations[spell.ID]
						newSpell.pos = caster.pos
						centerMatrix := pixel.IM.Moved(newSpell.pos)
						newSpell.caster = spell.ID
						newSpell.matrix = &centerMatrix
						newSpell.step = sd.Frames[0]
						newSpell.frame = pixel.NewSprite(*(sd.Pic), newSpell.step)
						sd.CurrentAnimations = append(sd.CurrentAnimations, newSpell)
					}
				}
			case models.Chat:
				chatMsg := models.ChatMsg{}
				json.Unmarshal(msg.Payload, &chatMsg)
				pd.CurrentAnimations[chatMsg.ID].chat.WriteSent(chatMsg.Message)
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

func (pd *PlayersData) Draw(win *pixelgl.Window) {
	pd.Skins.BatchClear()
	pd.AnimationsMutex.RLock()
	for _, p := range pd.CurrentAnimations {
		pd.AnimationsMutex.RUnlock()
		pd.Skins.DrawToBatch(p)
		if !p.invisible {
			p.name.Draw(win, p.nameMatrix.Moved(pixel.V(0, -8)))
			p.chat.Draw(win, p.pos)
			p.DrawHealthMana(win)
		}
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
	hval := Map(float64(p.hp), 0, float64(MaxHealth), 0, 32)
	info.Push(
		infoPos.Add(pixel.V(0, 0)),
		infoPos.Add(pixel.V(hval, 0)),
		infoPos.Add(pixel.V(0, -2)),
		infoPos.Add(pixel.V(hval, -2)),
	)
	info.Rectangle(0)

	info.Draw(win)
}

type Chat struct {
	msgTimeout      time.Time
	chatting        bool
	sent, writing   *text.Text
	ssent, swriting string
	scolor, wcolor  color.RGBA
	matrix          pixel.Matrix
}

func (c *Chat) WriteSent(message string) {
	c.ssent = message
	c.sent.WriteString(c.ssent)
	c.msgTimeout = time.Now()
}

func (c *Chat) Send(s *socket.Socket) {
	c.ssent = c.swriting
	c.sent.WriteString(c.ssent)
	c.msgTimeout = time.Now()
	chatMsg := &models.ChatMsg{
		ID:      s.ClientID,
		Message: c.ssent,
	}
	chatPayload, err := json.Marshal(chatMsg)
	if err != nil {
		return
	}
	s.O <- models.NewMesg(models.Chat, chatPayload)
	c.swriting = ""
	c.writing.Clear()
}

func (c *Chat) Write(win *pixelgl.Window) {
	c.writing.WriteString(win.Typed())
	if win.Typed() != "" {
		c.swriting = fmt.Sprint(c.swriting, win.Typed())
	}
	if win.JustPressed(pixelgl.KeyBackspace) || win.Repeated(pixelgl.KeyBackspace) {
		if c.swriting != "" {
			c.swriting = c.swriting[:len(c.swriting)-1]
			c.writing.Clear()
			c.writing.WriteString(c.swriting)
		}
	}
}

func (c *Chat) Draw(win *pixelgl.Window, pos pixel.Vec) {

	if c.chatting {
		c.writing.Clear()
		c.writing.WriteString(c.swriting)
		c.writing.Draw(win, pixel.IM.Moved(pos.Sub(c.writing.Bounds().Center().Floor()).Add(pixel.V(0, 46))))
		return
	}
	dt := time.Since(c.msgTimeout).Seconds()
	if dt < time.Second.Seconds()*5 {
		c.sent.Clear()
		c.sent.WriteString(c.ssent)
		c.sent.Draw(win, pixel.IM.Moved(pos.Sub(c.sent.Bounds().Center().Floor()).Add(pixel.V(0, 46))))
	} else {
		c.sent.Clear()
		c.ssent = ""
	}
}

func NewPlayer(name string, wizard *Wizard) Player {
	p := &Player{}
	p.sname = name
	basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	p.name = text.New(pixel.V(-28, 0), basicAtlas)

	p.chat = Chat{
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
		p.name.Color = colornames.Darkcyan
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
	p.lastCastPrimary = time.Now().Add(-time.Second * 2)
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
	p.mp = MaxMana
	p.hp = MaxHealth
	return *p
}

func (p *Player) OnMe(click pixel.Vec) bool {
	r := click.X < p.pos.X+14 && click.X > p.pos.X-14 && click.Y < p.pos.Y+30 && click.Y > p.pos.Y-20
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
		p.nameMatrix = pixel.IM.Moved(p.pos.Sub(p.name.Bounds().Center().Floor()).Add(pixel.V(0, -26)))
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
				p.mp += 117
				if p.mp > MaxMana {
					p.mp = MaxMana
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
		p.nameMatrix = pixel.IM.Moved(p.pos.Sub(p.name.Bounds().Center().Floor()).Add(pixel.V(0, -26)))
		p.head = pixel.NewSprite(*p.deadHeadPic, p.headFrame)
		p.body = pixel.NewSprite(*p.deadPic, p.bodyFrame)
		p.rooted = false
	}
	if time.Since(p.lastRootedStart).Seconds() > time.Second.Seconds()*1.5 {
		p.rooted = false
	}
}

func (p *Player) Draw(win *pixelgl.Window, s *socket.Socket) {
	p.Update()
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
	if !p.invisible {
		p.body.Draw(win, p.bodyMatrix)
		p.head.Draw(win, p.headMatrix)
		p.name.Draw(win, p.nameMatrix)
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

type TreeType int

const (
	ZombieTree TreeType = iota
	TallTree
	NoLeafsTallTree
)

type Tree struct {
	Frame pixel.Rect
	Batch *pixel.Batch
	Pic   *pixel.Picture
}

type Trees []*Tree

func (s Trees) Load(kind TreeType, imagPath string) {
	sheet := Pictures[imagPath]
	tree := &Tree{
		Pic:   &sheet,
		Batch: pixel.NewBatch(&pixel.TrianglesData{}, sheet),
		Frame: sheet.Bounds(),
	}
	s[kind] = tree
}

func (s Trees) Draw(win *pixelgl.Window) {
	for i := range s {
		s[i].Batch.Draw(win)
	}
}

type Forest struct {
	Pic, GrassPic, FencePicH, FencePicV                            pixel.Picture
	Frames                                                         []pixel.Rect
	FenceFrameH, FenceFrameV, GrassFrames                          pixel.Rect
	Batch, GrassBatch, FenceBatchHTOP, FenceBatchHBOT, FenceBatchV *pixel.Batch
	Trees                                                          Trees
}

func generateRandomPoss(c, bot, top, left, rigth int64) []pixel.Vec {
	poss := make([]pixel.Vec, c)
	rand.Seed(c + top + bot + left + rigth)
	for i := range poss {
		poss[i] = pixel.V(random(left, rigth), random(bot, top))
	}
	return poss
}
func random(min, max int64) float64 {
	return float64(rand.Int63n(max-min) + min)
}

func NewForest() *Forest {
	treeSheet := Pictures["./images/trees.png"]
	treeBatch := pixel.NewBatch(&pixel.TrianglesData{}, treeSheet)
	treeFrames := getFrames(treeSheet, 32, 32, 3, 3)

	trees := make(Trees, 3)
	trees.Load(ZombieTree, "./images/arbolmuerto.png")
	trees.Load(TallTree, "./images/talltree.png")
	trees.Load(NoLeafsTallTree, "./images/tallnoleafstree.png")

	grassSheet := Pictures["./images/newGrass.png"]
	grassBatch := pixel.NewBatch(&pixel.TrianglesData{}, grassSheet)
	grassFrames := grassSheet.Bounds()

	vfenceSheet := Pictures["./images/verticalfence.png"]
	vfenceBatch := pixel.NewBatch(&pixel.TrianglesData{}, vfenceSheet)
	vfenceFrame := vfenceSheet.Bounds()
	hfenceSheet := Pictures["./images/horizontalfence.png"]
	hfenceBatchBot := pixel.NewBatch(&pixel.TrianglesData{}, hfenceSheet)
	hfenceBatchTop := pixel.NewBatch(&pixel.TrianglesData{}, hfenceSheet)
	hfenceFrame := hfenceSheet.Bounds()

	for x := 0; x <= 13; x++ {
		for y := 0; y <= 13; y++ {
			pos := pixel.V(float64(x*320)-160, float64(y*320)-160)
			bread := pixel.NewSprite(grassSheet, grassFrames)
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
	poss := generateRandomPoss(200, 0, int64(pathTop)-100, 0, 1800)
	poss = append(poss, generateRandomPoss(200, 0, int64(pathTop)-100, 2200, 4000)...)
	for i := range poss {
		tree := pixel.NewSprite(treeSheet, treeFrames[rand.Intn(len(treeFrames))])
		tree.Draw(treeBatch, pixel.IM.Scaled(pixel.ZV, 3.5).Moved(poss[i]))
	}

	// fill with zombie trees
	zombieTrees := generateRandomPoss(5, int64(pathTop)+100, int64(arenaTop), 1100, 2900)
	ztree := trees[ZombieTree]
	for i := 0; i <= len(zombieTrees)-1; i++ {
		tree := pixel.NewSprite(*ztree.Pic, ztree.Frame)
		tree.Draw(ztree.Batch, pixel.IM.Moved(zombieTrees[i]))
	}

	// fill with tall trees
	// commented because they were ugly

	// tallTress := generateRandomPoss(10, int64(pathTop)+100, int64(arenaTop), 1100, 2900)
	// ttree := trees[TallTree]
	// for i := 0; i <= len(tallTress)-1; i++ {
	// 	tree := pixel.NewSprite(*ttree.Pic, ttree.Frame)
	// 	tree.Draw(ttree.Batch, pixel.IM.Scaled(pixel.ZV, 1.2).Moved(tallTress[i]))
	// }

	nltallTress := generateRandomPoss(8, int64(pathTop)+100, int64(arenaTop), 1100, 2900)
	nlttree := trees[NoLeafsTallTree]
	for i := 0; i <= len(nltallTress)-1; i++ {
		tree := pixel.NewSprite(*nlttree.Pic, nlttree.Frame)
		tree.Draw(nlttree.Batch, pixel.IM.Scaled(pixel.ZV, 1.2).Moved(nltallTress[i]))
	}

	return &Forest{
		Trees:          trees,
		Pic:            treeSheet,
		Frames:         treeFrames,
		Batch:          treeBatch,
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
