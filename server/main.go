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

	hub := newHub()
	game := NewGame()
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatalln(err)
			continue
		}
		log.Printf("Connected to: %v", conn.RemoteAddr().String())
		go serveWs(&conn, hub, game)
	}

}

func isTransportOver(data string) (over bool) {
	over = strings.HasSuffix(data, "\r\n\r\n")
	return
}

// serveWs handles websocket requests from the peer.
func serveWs(conn *net.Conn, hub *Hub, game *Game) {
	id := ksuid.New()
	client := &Client{ID: id, hub: hub, conn: conn, send: make(chan []byte, 512)}

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go game.Run(client)
	go client.writePump()
	go client.readPump()

	client.hub.register <- client
}

type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

var (
	Newline = []byte{'\n'}
)

type Client struct {
	ID ksuid.KSUID

	hub *Hub

	// The websocket connection.
	conn *net.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

func (c *Client) readPump() {
	defer func() {
		log.Printf("Disconnected: %v", (*c.conn).RemoteAddr().String())
		c.hub.unregister <- c
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

		c.hub.broadcast <- data.Bytes()
		data = bytes.Buffer{}

	}
}

func (c *Client) writePump() {
	defer func() {
		(*c.conn).Close()
	}()
	var w = bufio.NewWriter(*c.conn)
	for {

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
}

func makeMessage(d []byte) []byte {
	d = append(d, Newline...)
	return d
}

type Game struct {
	Online  int
	Apocas  map[ksuid.KSUID]*models.ApocaMsg
	Players map[ksuid.KSUID]*models.PlayerMsg
	Pmutex  *sync.RWMutex
	Amutex  *sync.RWMutex
}

func NewGame() *Game {
	return &Game{
		Online:  0,
		Apocas:  make(map[ksuid.KSUID]*models.ApocaMsg),
		Players: make(map[ksuid.KSUID]*models.PlayerMsg),
		Pmutex:  &sync.RWMutex{},
		Amutex:  &sync.RWMutex{},
	}
}

func (g *Game) Run(c *Client) {
	updater := time.Tick(time.Second / 30)
	for {
		select {
		case msg := <-c.hub.broadcast:
			g.UpdateServer(msg)
		case client := <-c.hub.register:
			c.hub.clients[client] = true
			client.send <- client.ID.Bytes()
		case <-updater:
			c.send <- g.UpdateClient(c)
		case client := <-c.hub.unregister:
			if _, ok := c.hub.clients[client]; ok {
				delete(g.Players, client.ID)
				delete(c.hub.clients, client)
				close(client.send)
				return
			}
		}

	}

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
		g.Apocas[msg.Player.ID] = &msg.Apoca
		g.Amutex.Unlock()
	} else {
		log.Printf("err: %v", err.Error())

	}
}

func (g *Game) UpdateClient(c *Client) []byte {
	g.Amutex.RLock()
	apocaSlice := getApocaList(g.Apocas)
	g.Amutex.RUnlock()

	g.Pmutex.RLock()
	playerSlice := getPlayerList(g.Players)
	g.Pmutex.RUnlock()

	message := models.UpdateMessage{
		Players: playerSlice,
		Apocas:  apocaSlice,
	}
	msg, err := json.Marshal(message)
	if err != nil {
		return []byte{}
	}
	return msg
}

func getApocaList(m map[ksuid.KSUID]*models.ApocaMsg) []*models.ApocaMsg {
	var res []*models.ApocaMsg
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
