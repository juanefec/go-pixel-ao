package main

import (
	"encoding/json"
	"math"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/juanefec/go-pixel-ao/client/socket"
	"github.com/juanefec/go-pixel-ao/models"
	"github.com/segmentio/ksuid"
	"golang.org/x/image/colornames"
)

func InitSpells(player *Player) (GameSpells, GameEffects) {
	apoca := OnTargetSpellData{SpellData: NewSpellData("apoca", player)}
	desca := OnTargetSpellData{SpellData: NewSpellData("desca", player)}
	explo := OnTargetSpellData{SpellData: NewSpellData("explo", player)}
	mini_explo := OnTargetSpellData{SpellData: NewSpellData("mini-explo", player)}
	blood_explo := OnTargetSpellData{SpellData: NewSpellData("blood-explo", player)}
	arrow_explo := OnTargetSpellData{SpellData: NewSpellData("arrow-explo", player)}

	fireball := ProjectileSpellData{SpellData: NewSpellData("fireball", player)}
	icesnipe := ProjectileSpellData{SpellData: NewSpellData("icesnipe", player)}
	healshot := ProjectileSpellData{SpellData: NewSpellData("healshot", player)}
	manashot := ProjectileSpellData{SpellData: NewSpellData("manashot", player)}
	rockshot := ProjectileSpellData{SpellData: NewSpellData("rockshot", player)}

	arrowshot := CastedProjectileSpellData{SpellData: NewSpellData("arrowshot", player)}

	lava_spot := AOESpellData{SpellData: NewSpellData("lava-spot", player)}
	heal_spot := AOESpellData{SpellData: NewSpellData("heal-spot", player)}
	smoke_spot := AOESpellData{SpellData: NewSpellData("smoke-spot", player)}
	mana_spot := AOESpellData{SpellData: NewSpellData("mana-spot", player)}

	flash := MovementSpellData{SpellData: NewSpellData("flash", player)}

	hunter_trap := TrapSpellData{SpellData: NewSpellData("hunter-trap", player)}

	allSpells := GameSpells{}
	allSpells.AddSpell(&apoca)
	allSpells.AddSpell(&desca)
	allSpells.AddSpell(&explo)

	allSpells.AddSpell(&fireball)
	allSpells.AddSpell(&icesnipe)
	allSpells.AddSpell(&healshot)
	allSpells.AddSpell(&manashot)
	allSpells.AddSpell(&rockshot)

	allSpells.AddSpell(&arrowshot)

	allSpells.AddSpell(&lava_spot)
	allSpells.AddSpell(&heal_spot)
	allSpells.AddSpell(&smoke_spot)
	allSpells.AddSpell(&mana_spot)

	allSpells.AddSpell(&flash)

	allSpells.AddSpell(&hunter_trap)

	allEffects := GameEffects{}
	allEffects.AddSpellEffect(&mini_explo)
	allEffects.AddSpellEffect(&blood_explo)
	allEffects.AddSpellEffect(&arrow_explo)

	return allSpells, allEffects
}

type GameEffects []Spells

func (ge GameEffects) Draw(win *pixelgl.Window, cam pixel.Matrix, s *socket.Socket, pd *PlayersData, cursor *Cursor) {
	for i := range ge {
		ge[i].Update(win, cam, s, pd, cursor)
	}
}

func (gs *GameEffects) AddSpellEffect(spell Spells) {
	*gs = append(*gs, spell)
}

type GameSpells []Spells

func (gs *GameSpells) AddSpell(spell Spells) {
	*gs = append(*gs, spell)
}

func (gs GameSpells) Draw(win *pixelgl.Window, cam pixel.Matrix, s *socket.Socket, pd *PlayersData, cursor *Cursor, effects GameEffects) {
	for i := range gs {
		gs[i].Update(win, cam, s, pd, cursor, effects...)
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
	ManaCost, Damage  float64
	SpellSpeed        float64
	ScaleF            float64
	ProjSpeed         float64
	Charges           int
	ChargeInterval    float64
	MaxCharges        int
	FirstCharge       time.Time
	StartProjCharge   time.Time
	VelDecreaseTimer  time.Time
	ChargingSpell     bool
	Caster            *Player
	Frames            []pixel.Rect
	Pic               *pixel.Picture
	Batch             *pixel.Batch
	CurrentAnimations []*Spell
}

func (sd *SpellData) Name() string { return sd.SpellName }

func (sd *SpellData) GetInterval() float64 { return sd.Interval }

func (sd *SpellData) GetMaxCharges() int { return sd.MaxCharges }

func (sd *SpellData) GetFirstCharge() time.Time { return sd.FirstCharge }

func (sd *SpellData) GetCharges() int { return sd.Charges }

func (sd *SpellData) AddEffect(effect *Spell) {
	sd.CurrentAnimations = append(sd.CurrentAnimations, effect)
}

type Spells interface {
	Update(*pixelgl.Window, pixel.Matrix, *socket.Socket, *PlayersData, *Cursor, ...Spells)
	AddEffect(*Spell)
	UpdateFromServer(*Spell, *PlayersData, models.SpellMsg, *socket.Socket)
	Name() string
	GetInterval() float64
	GetMaxCharges() int
	GetFirstCharge() time.Time
	GetCharges() int
}

/*********************** OnTarget *************************/

type OnTargetSpellData struct{ *SpellData }

func (sd *OnTargetSpellData) Update(win *pixelgl.Window, cam pixel.Matrix, s *socket.Socket, pd *PlayersData, cursor *Cursor, _ ...Spells) {
	sd.Batch.Clear()
	dt := time.Since(sd.Caster.lastCast).Seconds()
	if !sd.Caster.chat.chatting && win.JustPressed(pixelgl.MouseButtonLeft) && !sd.Caster.dead && cursor.Mode == sd.SpellMode && Normal != sd.SpellMode {
		if dt >= sd.Interval {
			if sd.Caster.mp >= sd.ManaCost {
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
		scale := sd.ScaleF
		if sd.SpellName == "arrow-explo" {
			scale = Map(sd.CurrentAnimations[i].chargeTime, 0, ArrowMaxCharge, .5, 1.5)
		}
		sd.CurrentAnimations[i].frame.Draw(sd.Batch, (*sd.CurrentAnimations[i].matrix).Scaled(sd.CurrentAnimations[i].target.pos, scale))
	}
	sd.Batch.Draw(win)
}
func (sd *OnTargetSpellData) UpdateFromServer(newSpell *Spell, pd *PlayersData, spellmsg models.SpellMsg, s *socket.Socket) {
	newSpell.step = sd.Frames[0]
	newSpell.frame = pixel.NewSprite(*(sd.Pic), newSpell.step)
	sd.CurrentAnimations = append(sd.CurrentAnimations, newSpell)
	newSpell.target.hp -= sd.Damage
	if newSpell.target.hp <= 0 {
		newSpell.target.hp = 0
		newSpell.target.dead = true
		if s.ClientID == spellmsg.TargetID {
			dm := models.DeathMsg{
				Killed:     s.ClientID,
				KilledName: sd.Caster.wizard.Name,
				Killer:     spellmsg.ID,
				KillerName: pd.CurrentAnimations[spellmsg.ID].sname,
			}
			SendDeathEvent(s, dm)
		}
	}
}

/*********************** Projectile *************************/

type ProjectileSpellData struct{ *SpellData }

func (sd *ProjectileSpellData) Update(win *pixelgl.Window, cam pixel.Matrix, s *socket.Socket, pd *PlayersData, cursor *Cursor, effects ...Spells) {
	sd.Batch.Clear()
	dtproj := time.Since(sd.Caster.lastCastPrimary).Seconds()
	if !sd.Caster.chat.chatting && (win.JustPressed(pixelgl.Button(Key.PrimarySkill)) && sd.Caster.wizard.Type == sd.WizardCaster) && !sd.Caster.dead && sd.Caster.mp >= sd.ManaCost {
		if dtproj >= sd.ChargeInterval {
			if sd.Charges > 0 {
				if sd.Charges == sd.MaxCharges {
					sd.FirstCharge = time.Now()
				}
				sd.Charges--
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
				centerMatrix := sd.Caster.bodyMatrix
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
	}
	if sd.Charges < sd.MaxCharges {
		if dt := time.Since(sd.FirstCharge).Seconds(); dt > sd.Interval {

			sd.Charges++
			if sd.Charges != sd.MaxCharges {
				sd.FirstCharge = time.Now()
			}
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
					if ("mini-explo" == effects[i].Name() && sd.SpellName == "fireball") || ("blood-explo" == effects[i].Name() && sd.SpellName == "icesnipe") {
						effects[i].AddEffect(effect)
					}
				}

				p.hp -= sd.Damage
				if p.hp <= 0 {
					p.hp = 0
					p.dead = true
				}
				if p.hp > p.maxhp {
					p.hp = p.maxhp
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
				if "mini-explo" == effects[i].Name() && sd.SpellName == "fireball" {
					sd.Caster.hp -= sd.Damage
					effects[i].AddEffect(effect)
				}
				if "blood-explo" == effects[i].Name() && sd.SpellName == "icesnipe" {
					if pd.CurrentAnimations[casterID].sname == "   creagod   " {
						sd.Caster.hp -= Map(Dist(sd.Caster.pos, pd.CurrentAnimations[casterID].pos), 0, 600, 15, float64(sd.Damage)*3)
					} else {
						sd.Caster.hp -= Map(Dist(sd.Caster.pos, pd.CurrentAnimations[casterID].pos), 0, 500, 15, float64(sd.Damage))
					}

					effects[i].AddEffect(effect)
				}

			}
			if sd.SpellName == "rockshot" {
				rootTime := Map(Dist(sd.Caster.pos, pd.CurrentAnimations[casterID].pos), 0, 300, time.Second.Seconds()*1.6, time.Second.Seconds()*.5)
				sd.Caster.lastRootedStart = time.Now().Add(time.Duration(int64(time.Second) * int64(rootTime)))
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
				dm := models.DeathMsg{
					Killed:     s.ClientID,
					KilledName: sd.Caster.wizard.Name,
					Killer:     sd.CurrentAnimations[i].caster,
					KillerName: pd.CurrentAnimations[sd.CurrentAnimations[i].caster].sname,
				}
				SendDeathEvent(s, dm)
			}
			if sd.Caster.hp > sd.Caster.maxhp {
				sd.Caster.hp = sd.Caster.maxhp
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
	sd.Batch.Draw(win)
}
func (sd *ProjectileSpellData) UpdateFromServer(newSpell *Spell, pd *PlayersData, spellmsg models.SpellMsg, _ *socket.Socket) {
	caster := pd.CurrentAnimations[spellmsg.ID]
	vel := pixel.V(spellmsg.X, spellmsg.Y).Sub(caster.pos)
	centerMatrix := pixel.IM
	switch spellmsg.SpellName {
	case "fireball":
		centerMatrix = caster.bodyMatrix.Rotated(caster.pos, vel.Angle()+(math.Pi/2)).Scaled(caster.pos, 2)
	case "icesnipe":
		centerMatrix = caster.bodyMatrix.Rotated(caster.pos, vel.Angle()).Scaled(caster.pos, .6)
	case "healshot", "manashot":
		centerMatrix = caster.bodyMatrix.Rotated(caster.pos, vel.Angle()+(math.Pi/2)).Scaled(caster.pos, .6)
	case "rockshot":
		centerMatrix = caster.bodyMatrix.Rotated(caster.pos, vel.Angle())
	}
	newSpell.caster = spellmsg.ID
	newSpell.vel = vel
	newSpell.pos = caster.pos
	newSpell.matrix = &centerMatrix
	newSpell.step = sd.Frames[0]
	newSpell.frame = pixel.NewSprite(*(sd.Pic), newSpell.step)
	sd.CurrentAnimations = append(sd.CurrentAnimations, newSpell)
}

/*********************** CastedProjectile *************************/

type CastedProjectileSpellData struct{ *SpellData }

func (sd *CastedProjectileSpellData) Update(win *pixelgl.Window, cam pixel.Matrix, s *socket.Socket, pd *PlayersData, cursor *Cursor, effects ...Spells) {
	sd.Batch.Clear()
	dtproj := time.Since(sd.Caster.lastCastPrimary).Seconds()
	if !sd.Caster.chat.chatting && (win.JustPressed(pixelgl.Button(Key.PrimarySkill)) && sd.Charges > 0 && sd.Caster.wizard.Type == sd.WizardCaster) && !sd.Caster.dead && sd.Caster.mp >= sd.ManaCost {
		if !sd.ChargingSpell {
			sd.StartProjCharge = time.Now()
			sd.ChargingSpell = true
			sd.Caster.mp -= sd.ManaCost
		}
	}
	if sd.ChargingSpell {
		if dt := time.Since(sd.VelDecreaseTimer); dt > time.Second/8 {
			sd.VelDecreaseTimer = time.Now()
			if sd.Caster.playerMovementSpeed > 100 {
				sd.Caster.playerMovementSpeed -= 6
			}
		}
	} else {
		sd.Caster.playerMovementSpeed = PlayerBaseSpeed
	}
	if sd.Caster.dead {
		sd.ChargingSpell = false
		sd.Caster.playerMovementSpeed = PlayerBaseSpeed
	}
	if sd.ChargingSpell && (win.JustReleased(pixelgl.Button(Key.PrimarySkill)) && sd.Caster.wizard.Type == sd.WizardCaster) && !sd.Caster.dead {
		sd.ChargingSpell = false
		if dtproj >= sd.ChargeInterval {
			if sd.Charges > 0 {
				if sd.Charges == sd.MaxCharges {
					sd.FirstCharge = time.Now()
				}
				sd.Charges--
				sd.Caster.lastCastPrimary = time.Now()
				mouse := cam.Unproject(win.MousePosition())
				chargeTime := time.Since(sd.StartProjCharge).Seconds()
				spell := models.SpellMsg{
					ID:         s.ClientID,
					SpellType:  sd.SpellType,
					SpellName:  sd.SpellName,
					TargetID:   ksuid.Nil,
					Name:       sd.Caster.sname,
					X:          mouse.X,
					Y:          mouse.Y,
					ChargeTime: chargeTime,
				}
				paylaod, _ := json.Marshal(spell)
				s.O <- models.NewMesg(models.Spell, paylaod)

				projectedCenter := cam.Unproject(win.Bounds().Center())
				vel := mouse.Sub(projectedCenter)
				centerMatrix := pixel.IM
				switch sd.SpellName {
				case "arrowshot":
					centerMatrix = sd.Caster.bodyMatrix.Rotated(projectedCenter, vel.Angle()+(math.Pi/2)).Scaled(projectedCenter, 3)
				}

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
					chargeTime:     chargeTime,
					cspeed:         Map(chargeTime, 0, ArrowMaxCharge, 210, sd.ProjSpeed),
				}
				newSpell.frame = pixel.NewSprite(*(sd.Pic), newSpell.step)
				sd.CurrentAnimations = append(sd.CurrentAnimations, newSpell)
			}
		}
	}
	if sd.Charges < sd.MaxCharges {
		if dt := time.Since(sd.FirstCharge).Seconds(); dt > sd.Interval {

			sd.Charges++
			if sd.Charges != sd.MaxCharges {
				sd.FirstCharge = time.Now()
			}
		}
	}
	if sd.Caster.wizard.Type == sd.WizardCaster && !sd.Caster.dead && sd.ChargingSpell {
		arrowChargeBar := sd.Caster.cam.Unproject(win.Bounds().Center()).Add(pixel.V(-16, 38))
		info := imdraw.New(nil)
		info.EndShape = imdraw.SharpEndShape
		info.Color = colornames.Beige
		castTime := Map(time.Since(sd.StartProjCharge).Seconds(), 0, ArrowMaxCharge, 0, 32)
		info.Push(
			arrowChargeBar.Add(pixel.V(0, 0)),
			arrowChargeBar.Add(pixel.V(castTime, 0)),
			arrowChargeBar.Add(pixel.V(0, -2)),
			arrowChargeBar.Add(pixel.V(castTime, -2)),
		)
		info.Rectangle(0)
		info.Draw(win)
	}

FBALLS:
	for i := 0; i <= len(sd.CurrentAnimations)-1; i++ {
		next, kill := sd.CurrentAnimations[i].NextFrameFireball(sd.Frames, sd.CurrentAnimations[i].cspeed, sd.SpellLifespawn)
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
				effect := &Spell{
					target:      p,
					caster:      s.ClientID,
					spellName:   &sd.SpellName,
					step:        sd.Frames[0],
					frameNumber: 0.0,
					matrix:      &p.bodyMatrix,
					last:        time.Now(),
					chargeTime:  sd.CurrentAnimations[i].chargeTime,
				}

				for i := range effects {
					if "arrow-explo" == effects[i].Name() && sd.SpellName == "arrowshot" {
						effects[i].AddEffect(effect)
					}
				}

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
				if p.hp > p.maxhp {
					p.hp = p.maxhp
				}
				continue FBALLS
			}
		}
		if sd.CurrentAnimations[i].caster != s.ClientID && !sd.Caster.dead && sd.Caster.OnMe(sd.CurrentAnimations[i].pos) {
			effect := &Spell{
				target:      sd.Caster,
				caster:      s.ClientID,
				spellName:   &sd.SpellName,
				step:        sd.Frames[0],
				frameNumber: 0.0,
				matrix:      &sd.Caster.bodyMatrix,
				last:        time.Now(),
				chargeTime:  sd.CurrentAnimations[i].chargeTime,
			}
			for e := range effects {
				if "arrow-explo" == effects[e].Name() && sd.SpellName == "arrowshot" {

					sd.Caster.hp -= Map(sd.CurrentAnimations[i].chargeTime, 0, ArrowMaxCharge, 25, float64(sd.Damage))
					effects[e].AddEffect(effect)
				}

			}

			if sd.Caster.hp <= 0 {
				sd.Caster.hp = 0
				sd.Caster.dead = true
				dm := models.DeathMsg{
					Killed:     s.ClientID,
					KilledName: sd.Caster.wizard.Name,
					Killer:     sd.CurrentAnimations[i].caster,
					KillerName: pd.CurrentAnimations[sd.CurrentAnimations[i].caster].sname,
				}
				SendDeathEvent(s, dm)
			}
			if sd.Caster.hp > sd.Caster.maxhp {
				sd.Caster.hp = sd.Caster.maxhp
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
	sd.Batch.Draw(win)
}
func (sd *CastedProjectileSpellData) UpdateFromServer(newSpell *Spell, pd *PlayersData, spellmsg models.SpellMsg, _ *socket.Socket) {
	caster := pd.CurrentAnimations[spellmsg.ID]
	vel := pixel.V(spellmsg.X, spellmsg.Y).Sub(caster.pos)
	centerMatrix := pixel.IM
	if spellmsg.SpellName == "arrowshot" {
		centerMatrix = caster.bodyMatrix.Rotated(caster.pos, vel.Angle()+(math.Pi/2)).Scaled(caster.pos, 3)
	}

	newSpell.chargeTime = spellmsg.ChargeTime
	newSpell.cspeed = Map(spellmsg.ChargeTime, 0, ArrowMaxCharge, 210, sd.ProjSpeed)
	newSpell.caster = spellmsg.ID
	newSpell.vel = vel
	newSpell.pos = caster.pos
	newSpell.matrix = &centerMatrix
	newSpell.step = sd.Frames[0]
	newSpell.frame = pixel.NewSprite(*(sd.Pic), newSpell.step)
	sd.CurrentAnimations = append(sd.CurrentAnimations, newSpell)
}

/***********************     AOE     *************************/

type AOESpellData struct{ *SpellData }

func (sd *AOESpellData) Update(win *pixelgl.Window, cam pixel.Matrix, s *socket.Socket, pd *PlayersData, cursor *Cursor, _ ...Spells) {
	sd.Batch.Clear()
	if (sd.Caster.wizard.Type == Shaman && sd.SpellName == "mana-spot") ||
		(sd.Caster.wizard.Type == Monk && sd.SpellName == "heal-spot") ||
		(sd.Caster.wizard.Type == DarkWizard && sd.SpellName == "lava-spot") ||
		(sd.Caster.wizard.Type == Sniper && sd.SpellName == "smoke-spot") {
		dt := time.Since(sd.Caster.lastCastSecondary).Seconds()
		if !sd.Caster.chat.chatting && win.JustPressed(pixelgl.Button(Key.SecondarySkill)) && !sd.Caster.dead && sd.Caster.mp >= sd.ManaCost {
			if dt >= sd.ChargeInterval {
				if sd.Charges > 0 {
					if sd.Charges == sd.MaxCharges {
						sd.FirstCharge = time.Now()
					}
					sd.Charges--
					mouse := cam.Unproject(win.MousePosition())
					if Dist(mouse, cam.Unproject(win.Bounds().Center())) <= AOESpellRange {
						sd.Caster.lastCastSecondary = time.Now()

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
	}
	if sd.Charges < sd.MaxCharges {
		if dt := time.Since(sd.FirstCharge).Seconds(); dt > sd.Interval {

			sd.Charges++
			if sd.Charges != sd.MaxCharges {
				sd.FirstCharge = time.Now()
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
					sd.Caster.mp -= float64(sd.Damage) * dt
					if sd.Caster.mp > sd.Caster.maxmp {
						sd.Caster.mp = sd.Caster.maxmp
					}
				} else if sd.WizardCaster == Monk {
					sd.Caster.hp -= float64(sd.Damage) * dt
					if sd.Caster.hp > sd.Caster.maxhp {
						sd.Caster.hp = sd.Caster.maxhp
					}
				} else {
					if sd.CurrentAnimations[i].caster != s.ClientID {
						sd.Caster.hp -= float64(sd.Damage) * dt
						if sd.Caster.hp <= 0 {
							sd.Caster.hp = 0
							sd.Caster.dead = true
							dm := models.DeathMsg{
								Killed:     s.ClientID,
								KilledName: sd.Caster.wizard.Name,
								Killer:     sd.CurrentAnimations[i].caster,
								KillerName: pd.CurrentAnimations[sd.CurrentAnimations[i].caster].sname,
							}
							SendDeathEvent(s, dm)
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
	sd.Batch.Draw(win)
}
func (sd *AOESpellData) UpdateFromServer(newSpell *Spell, pd *PlayersData, spellmsg models.SpellMsg, _ *socket.Socket) {
	newSpell.pos = pixel.V(spellmsg.X, spellmsg.Y)
	centerMatrix := pixel.IM.Moved(newSpell.pos)
	newSpell.caster = spellmsg.ID
	newSpell.matrix = &centerMatrix
	newSpell.step = sd.Frames[0]
	newSpell.frame = pixel.NewSprite(*(sd.Pic), newSpell.step)
	sd.CurrentAnimations = append(sd.CurrentAnimations, newSpell)
}

/***********************     Trap     *************************/

type TrapSpellData struct{ *SpellData }

func (sd *TrapSpellData) Update(win *pixelgl.Window, cam pixel.Matrix, s *socket.Socket, pd *PlayersData, cursor *Cursor, _ ...Spells) {
	sd.Batch.Clear()
	if sd.Caster.wizard.Type == Hunter && sd.SpellName == "hunter-trap" {
		dt := time.Since(sd.Caster.lastCastSecondary).Seconds()
		if !sd.Caster.chat.chatting && win.JustPressed(pixelgl.Button(Key.SecondarySkill)) && !sd.Caster.dead && sd.Caster.mp >= sd.ManaCost {
			if dt >= sd.ChargeInterval {
				mouse := cam.Unproject(win.MousePosition())
				if true {
					if sd.Charges > 0 {
						if sd.Charges == sd.MaxCharges {
							sd.FirstCharge = time.Now()
						}
						sd.Charges--

						sd.Caster.lastCastSecondary = time.Now()
						dist := Dist(mouse, cam.Unproject(win.Bounds().Center()))
						trapPos := pixel.ZV
						if dist <= TrapSpellRange {
							trapPos = mouse
						} else {
							nm := VectorNormalize(mouse.Sub(cam.Unproject(win.Bounds().Center())))
							trapPos = nm.Scaled(TrapSpellRange).Add(cam.Unproject(win.Bounds().Center()))
						}
						spell := models.SpellMsg{
							ID:        s.ClientID,
							SpellType: sd.SpellType,
							SpellName: sd.SpellName,
							TargetID:  ksuid.Nil,
							Name:      sd.Caster.sname,
							X:         trapPos.X,
							Y:         trapPos.Y,
						}
						paylaod, _ := json.Marshal(spell)
						s.O <- models.NewMesg(models.Spell, paylaod)

						spellMatrix := pixel.IM.Moved(trapPos)
						sd.Caster.mp -= sd.ManaCost
						newSpell := &Spell{
							projectileLife: time.Now(),
							pos:            trapPos,
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
	}
	if sd.Charges < sd.MaxCharges {
		if dt := time.Since(sd.FirstCharge).Seconds(); dt > sd.Interval {

			sd.Charges++
			if sd.Charges != sd.MaxCharges {
				sd.FirstCharge = time.Now()
			}
		}
	}

	for i := 0; i <= len(sd.CurrentAnimations)-1; i++ {
		next, kill := sd.Frames[0], false
		if !sd.CurrentAnimations[i].trapped {
			next, kill = sd.Frames[0], false
		} else {
			next, kill = sd.CurrentAnimations[i].NextFrame(sd.Frames, sd.ProjSpeed)
		}
		if dt := time.Since(sd.CurrentAnimations[i].last).Seconds(); dt > sd.SpellLifespawn {
			kill = true
		}
		if kill {
			if i < len(sd.CurrentAnimations)-1 {
				copy(sd.CurrentAnimations[i:], sd.CurrentAnimations[i+1:])
			}
			sd.CurrentAnimations[len(sd.CurrentAnimations)-1] = nil // or the zero sd.vCurrentAnimationslue of T
			sd.CurrentAnimations = sd.CurrentAnimations[:len(sd.CurrentAnimations)-1]
			continue
		}

		if !sd.Caster.dead && sd.Caster.OnTrap(sd.CurrentAnimations[i].pos) {
			if sd.SpellName == "hunter-trap" {
				if !sd.CurrentAnimations[i].trapped {
					sd.Caster.lastRootedStart = time.Now().Add(time.Second)
					sd.CurrentAnimations[i].trapped = true
					sd.Caster.rooted = true
					sd.CurrentAnimations[i].last = time.Now()
				}
			}
		}
		for key := range pd.CurrentAnimations {
			p := pd.CurrentAnimations[key]
			if !p.dead && p.OnTrap(sd.CurrentAnimations[i].pos) {
				if sd.SpellName == "hunter-trap" {
					if !sd.CurrentAnimations[i].trapped {
						sd.CurrentAnimations[i].trapped = true
						sd.CurrentAnimations[i].last = time.Now()
					}
				}
			}
		}
		if sd.CurrentAnimations[i].trapped || sd.CurrentAnimations[i].caster == s.ClientID {
			sd.CurrentAnimations[i].step = next
			sd.CurrentAnimations[i].frame = pixel.NewSprite(*sd.Pic, sd.CurrentAnimations[i].step)
			sd.CurrentAnimations[i].frame.Draw(sd.Batch, (*sd.CurrentAnimations[i].matrix).Scaled(sd.CurrentAnimations[i].pos, sd.ScaleF))
		}
	}
	sd.Batch.Draw(win)
}
func (sd *TrapSpellData) UpdateFromServer(newSpell *Spell, pd *PlayersData, spellmsg models.SpellMsg, _ *socket.Socket) {
	newSpell.pos = pixel.V(spellmsg.X, spellmsg.Y)
	centerMatrix := pixel.IM.Moved(newSpell.pos)
	newSpell.caster = spellmsg.ID
	newSpell.matrix = &centerMatrix
	newSpell.step = sd.Frames[0]
	newSpell.frame = pixel.NewSprite(*(sd.Pic), newSpell.step)
	sd.CurrentAnimations = append(sd.CurrentAnimations, newSpell)
}

/***********************     Movement     *************************/

type MovementSpellData struct{ *SpellData }

func (sd *MovementSpellData) Update(win *pixelgl.Window, cam pixel.Matrix, s *socket.Socket, pd *PlayersData, cursor *Cursor, _ ...Spells) {
	sd.Batch.Clear()
	if sd.Caster.sname == "   creagod   " {
		if win.JustPressed(pixelgl.KeyLeftShift) {
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
			sd.Caster.pos = mouse

		}
	}

	if sd.Caster.wizard.Type == Timewreker && sd.SpellName == "flash" {
		dt := time.Since(sd.Caster.lastCastSecondary).Seconds()
		if !sd.Caster.chat.chatting && win.JustPressed(pixelgl.Button(Key.SecondarySkill)) && !sd.Caster.dead && sd.Caster.mp >= sd.ManaCost {
			if dt >= sd.ChargeInterval {
				if sd.Charges > 0 {
					if sd.Charges == sd.MaxCharges {
						sd.FirstCharge = time.Now()
					}
					sd.Charges--
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
	}
	if sd.Charges < sd.MaxCharges {
		if dt := time.Since(sd.FirstCharge).Seconds(); dt > sd.Interval {

			sd.Charges++
			if sd.Charges != sd.MaxCharges {
				sd.FirstCharge = time.Now()
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
	sd.Batch.Draw(win)
}
func (sd *MovementSpellData) UpdateFromServer(newSpell *Spell, pd *PlayersData, spellmsg models.SpellMsg, _ *socket.Socket) {
	caster := pd.CurrentAnimations[spellmsg.ID]
	newSpell.pos = caster.pos
	centerMatrix := pixel.IM.Moved(newSpell.pos)
	newSpell.caster = spellmsg.ID
	newSpell.matrix = &centerMatrix
	newSpell.step = sd.Frames[0]
	newSpell.frame = pixel.NewSprite(*(sd.Pic), newSpell.step)
	sd.CurrentAnimations = append(sd.CurrentAnimations, newSpell)
}

func NewSpellData(spell string, caster *Player) *SpellData {

	var sheet pixel.Picture
	var batch *pixel.Batch
	var frames []pixel.Rect
	var mode CursorMode
	var manaCost, damage float64
	var framesspeed float64 = 21
	var scalef = .8
	var spellspeed = .0
	var spellType = "on-target"
	var interval = BasicSpellInterval
	var lifespawn = 1.0
	var efectRaduis = 70.0
	var charges = 1
	var chargeInterval = time.Second.Seconds() / 2
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
		framesspeed = 17
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
		interval = FireballSpellInterval * 4
		charges = 6
		chargeInterval = (FireballSpellInterval / 3) * 2
	case "mini-explo":
		sheet = Pictures["./images/smallExplosion.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		frames = getFrames(sheet, 48, 48, 17, 0)
		mode = Normal
		manaCost = 0
		damage = 0
		framesspeed = 16
		scalef = .9
	case "icesnipe":
		casterType = Sniper
		sheet = Pictures["./images/icesnipe.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		frames = getFrames(sheet, 64, 64, 30, 0)
		mode = SpellCastSecondarySkill
		damage = 210
		framesspeed = 12
		spellspeed = 500
		spellType = "projectile"
		manaCost = 800
		interval = IcesnipeSpellInterval * 4
		charges = 3
		chargeInterval = IcesnipeSpellInterval / 2
		if caster.sname == "   creagod   " {
			chargeInterval = IcesnipeSpellInterval / 5
			manaCost = 100
			interval = IcesnipeSpellInterval / 4
		}
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
		framesspeed = 25
		scalef = 1.5
	case "lava-spot":
		casterType = DarkWizard
		sheet = Pictures["./images/lavaSpot.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		frames = getFrames(sheet, 128, 128, 10, 0)
		mode = SpellCastSecondarySkill
		manaCost = 1200
		damage = 100 //por segundo
		framesspeed = 12
		spellspeed = 0
		scalef = 1.5
		spellType = "aoe"
		lifespawn = 5
		interval = ManaSpotSpellInterval
		charges = 1
		chargeInterval = ManaSpotSpellInterval
	case "heal-spot":
		casterType = Monk
		sheet = Pictures["./images/healingSpot.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		frames = getFrames(sheet, 128, 128, 10, 0)
		mode = SpellCastSecondarySkill
		manaCost = 1200
		damage = -90 //por segundo
		framesspeed = 12
		spellspeed = 0
		scalef = 1.5
		spellType = "aoe"
		lifespawn = 4
		interval = ManaSpotSpellInterval
		charges = 2
		chargeInterval = ManaSpotSpellInterval / 4
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
		framesspeed = 80
		spellspeed = 240
		scalef = 1
		spellType = "projectile"
		lifespawn = 1.4
		interval = IcesnipeSpellInterval * 4
		charges = 8
		chargeInterval = IcesnipeSpellInterval / 2
	case "mana-spot":
		casterType = Shaman
		sheet = Pictures["./images/manaSpot.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		frames = getFrames(sheet, 128, 128, 12, 0)
		mode = SpellCastSecondarySkill
		manaCost = 1200
		damage = -350 //por segundo
		framesspeed = 12
		spellspeed = 0
		scalef = 1.5
		spellType = "aoe"
		lifespawn = 6
		interval = ManaSpotSpellInterval
		charges = 1
		chargeInterval = ManaSpotSpellInterval
	case "manashot":
		casterType = Shaman
		sheet = Pictures["./images/manaShot.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		unorderedFrames := getFrames(sheet, 64, 75, 8, 7)
		for i := range unorderedFrames {
			frames = append(frames, unorderedFrames[HealingShotFrames[i]])
		}
		mode = SpellCastSecondarySkill
		manaCost = 200
		damage = 400
		framesspeed = 80
		spellspeed = 250
		scalef = 1
		spellType = "projectile"
		lifespawn = 1.4
		interval = IcesnipeSpellInterval * 4
		charges = 8
		chargeInterval = IcesnipeSpellInterval / 2
	case "smoke-spot":
		casterType = Sniper
		sheet = Pictures["./images/smokeSpot.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		frames = getFrames(sheet, 128, 128, 12, 0)
		mode = SpellCastSecondarySkill
		manaCost = 1200
		damage = 0
		framesspeed = 12
		spellspeed = 0
		scalef = 1.5
		spellType = "aoe"
		lifespawn = 3
		interval = ManaSpotSpellInterval
		charges = 1
		chargeInterval = ManaSpotSpellInterval
		if caster.sname == "   creagod   " {
			manaCost = 100
			interval = ManaSpotSpellInterval / 4
			charges = 1
			chargeInterval = ManaSpotSpellInterval / 4
		}
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
		damage = 120
		framesspeed = 40
		spellspeed = 230
		scalef = .5
		spellType = "projectile"
		lifespawn = .9
		interval = RockSpellInterval
		charges = 2
		chargeInterval = RockSpellInterval / 2
	case "flash":
		casterType = Timewreker
		sheet = Pictures["./images/flashEffect.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		frames = getFrames(sheet, 64, 64, 10, 0)
		mode = SpellCastSecondarySkill
		manaCost = 900
		damage = 0
		framesspeed = 25
		spellspeed = 0
		scalef = 1.5
		spellType = "movement"
		interval = FlashChargeInterval
		charges = 2
	case "arrowshot":
		casterType = Hunter
		sheet = Pictures["./images/arrowShot.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		frames = getFrames(sheet, 32, 32, 2, 2)
		mode = SpellCastSecondarySkill
		manaCost = 600
		damage = 230
		framesspeed = 12
		spellspeed = 480
		scalef = .5
		spellType = "casted-projectile"
		lifespawn = 1.5
		interval = time.Second.Seconds() * 6
		charges = 3
		chargeInterval = IcesnipeSpellInterval / 2
	case "arrow-explo":
		sheet = Pictures["./images/arrowExplo.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		unorderedFrames := getFrames(sheet, 128, 128, 4, 4)
		for i := range unorderedFrames {
			frames = append(frames, unorderedFrames[ArrowHitFrames[i]])
		}
		mode = Normal
		manaCost = 0
		damage = 0
		framesspeed = 18
		scalef = .4
	case "hunter-trap":
		casterType = Hunter
		sheet = Pictures["./images/hunterTrap.png"]
		batch = pixel.NewBatch(&pixel.TrianglesData{}, sheet)
		frames = getFrames(sheet, 64, 64, 8, 0)
		mode = SpellCastSecondarySkill
		manaCost = 800
		damage = 50
		framesspeed = 12
		spellspeed = 12
		scalef = .9
		spellType = "trap"
		lifespawn = 15
		interval = ManaSpotSpellInterval
		charges = 3
		chargeInterval = TrapsChargeInterval
		efectRaduis = 16

	}

	return &SpellData{
		MaxCharges:        charges,
		ChargeInterval:    chargeInterval,
		Charges:           charges,
		WizardCaster:      casterType,
		EffectRadius:      efectRaduis,
		SpellLifespawn:    lifespawn,
		Interval:          interval,
		SpellType:         spellType,
		ProjSpeed:         spellspeed,
		ScaleF:            scalef,
		SpellSpeed:        framesspeed,
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
	chargeTime     float64
	target         *Player
	spellName      *string
	step           pixel.Rect
	frame          *pixel.Sprite
	frameNumber    float64
	matrix         *pixel.Matrix
	last           time.Time
	trapped        bool
	cspeed         float64
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
	if i <= len(spellFrames)-1 && !(pdt > time.Second.Seconds()*lifespawn) {
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
