package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

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
	// if r.CollidinMe(p.pos) {
	// 	p.colliding = true
	// 	p.collitionDir = p.dir
	// } else {
	// 	p.colliding = false
	// }
	if win.JustPressed(pixelgl.MouseButtonRight) {
		mouse := cam.Unproject(win.MousePosition())
		if r.OnMe(mouse) && p.dead {
			p.dead = false
			p.hp = p.maxhp
			p.mp = p.maxmp
		}
	}
	r.HeadSprite.Draw(win, pixel.IM.Moved(r.PosHead))
	r.BodySprite.Draw(win, pixel.IM.Moved(r.PosBody))
}

func (r *Resu) OnMe(click pixel.Vec) bool {
	b := click.X < r.PosBody.X+14 && click.X > r.PosBody.X-14 && click.Y < r.PosBody.Y+30 && click.Y > r.PosBody.Y-20
	return b
}

func (r *Resu) CollidinMe(player pixel.Vec) bool {
	b := player.X-14 < r.PosBody.X+14 && player.X+14 > r.PosBody.X-14 && player.Y-20 < r.PosBody.Y+20 && player.Y+20 > r.PosBody.Y-20

	return b
}
