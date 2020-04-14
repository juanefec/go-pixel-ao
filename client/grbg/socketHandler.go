package main

import (
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	message       = "Ping"
	StopCharacter = "\r\n\r\n"
)

var UpdatePlayers chan []byte
var UpdateClient chan []byte
var players = []*OPlayer{}
var ClientID = ""

type OPlayer struct {
	update chan []byte
	p      Player
}

//message order [id;name;playerX;playerY;dir;moving]

func SocketClient(ip string, port int) {

	addr := strings.Join([]string{ip, strconv.Itoa(port)}, ":")
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	ClientID := ""
	defer conn.Close()
	for ClientID == "" {
		buff := make([]byte, 1024)
		n, _ := conn.Read(buff)
		data := buff[:n]
		if len(data) == 20 {
			ClientID = string(data)
		}
	}

	go reciver(conn)
	go sender(conn)

}

func reciver(conn net.Conn) {
	for {
		buff := make([]byte, 1024)
		n, _ := conn.Read(buff)
		data := buff[:n]
		if n > 0 {
			UpdatePlayers <- data
			log.Printf("Receive: %s", data)
		}
	}
}

func sender(conn net.Conn) {
	for {
		select {
		case message := <-UpdateClient:
			conn.Write([]byte(message))
			conn.Write([]byte(StopCharacter))
			log.Printf("Send: %s", message)
		}
	}
}

func parseMessage() {

}

func main() {

	SocketClient("127.0.0.1", 3333)

}
