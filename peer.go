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

	bitField     *bitarray.BitArray
	amInterested bool
	IsChoked     bool
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
	p.IsChoked = true
	p.amInterested = false

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
		//p.SendInterested()
	}()
}

func (p *Peer) SendUnchoke() {
	log.Debugf("Peer %v - Sending unchoke", p.String())
	p.connection.SendMessage(messages.NewUnchoke())
}

func (p *Peer) SendInterested(interested bool) {
	if interested {
		log.Debugf("Peer %v - Sending interested", p.String())
		p.connection.SendMessage(messages.NewInterested())
	} else {
		log.Debugf("Peer %v - Sending not interested", p.String())
		p.connection.SendMessage(messages.NewNotInterested())
	}
}

func (p *Peer) SendHandshake(infoHash, peerId string) {
	p.connection.SendMessage(messages.NewHandshake(infoHash, peerId))
}

func (p *Peer) SendRequest(pieceIndex, blockOffset, blockLength int) {
	p.connection.SendMessage(messages.NewRequest(uint32(pieceIndex), uint32(blockOffset), uint32(blockLength)))
}

func (p *Peer) SetKeepAlive() {
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
		log.Errorf("Peer %v - Invalid bitfield, bit count: %v, piece count: %v", p.String(), bitCount, pieceCount)
		return
	}
	p.bitField = bitarray.NewFromBytes(bitMsg.BitField, pieceCount)
	log.Debugf("Peer %v - New BitField:", p.String())
}

func (p *Peer) BitField() *bitarray.BitArray {
	return p.bitField
}

func (p *Peer) RequestPiece(piece *Piece) {
	for {
		block := piece.NextBlock()
		if block == nil {
			break
		}
		p.SendRequest(piece.Index(), block.Begin, block.Length)
	}
}

func (p *Peer) AmInterested(activePieces, completedPieces *bitarray.BitArray) (interested bool, pieces []int) {
	// pieces = peerHas ^ (peerHas & (active | completed))
	pieces = p.BitField().Xor(p.bitField.And(activePieces.Or(completedPieces))).SetIndices()
	interested = len(pieces) > 0

	if interested != p.amInterested {
		p.SendInterested(interested)
		p.amInterested = interested
	}

	log.Debugf("Peer %v - Those are the indices of the pieces that I don't have: %v", p.String(), pieces)
	return
}
