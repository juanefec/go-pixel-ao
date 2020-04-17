package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/juanefec/go-pixel-ao/models"
	"github.com/segmentio/ksuid"
)

func main() {

	port := 3333

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

func isTransportOver(data string) (over bool) {
	over = strings.HasSuffix(data, "\r\n\r\n")
	return
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
	go game.ClientUpdater(client)

}

var (
	Newline = []byte{'\n'}
)

type Client struct {
	ID ksuid.KSUID

	game *Game

	// The websocket connection.
	conn *net.Conn

	// Buffered channel of outbound messages.
	send chan []byte

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

		c.game.broadcast <- data.Bytes()
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

type Game struct {
	Online     int
	Spells     map[ksuid.KSUID]*models.SpellMsg
	Players    map[ksuid.KSUID]*models.PlayerMsg
	Pmutex     *sync.RWMutex
	Amutex     *sync.RWMutex
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func NewGame() *Game {
	return &Game{
		Online:     0,
		Spells:     make(map[ksuid.KSUID]*models.SpellMsg),
		Players:    make(map[ksuid.KSUID]*models.PlayerMsg),
		Pmutex:     &sync.RWMutex{},
		Amutex:     &sync.RWMutex{},
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (g *Game) End() {
	close(g.broadcast)
	close(g.register)
	close(g.unregister)
}

func (g *Game) Run() {
	logger := time.Tick(time.Second * 5)
	for {
		select {
		case msg := <-g.broadcast:
			g.UpdateServer(msg)
		case client := <-g.register:
			g.clients[client] = true
			client.send <- client.ID.Bytes()

		case client := <-g.unregister:
			if _, ok := g.clients[client]; ok {
				client.endupdate <- struct{}{}
				delete(g.Players, client.ID)
				delete(g.clients, client)

			}
		case <-logger:
			log.Println("player list len: ", len(g.Players))
		}

	}
}
func (g *Game) ClientUpdater(c *Client) {
	updater := time.Tick(time.Second / 22)
ULOOP:
	for {
		select {
		case <-c.endupdate:
			break ULOOP
		case <-updater:
			c.send <- g.UpdateClient(c)

		}

	}
	close(c.send)
	log.Printf("Exited Game.ClientUpdater: %v", c.ID)
}

func (g *Game) UpdateServer(message []byte) {
	var msg models.PlayerMessage
	err := json.Unmarshal(message, &msg)
	if err == nil {

		g.Pmutex.Lock()
		if _, ok := g.Players[msg.Player.ID]; !ok {
			g.Online++
		}
		g.Players[msg.Player.ID] = &msg.Player
		g.Pmutex.Unlock()
		g.Amutex.Lock()
		g.Spells[msg.Player.ID] = &msg.Spell
		g.Amutex.Unlock()
	} else {
		log.Printf("err: %v", err.Error())

	}
}

func (g *Game) UpdateClient(c *Client) []byte {
	g.Amutex.RLock()
	SpellSlice := getSpellList(g.Spells)
	g.Amutex.RUnlock()

	g.Pmutex.RLock()
	playerSlice := getPlayerList(g.Players)
	g.Pmutex.RUnlock()

	message := models.UpdateMessage{
		Players: playerSlice,
		Spells:  SpellSlice,
	}
	msg, err := json.Marshal(message)
	if err != nil {
		return []byte{}
	}
	return msg
}

func getSpellList(m map[ksuid.KSUID]*models.SpellMsg) []*models.SpellMsg {
	var res []*models.SpellMsg
	for _, v := range m {
		res = append(res, v)
	}
	return res
}
func getPlayerList(m map[ksuid.KSUID]*models.PlayerMsg) []*models.PlayerMsg {
	var res []*models.PlayerMsg
	for _, v := range m {
		res = append(res, v)
	}
	return res
}
