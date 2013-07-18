package gotorrent

import (
	"bytes"
	log "code.google.com/p/tcgl/applog"
	"encoding/binary"
	"errors"
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
	nBytes := 0
	n := 0
	var err error
	buf := make([]byte, 2048)

	err = pc.readHandshake()
	if err != nil {
		log.Debugf("Finished reading from peer %v, err: %v", pc.addr, err)
		pc.conn.Close()
	}

	for {
		pc.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, err = pc.conn.Read(buf)

		if err != nil && err != io.EOF {
			if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
				break
			}
		}

		readed := buf[:n]
		//log.Debugf("Peer %v - Readed %v bytes", pc.addr, n)

		nBytes += n
		pc.data = append(pc.data, readed...)

		//log.Debugf("Peer: %v - Buffer: %v", pc.addr, buf)
		//log.Debugf("Peer: %v - Readed: %v", pc.addr, readed)
		//log.Debugf("Peer: %v - Current data: %v", pc.addr, pc.data)

		if len(pc.data) > 4 {
			//log.Debugf("Peer: %v - Parsing the data", pc.addr)
			pc.parseData()
		}
	}

	log.Debugf("Finished reading from peer %v, err: %v", pc.addr, err)
	pc.conn.Close()
}

func (pc *PeerConnection) parseData() {
	var message messages.Message
	err := binary.Read(bytes.NewBuffer(pc.data[:4]), binary.BigEndian, &message.Header.Length)
	if err != nil {
		log.Errorf("Cannot read the message length: %v", err)
	}

	if int(message.Header.Length) > len(pc.data) {
		log.Debugf("Peer: %v - I need more data, message length: %v, data length %v", pc.addr, message.Header.Length, len(pc.data))
		return
	}

	message.Data = pc.data[:4]
	pc.data = pc.data[4:]

	if message.Header.Length > 0 {
		message.Header.Id = pc.data[0]
		message.Data = append(message.Data, pc.data[:message.Header.Length]...)

		pc.data = pc.data[message.Header.Length:]
	}
	//log.Debugf("Peer: %v - Sending message %v", pc.addr, message.Header)
	pc.outMessage(message)
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
