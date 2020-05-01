package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/juanefec/go-pixel-ao/models"
	"github.com/segmentio/ksuid"
)

func main() {

	port := 33333

	SocketServer(port)

}

func SocketServer(port int) {

	listen, err := net.Listen("tcp4", ":"+strconv.Itoa(port))

	if err != nil {
		log.Fatalf("Socket listen port %d failed,%s", port, err)
		os.Exit(1)
	}

	defer listen.Close()

	log.Printf("Begin listen port: %d", port)

	game := NewGame()
	defer game.End()
	go game.Run()
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatalln(err)
			continue
		}
		log.Printf("Connected to: %v", conn.RemoteAddr().String())
		go ServeGame(&conn, game)
	}

}

// ServeGame handles websocket requests from the peer.
func ServeGame(conn *net.Conn, game *Game) {
	id := ksuid.New()
	client := &Client{ID: id, game: game, conn: conn, send: make(chan []byte, 512), endupdate: make(chan struct{})}
	client.game.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
	go game.RankingUpdater(client)

}

var (
	Newline = []byte{'\n'}
)

type Ranking []*models.RankingPosMsg

func (r Ranking) ToMsg() []byte {
	bs, _ := json.Marshal(r)
	msg := models.Mesg{
		Type:    models.UpdateRanking,
		Payload: bs,
	}
	m, _ := json.Marshal(msg)
	return m
}

func (r *Ranking) Update(payload json.RawMessage) {
	d := models.DeathMsg{}
	err := json.Unmarshal(payload, &d)
	if err != nil {
		return
	}
	killerExist := false
	killedExist := false
	for i := range *r {
		if (*r)[i].ID == d.Killer {
			(*r)[i].K++
			killerExist = true
		}
		if (*r)[i].ID == d.Killed {
			(*r)[i].D++
			killedExist = true
		}
	}
	if !killedExist {
		rr := &models.RankingPosMsg{
			ID:   d.Killed,
			Name: d.KilledName,
			K:    0,
			D:    1,
		}
		*r = append(*r, rr)
	}
	if !killerExist {
		rr := &models.RankingPosMsg{
			ID:   d.Killer,
			Name: d.KillerName,
			K:    1,
			D:    0,
		}
		*r = append(*r, rr)
	}
}

type Client struct {
	ID        ksuid.KSUID
	game      *Game
	conn      *net.Conn
	send      chan []byte
	endupdate chan struct{}
}

func (c *Client) readPump() {
	defer func() {
		log.Printf("Disconnected: %v", (*c.conn).RemoteAddr().String())
		log.Printf("Exited Client.readPump: %v", c.ID)
		c.game.unregister <- c
		(*c.conn).Close()
	}()
	var (
		data bytes.Buffer
		r    = bufio.NewReader(*c.conn)
	)

	for {
		dataRead, isPrefix, err := r.ReadLine()
		if err != nil {
			log.Printf("Error: %v", err.Error())
			break

		}

		data.Write(dataRead)
		if isPrefix {
			continue
		}
		msg := models.UnmarshallMesg(data.Bytes())
		switch msg.Type {
		case models.Chat, models.Spell:
			c.game.eventBroadcast <- BroadcastEvent{
				Client:  c,
				Event:   msg.Type,
				Payload: msg.Payload}
			break
		case models.UpdateServer:
			c.game.clientsUpdate <- BroadcastEvent{
				Client:  c,
				Event:   msg.Type,
				Payload: msg.Payload}
			break
		case models.Death:
			c.game.Ranking.Update(msg.Payload)
			break
		}
		data = bytes.Buffer{}
	}

}

func (c *Client) writePump() {
	defer func() {
		log.Printf("Exited Client.writePump: %v", c.ID)
		(*c.conn).Close()
	}()
	var w = bufio.NewWriter(*c.conn)

	for msg := range c.send {
		msg = makeMessage(msg)
		w.Write(msg)
		if err := w.Flush(); err != nil {
			log.Printf("Error: %v", err.Error())
			return
		}
		//log.Printf("Send: %v|END", string(message))
	}

}

func makeMessage(d []byte) []byte {
	d = append(d, Newline...)
	return d
}

type BroadcastEvent struct {
	Client  *Client
	Event   models.Event
	Payload json.RawMessage
}

type Game struct {
	Online         int
	Ranking        Ranking
	Players        map[ksuid.KSUID]*models.PlayerMsg
	Pmutex         *sync.RWMutex
	clientsUpdate  chan BroadcastEvent
	clients        map[*Client]bool
	register       chan *Client
	unregister     chan *Client
	eventBroadcast chan BroadcastEvent
}

func NewGame() *Game {
	return &Game{
		Online:         0,
		Ranking:        make(Ranking, 0),
		Players:        make(map[ksuid.KSUID]*models.PlayerMsg),
		clientsUpdate:  make(chan BroadcastEvent),
		Pmutex:         &sync.RWMutex{},
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		clients:        make(map[*Client]bool),
		eventBroadcast: make(chan BroadcastEvent),
	}
}

func (g *Game) End() {
	close(g.clientsUpdate)
	close(g.register)
	close(g.unregister)
}

func (g *Game) Run() {
	go func() {
		for {
			select {
			case event := <-g.eventBroadcast:
				for c := range g.clients {
					if c.ID != event.Client.ID {
						c.send <- models.NewMesg(event.Event, event.Payload)
					}
				}
				if event.Event == models.Disconect {
					delete(g.clients, event.Client)
				}
			}
		}
	}()

	logger := time.Tick(time.Second * 5)
	for {
		select {
		case msg := <-g.clientsUpdate:
			g.UpdateServer(msg)
			msg.Client.send <- g.UpdateClient(msg.Client)

		case client := <-g.register:
			g.clients[client] = true
			client.send <- client.ID.Bytes()

		case client := <-g.unregister:
			if _, ok := g.clients[client]; ok {
				p := models.DisconectMsg{ID: client.ID}
				payload, _ := json.Marshal(p)
				g.eventBroadcast <- BroadcastEvent{
					Client:  client,
					Event:   models.Disconect,
					Payload: payload,
				}
				client.endupdate <- struct{}{}

				delete(g.Players, client.ID)
			}

		case <-logger:
			log.Println("player list len: ", len(g.Players))
		}

	}
}

func (g *Game) RankingUpdater(c *Client) {
	updater := time.Tick(time.Second)
ULOOP:
	for {
		select {
		case <-c.endupdate:
			break ULOOP
		case <-updater:
			c.send <- g.Ranking.ToMsg()

		}

	}
	close(c.send)
	log.Printf("Exited Game.ClientUpdater: %v", c.ID)
}

func (g *Game) UpdateServer(message BroadcastEvent) {
	var msg models.PlayerMsg
	err := json.Unmarshal(message.Payload, &msg)
	if err == nil {

		g.Pmutex.Lock()
		if _, ok := g.Players[msg.ID]; !ok {
			g.Online++
		}
		g.Players[msg.ID] = &msg
		g.Pmutex.Unlock()

	} else {
		log.Printf("err: %v", err.Error())

	}
}

func (g *Game) UpdateClient(c *Client) []byte {

	g.Pmutex.RLock()
	playerSlice := getPlayerList(g.Players)
	g.Pmutex.RUnlock()

	playersMsg, _ := json.Marshal(playerSlice)
	msg := models.NewMesg(models.UpdateClient, playersMsg)
	return msg
}

func getPlayerList(m map[ksuid.KSUID]*models.PlayerMsg) []*models.PlayerMsg {
	var res []*models.PlayerMsg
	for _, v := range m {
		res = append(res, v)
	}
	return res
}
