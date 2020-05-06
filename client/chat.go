package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/juanefec/go-pixel-ao/client/socket"
	"github.com/juanefec/go-pixel-ao/models"
)

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
