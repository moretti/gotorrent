package gotorrent

import (
	"bytes"
	log "code.google.com/p/tcgl/applog"
	"encoding/binary"
	"github.com/moretti/gotorrent/bitarray"
	"github.com/moretti/gotorrent/messages"
	"net"
)

type Peer struct {
	connection *PeerConnection
	torrent    *Torrent

	bitField   *bitarray.BitArray
	interested bool
	choked     bool
}

func NewPeer(
	addr net.TCPAddr,
	torrent *Torrent,
	peerErrors chan<- PeerError,
	outMessages chan<- PeerMessage,
) *Peer {

	p := new(Peer)
	p.torrent = torrent
	p.connection = NewPeerConnection(addr, peerErrors, outMessages)

	p.bitField = bitarray.New(p.torrent.PieceCount)
	p.interested = false
	p.choked = true

	return p
}

func (p *Peer) String() string {
	return p.connection.addr.String()
}

func (p *Peer) Connect() {
	go func() {
		p.connection.Connect()
		p.SendHandshake(p.torrent.InfoHash, string(p.torrent.ClientId))
		p.SendUnchoke()
		p.SendInterested()
	}()
}

func (p *Peer) SendUnchoke() {
	p.connection.SendMessage(messages.NewUnchoke())
}

func (p *Peer) SendInterested() {
	p.connection.SendMessage(messages.NewInterested())
}

func (p *Peer) SendHandshake(infoHash, peerId string) {
	p.connection.SendMessage(messages.NewHandshake(infoHash, peerId))
}

func (p *Peer) SetKeepAlive() {
}

func (p *Peer) SetChoke() {
	p.choked = true
}

func (p *Peer) SetUnchoke() {
	p.choked = false
}

func (p *Peer) SetHave(message []byte) {
	var haveMsg messages.Have
	err := binary.Read(bytes.NewBuffer(message), binary.BigEndian, &haveMsg)

	if err != nil {
		log.Errorf("Peer %v - Unable to parse the have message: %v", p.String(), err)
		return
	}

	if int(haveMsg.PieceIndex) >= p.bitField.Len() {
		log.Errorf("Peer %v - Invalid piece index: %v, piece count: %v", haveMsg.PieceIndex, p.bitField.Len())
		return
	}

	log.Debugf("Peer %v - Have piece #%v", p.String(), haveMsg.PieceIndex)
	p.bitField.Set(int(haveMsg.PieceIndex))
}

func (p *Peer) SetBitField(message []byte) {
	var bitMsg messages.BitArray
	err := binary.Read(bytes.NewBuffer(message), binary.BigEndian, &bitMsg.Header)

	if err != nil {
		log.Errorf("Peer %v - Unable to parse the bitfield message: %v", p.String(), err)
		return
	}

	begin := 5
	bitMsg.BitField = message[begin:]
	pieceCount := p.torrent.PieceCount

	if bitCount := (bitMsg.Header.Length - 1) * 8; pieceCount > int(bitCount) {
		log.Errorf("Peer %v - Invalid BitField, bit count: %v, piece count: %v", p.String(), bitCount, pieceCount)
		return
	}
	p.bitField = bitarray.NewFromBytes(bitMsg.BitField, pieceCount)
	log.Debugf("Peer %v - New BitField:", p.String())
}

func (p *Peer) BitField() *bitarray.BitArray {
	return p.bitField
}
