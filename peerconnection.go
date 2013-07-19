package gotorrent

import (
	"bytes"
	log "code.google.com/p/tcgl/applog"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/moretti/gotorrent/messages"
	"io"
	"net"
	"time"
)

type PeerConnection struct {
	addr net.TCPAddr
	conn net.Conn
	data []byte

	outErrors   chan<- PeerError
	inMessages  chan interface{}
	outMessages chan<- PeerMessage

	handshake bool
}

type PeerError struct {
	Addr net.TCPAddr
	Err  error
}

type PeerMessage struct {
	Addr    net.TCPAddr
	Message messages.Message
}

func NewPeerConnection(
	addr net.TCPAddr,
	outErrors chan<- PeerError,
	outMessages chan<- PeerMessage,
) *PeerConnection {

	pc := new(PeerConnection)
	pc.addr = addr
	pc.outErrors = outErrors
	pc.outMessages = outMessages
	pc.inMessages = make(chan interface{})

	return pc
}

func (pc *PeerConnection) Connect() {
	addr := pc.addr.String()
	log.Debugf("Connecting to %s...", addr)
	var err error
	pc.conn, err = net.DialTimeout("tcp", addr, time.Second*5)
	if err != nil {
		log.Debugf("Unable to connect to %s", addr)
		pc.outError(err)
	} else {
		log.Debugf("Connected to %s", addr)
		go pc.reader()
		go pc.writer()
	}
}

func (pc *PeerConnection) SendMessage(message interface{}) {
	go func() {
		pc.inMessages <- message
	}()
}

func (pc *PeerConnection) outError(err error) {
	pc.outErrors <- PeerError{Addr: pc.addr, Err: err}
}

func (pc *PeerConnection) outMessage(message messages.Message) {
	pc.outMessages <- PeerMessage{Addr: pc.addr, Message: message}
}

func (pc *PeerConnection) reader() {
	var err error
	defer func() {
		log.Debugf("Peer %v - Finished reading, err: %v", pc.addr, err)
		pc.conn.Close()
	}()

	if err := pc.readHandshake(); err != nil {
		return
	}

	bytesCount := 0
	buf := make([]byte, 2048)

	for {
		pc.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, err := pc.conn.Read(buf)

		if err != nil && err != io.EOF {
			if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
				return
			}
		}

		readed := buf[:n]
		//log.Debugf("Peer %v - Readed %v bytes", pc.addr, n)

		bytesCount += n
		pc.data = append(pc.data, readed...)

		if len(pc.data) > 4 {
			message, n, err := parseData(pc.data)
			if err != nil {
				return
			}
			if n > 0 {
				pc.data = pc.data[n:]
				//log.Debugf("Peer: %v - Sending message %v", pc.addr, message.Header)
				pc.outMessage(message)
			}
		}
	}
}

func parseData(data []byte) (message messages.Message, n int, err error) {
	n = 0
	var length uint32

	err = binary.Read(bytes.NewBuffer(data[:4]), binary.BigEndian, &length)
	if err != nil {
		log.Errorf("Cannot read the message length: %v", err)
		return
	}

	if length > 30*1024 {
		err = fmt.Errorf("Whoops, something went wrong, message length is too long: %v", length)
		return
	}

	if int(length) > len(data) {
		//log.Debugf("Peer: %v - I need more data, message length: %v, data length %v", pc.addr, length, len(data))
		return
	}

	message = messages.Message{}
	message.Header.Length = length

	data = data[4:]
	n += 4

	if message.Header.Length > 0 {
		message.Header.Id = data[0]
		message.Payload = data[1:message.Header.Length]
		n += int(message.Header.Length)
	}
	return
}

func (pc *PeerConnection) writer() {
	for message := range pc.inMessages {
		err := binary.Write(pc.conn, binary.BigEndian, message)
		if err != nil {
			pc.outError(err)
		}
	}
	log.Debugf("Finished writing from peer %v", pc.addr)
	pc.conn.Close()
}

func (pc *PeerConnection) readHandshake() (err error) {
	buf := make([]byte, messages.HandshakeLength)
	n, err := pc.conn.Read(buf)

	if err != nil {
		return
	}

	if n != messages.HandshakeLength {
		err = errors.New("Cannot read the handshake")
		return
	}

	var hand messages.Handshake
	err = binary.Read(bytes.NewBuffer(buf), binary.BigEndian, &hand)
	if err != nil {
		return
	}

	log.Debugf("Peer %v - Handshake: %v", pc.addr.String(), hand.String())
	return
}
