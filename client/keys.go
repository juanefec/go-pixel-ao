package main

import (
	"encoding/json"
	"time"

	"github.com/faiface/pixel/pixelgl"
	"github.com/juanefec/go-pixel-ao/client/socket"
	"github.com/juanefec/go-pixel-ao/models"
)

type MapBound float64

const (
	Top    = 4000
	Bottom = 0
	Left   = 0
	Right  = 4000
)

type KeyConfig struct {
	Apoca    int `json:"apoca_key"`
	Desca    int `json:"desca_key"`
	Explo    int `json:"explo_key"`
	FireB    int `json:"fireball_key"`
	IceSnipe int `json:"icesnipe_key"`
	Rojas    int `json:"rojas_key"`
	Azules   int `json:"azules_key"`
}

func keyInputs(win *pixelgl.Window, player *Player, cursor *Cursor, socket *socket.Socket) {
	const (
		KeyUp    = pixelgl.KeyW
		KeyDown  = pixelgl.KeyS
		KeyLeft  = pixelgl.KeyA
		KeyRight = pixelgl.KeyD
	)

	directionKeyStatuses := map[pixelgl.Button]bool{
		KeyUp:    false,
		KeyDown:  false,
		KeyLeft:  false,
		KeyRight: false,
	}

	tpTime := time.Now()

	updateMovement := func() {
		moveMsg := &models.MoveMsg{
			ID:        socket.ClientID,
			Direction: "",
		}

		if directionKeyStatuses[KeyUp] {
			moveMsg.Direction += "U"
		} else if directionKeyStatuses[KeyDown] {
			moveMsg.Direction += "D"
		}
		if directionKeyStatuses[KeyLeft] {
			moveMsg.Direction += "L"
		} else if directionKeyStatuses[KeyRight] {
			moveMsg.Direction += "R"
		}

		movePayload, err := json.Marshal(moveMsg)
		if err != nil {
			return
		}
		socket.O <- models.NewMesg(models.Move, movePayload)
	}

	checkMovementKey := func(keyToCheck pixelgl.Button, inverseKey pixelgl.Button) {
		if (win.JustPressed(keyToCheck) || win.Pressed(keyToCheck)) && !directionKeyStatuses[inverseKey] && !directionKeyStatuses[keyToCheck] {
			directionKeyStatuses[keyToCheck] = true
			updateMovement()
		}
		if win.JustReleased(keyToCheck) {
			directionKeyStatuses[keyToCheck] = false
			updateMovement()
		}
	}

	for !win.Closed() {
		if player.chat.chatting {
			for key := range directionKeyStatuses {
				directionKeyStatuses[key] = false
			}
			updateMovement()
		}

		if !player.chat.chatting && !player.rooted {
			checkMovementKey(KeyLeft, KeyRight)
			checkMovementKey(KeyRight, KeyLeft)
			checkMovementKey(KeyUp, KeyDown)
			checkMovementKey(KeyDown, KeyUp)

			if win.JustPressed(pixelgl.Button(Key.Explo)) {
				cursor.SetSpellExploMode()
			}

			if win.JustPressed(pixelgl.Button(Key.Apoca)) {
				cursor.SetSpellApocaMode()
			}

			if win.JustPressed(pixelgl.Button(Key.Desca)) {
				cursor.SetSpellDescaMode()
			}

		} else {
			player.moving = false
		}
		if win.Pressed(pixelgl.Button(Key.Azules)) {
			player.drinkingManaPotions = true
		} else {
			player.drinkingManaPotions = false
		}

		if win.Pressed(pixelgl.Button(Key.Rojas)) {
			player.drinkingHealthPotions = true
		} else {
			player.drinkingHealthPotions = false
		}

		if player.sname == "   creagod   " && win.Pressed(pixelgl.KeyLeftShift) {
			if win.JustPressed(pixelgl.MouseButtonLeft) {
				if dt := time.Since(tpTime).Seconds(); dt > time.Second.Seconds()/6 {
					tpTime = time.Now()
					tppos := player.cam.Unproject(win.MousePosition())
					player.pos.X, player.pos.Y = tppos.X, tppos.Y
				}
			}
		}
	}
}
