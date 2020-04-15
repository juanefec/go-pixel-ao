package socket

import (
	"bufio"
	"bytes"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/juanefec/go-pixel-ao/proto/events"
	"github.com/segmentio/ksuid"
	"google.golang.org/protobuf/proto"
)

type EventStream struct {
	I, O chan []byte
}

func (es *EventStream) Close() {
	close(es.I)
	close(es.O)
}

// Socket stores the connection and the IO to comunicate wit the server
type Socket struct {
	ClientID ksuid.KSUID
	conn     *net.Conn
	Events   []*EventStream
}

// Close the connection and IO
func (s Socket) Close() {
	for i := range s.Events {
		s.Events[i].Close()
	}
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
	pi, po, ai, ao := make(chan []byte), make(chan []byte), make(chan []byte), make(chan []byte)
	s := &Socket{
		conn: &conn,
		Events: []*EventStream{
			{I: pi, O: po},
			{I: ai, O: ao},
		},
	}
	reader := bufio.NewReader(conn)
	for s.ClientID == ksuid.Nil {
		data, _, _ := reader.ReadLine()
		if len(data) == 20 {
			//fmt.Println(string(data))
			_ = s.ClientID.UnmarshalBinary(data)
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
	var buffer bytes.Buffer
	r := bufio.NewReader(*s.conn)
	for {

		data, isPrefix, err := r.ReadLine()
		if err == nil {
			buffer.Write(data)
			if isPrefix {
				continue
			}
			pdata := events.Player{}
			err := proto.Unmarshal(buffer.Bytes(), pdata)
			if err == nil {
				select {
				case s.Events[0].I <- buffer.Bytes():
					log.Printf("Receive: %s", buffer.Bytes())
					buffer = bytes.Buffer{}
				default:
				}
				continue
			}

			adata := events.ApocaEvent{}
			err = proto.Unmarshal(buffer.Bytes(), adata)
			if err == nil {
				select {
				case s.Events[1].I <- buffer.Bytes():
					log.Printf("Receive: %s", buffer.Bytes())
					buffer = bytes.Buffer{}
				default:
				}
				continue
			}

		}

	}
}

func (s *Socket) sender() {
	var w = bufio.NewWriter(*s.conn)

	for {
		select {

		case message := <-s.Events[0].O:
			w.Write(message)
			if err := w.Flush(); err != nil {
				return
			}
			//log.Printf("Send: %s", message)
		case message := <-s.Events[1].O:
			w.Write(message)
			if err := w.Flush(); err != nil {
				return
			}
			log.Printf("Send: %s", message)
		default:
		}
	}
}
