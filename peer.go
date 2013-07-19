package gotorrent

import (
	log "code.google.com/p/tcgl/applog"
	"github.com/moretti/gotorrent/bitarray"
	"github.com/moretti/gotorrent/messages"
	"net"
)

const (
	PeerMaxRequests = 16
)

type Peer struct {
	connection *PeerConnection
	torrent    *Torrent

	bitField     *bitarray.BitArray
	amInterested bool
	IsChoked     bool

	RequestsCount int
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

func (p *Peer) SetHave(message messages.Message) {
	haveMsg, err := message.ToHave()

	if err != nil {
		log.Errorf("Peer %v - Unable to parse the have message: %v", p.String(), err)
		return
	}

	if int(haveMsg.PieceIndex) >= p.bitField.Len() {
		log.Errorf("Peer %v - Have message, invalid piece index: %v, piece count: %v", p.String(), haveMsg.PieceIndex, p.bitField.Len())
		return
	}

	log.Debugf("Peer %v - Have piece #%v", p.String(), haveMsg.PieceIndex)
	p.bitField.Set(int(haveMsg.PieceIndex))
}

func (p *Peer) SetBitField(message messages.Message) {
	bitMsg, err := message.ToBitArray()

	if err != nil {
		log.Errorf("Peer %v - Unable to parse the bitarray message: %v", p.String(), err)
		return
	}

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
		p.amInterested = interested
		p.SendInterested(interested)
	}

	//log.Debugf("Peer %v - Piece indices that I don't have: %v", p.String(), pieces)
	return
}
