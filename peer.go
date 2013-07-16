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
	Connection *PeerConnection
	Torrent    *Torrent

	BitField   *bitarray.BitArray
	Interested bool
	Choked     bool
}

func NewPeer(
	addr net.TCPAddr,
	torrent *Torrent,
	peerErrors chan<- PeerError,
	outMessages chan<- PeerMessage,
) *Peer {

	p := new(Peer)
	p.Torrent = torrent
	p.Connection = NewPeerConnection(addr, peerErrors, outMessages)

	p.BitField = bitarray.New(p.Torrent.PieceCount)
	p.Interested = false
	p.Choked = true

	return p
}

func (p *Peer) String() string {
	return p.Connection.addr.String()
}

func (p *Peer) Connect() {
	go func() {
		p.Connection.Connect()
		p.SendHandshake(p.Torrent.InfoHash, string(p.Torrent.ClientId))
		p.SendUnchoke()
		p.SendInterested()
	}()
}

func (p *Peer) SendUnchoke() {
	p.Connection.SendMessage(messages.NewUnchoke())
}

func (p *Peer) SendInterested() {
	p.Connection.SendMessage(messages.NewInterested())
}

func (p *Peer) SendHandshake(infoHash, peerId string) {
	p.Connection.SendMessage(messages.NewHandshake(infoHash, peerId))
}

func (p *Peer) SetKeepAlive() {
}

func (p *Peer) SetChoke() {
	p.Choked = true
}

func (p *Peer) SetUnchoke() {
	p.Choked = false
}

func (p *Peer) SetHave(message []byte) {
	var haveMsg messages.Have
	err := binary.Read(bytes.NewBuffer(message), binary.BigEndian, &haveMsg)

	if err != nil {
		log.Errorf("Peer %v - Unable to parse the have message: %v", p.String(), err)
		return
	}

	if p.BitField == nil {
		log.Errorf("Peer %v - Bitfield not initialized yet", p.String())
		return
	}

	if int(haveMsg.PieceIndex) >= p.BitField.Len() {
		log.Errorf("Peer %v - Invalid piece index: %v, piece count: %v", haveMsg.PieceIndex, p.BitField.Len())
		return
	}

	log.Debugf("Peer %v - Have piece #%v", p.String(), haveMsg.PieceIndex)
	p.BitField.Set(int(haveMsg.PieceIndex), true)
}

func (p *Peer) SetBitField(message []byte) {
	var bitMsg messages.BitArray
	err := binary.Read(bytes.NewBuffer(message), binary.BigEndian, &bitMsg.Header)

	if err != nil {
		log.Errorf("Peer %v - Unable to parse the bitfield message: %v", p.String(), err)
		return
	}

	bitMsg.BitField = message[:bitMsg.Header.Length-1]
	pieceCount := p.Torrent.PieceCount

	if bitCount := (bitMsg.Header.Length - 1) * 8; pieceCount > int(bitCount) {
		log.Errorf("Peer %v - Invalid BitField, bit count: %v, piece count: %v", p.String(), bitCount, pieceCount)
		return
	}
	p.BitField = bitarray.NewFromBytes(bitMsg.BitField, pieceCount)
	log.Debugf("Peer %v - New BitField:", p.String())
}
