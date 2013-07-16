package messages

import (
	"bytes"
	"fmt"
)

const (
	ChokeId = iota
	UnchokeId
	InterestedId
	NotInterestedId
	HaveId
	BitFieldId
	RequestId
	PieceId
	CancelId
	PortId
)

const (
	BitTorrentProtocol  = "BitTorrent protocol"
	HandshakeLength     = 68
	ChokeLength         = 1
	UnchokeLength       = 1
	InterestedLength    = 1
	NotInterestedLength = 1
	RequestLength       = 13
)

// handshake: <pstrlen><pstr><reserved><info_hash><peer_id>
type Handshake struct {
	Pstrlen  byte
	Pstr     [19]byte
	Reserved [8]byte
	InfoHash [20]byte
	PeerId   [20]byte
}

func NewHandshake(infoHash, peerId string) *Handshake {
	h := Handshake{
		Pstrlen:  byte(len(BitTorrentProtocol)),
		Reserved: [...]byte{0, 0, 0, 0, 0, 0, 0, 0},
	}
	// TODO: Is there a better way to convert a string to a byte array?
	copy(h.Pstr[:], BitTorrentProtocol)
	copy(h.InfoHash[:], infoHash)
	copy(h.PeerId[:], peerId)

	return &h
}

func (hand *Handshake) String() string {
	var buffer = new(bytes.Buffer)

	buffer.WriteString("{ ")
	fmt.Fprintf(buffer, "Pstrlen : %v ", hand.Pstrlen)
	fmt.Fprintf(buffer, "Pstr : %s ", string(hand.Pstr[:]))
	fmt.Fprintf(buffer, "Reserved : %s ", string(hand.Reserved[:]))
	fmt.Fprintf(buffer, "InfoHash : %s ", string(hand.InfoHash[:]))
	fmt.Fprintf(buffer, "PeerId : %s ", string(hand.PeerId[:]))
	buffer.WriteString(" }")

	return buffer.String()
}

// message: <length prefix><message ID><payload>
type Message struct {
	Length  uint32
	Id      byte
	Payload []byte
}

type Header struct {
	Length uint32
	Id     byte
}

// unchoke: <len=0001><id=1>
func NewUnchoke() *Header {
	u := Header{
		Length: UnchokeLength,
		Id:     UnchokeId,
	}
	return &u
}

// interested: <len=0001><id=2>
func NewInterested() *Header {
	u := Header{
		Length: InterestedLength,
		Id:     InterestedId,
	}
	return &u
}

// not interested: <len=0001><id=3>

// have: <len=0005><id=4><piece index>
type Have struct {
	Header     Header
	PieceIndex uint32
}

// bitArray: <len=0001+X><id=5><bitArray>
type BitArray struct {
	Header   Header
	BitField []byte
}

// request: <len=0013><id=6><index><begin><length>
type Request struct {
	Header      Header
	PieceIndex  uint32
	BlockOffset uint32
	BlockLength uint32
}

func NewRequest(pieceIndex, blockOffset, blockLength uint32) *Request {
	r := Request{
		Header: Header{
			Length: RequestLength,
			Id:     RequestId,
		},
		PieceIndex:  pieceIndex,
		BlockOffset: blockOffset,
		BlockLength: blockLength,
	}
	return &r
}

// piece: <len=0009+X><id=7><index><begin><block>
type Piece struct {
	Header      Header
	PieceIndex  uint32
	BlockOffset uint32
	BlockData   []byte
}

// cancel: <len=0013><id=8><index><begin><length>
type Cancel struct {
	Header      Header
	PieceIndex  uint32
	BlockOffset uint32
	BlockLength uint32
}
