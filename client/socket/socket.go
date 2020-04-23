package socket

import (
	"bufio"
	"bytes"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/segmentio/ksuid"
)

// Socket stores the connection and the IO to comunicate wit the server
type Socket struct {
	Online   bool
	ClientID ksuid.KSUID
	conn     *net.Conn
	I, O     chan []byte
}

// Close the connection and IO
func (s *Socket) Close() {
	s.Online = false
	(*s.conn).Close()
}

// NewSocket generation
func NewSocket(ip string, port int) *Socket {

	addr := strings.Join([]string{ip, strconv.Itoa(port)}, ":")
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
	s := &Socket{
		Online: true,
		conn:   &conn,
		I:      make(chan []byte),
		O:      make(chan []byte),
	}
	reader := bufio.NewReader(conn)
	for s.ClientID == ksuid.Nil {
		data, _, _ := reader.ReadLine()
		if len(data) == 20 {
			_ = s.ClientID.UnmarshalBinary(data)
			log.Printf("Client ID: %v", s.ClientID.String())
		}
	}
	go s.reciver()
	go s.sender()
	//log.Printf("listening io")

	return s
}

//message order [updatePlayer|id;name;playerX;playerY;dir;moving]
//message order [newApoca|id;name;x;y]

func (s *Socket) reciver() {
	defer s.Close()

	var buffer bytes.Buffer
	r := bufio.NewReader(*s.conn)

	for {

		data, isPrefix, err := r.ReadLine()
		if err == nil {
			buffer.Write(data)
			if isPrefix {
				continue
			}
			s.I <- buffer.Bytes()
			buffer = bytes.Buffer{}
		} else {
			return
		}

	}
}

func (s *Socket) sender() {
	defer s.Close()

	var w = bufio.NewWriter(*s.conn)

	for {
		select {

		case message := <-s.O:
			message = makeMessage(message)
			w.Write(message)
			if err := w.Flush(); err != nil {
				s.Close()
				return
			}
		default:
		}
	}
}

var Newline = []byte{'\n'}

func makeMessage(d []byte) []byte {
	d = append(d, Newline...)
	return d
}
