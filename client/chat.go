package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/juanefec/go-pixel-ao/client/socket"
	"github.com/juanefec/go-pixel-ao/models"
	"github.com/segmentio/ksuid"
)

const (
	ChatMasgSpace = 14.0
)

type Chatlog struct {
	msgs       []*ChatlogMsg
	lastUpdate time.Time
}

type ChatlogMsg struct {
	ID              ksuid.KSUID
	sender, message string
	txt             *text.Text
	tcreate         time.Time
}

func NewChatlog() *Chatlog {
	return &Chatlog{
		msgs:       make([]*ChatlogMsg, 0),
		lastUpdate: time.Now(),
	}
}

func (cl *Chatlog) Load(id ksuid.KSUID, sender, message string, tcreate time.Time) {
	//dt := time.Since(cl.lastUpdate)
	cm := &ChatlogMsg{
		ID:      id,
		sender:  sender,
		message: message,
		txt:     text.New(pixel.ZV, basicAtlas),
		tcreate: tcreate,
	}
	fmt.Fprintf(cm.txt, "[%v]: %v\n", strings.TrimSpace(sender), message)
	if len(cl.msgs) >= 8 {
		cl.msgs = cl.msgs[1:]
	}
	cl.msgs = append(cl.msgs, cm)
}

func (cl *Chatlog) Draw(win *pixelgl.Window, cam pixel.Matrix) {
	cl.lastUpdate = time.Now()
	for i := 0; i <= len(cl.msgs)-1; i++ {
		if mdt := time.Since(cl.msgs[i].tcreate); mdt < time.Second*30 {
			cl.msgs[i].txt.Draw(win, pixel.IM.Moved(cam.Unproject(pixel.V(win.Bounds().W()/6, (win.Bounds().H()-(win.Bounds().H()/40)-(float64(i)*ChatMasgSpace))))))
		} else {
			if i < len(cl.msgs)-1 {
				copy(cl.msgs[i:], cl.msgs[i+1:])
			}
			cl.msgs[len(cl.msgs)-1] = nil
			cl.msgs = cl.msgs[:len(cl.msgs)-1]
		}
	}
}

type Chat struct {
	p               *Player
	chatlog         *Chatlog
	msgTimeout      time.Time
	chatting        bool
	sent, writing   *text.Text
	ssent, swriting string
	scolor, wcolor  color.RGBA
	matrix          pixel.Matrix
}

func (c *Chat) WriteSent(id ksuid.KSUID, sender, message string) {
	c.ssent = message
	c.sent.WriteString(c.ssent)
	c.msgTimeout = time.Now()
	c.chatlog.Load(id, sender, c.ssent, c.msgTimeout)
}

func (c *Chat) Send(s *socket.Socket) {
	if c.swriting != "" {
		c.ssent = c.swriting
		c.sent.WriteString(c.ssent)
		c.msgTimeout = time.Now()
		chatMsg := &models.ChatMsg{
			ID:      s.ClientID,
			Name:    c.p.sname,
			Message: c.ssent,
		}
		chatPayload, err := json.Marshal(chatMsg)
		if err != nil {
			return
		}
		s.O <- models.NewMesg(models.Chat, chatPayload)
		c.swriting = ""
		c.writing.Clear()
		c.chatlog.Load(s.ClientID, c.p.sname, c.ssent, c.msgTimeout)
	}
}

func (c *Chat) Write(win *pixelgl.Window) {
	if win.Typed() != "" && len(c.swriting) <= 80 {
		c.writing.WriteString(win.Typed())
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
