package gotorrent

import (
	log "code.google.com/p/tcgl/applog"
	"crypto/sha1"
	"github.com/moretti/gotorrent/bitarray"
)

const (
	// MaxBlockLength is generally a power of two unless it gets truncated by the end of the file.
	// All current implementations use 2 15 (32 KB), and close connections which request an amount greater than 2 17.
	MaxBlockLength = 1024 * 16
)

// Piece vs Block:
// A piece refers to a portion of the downloaded data that is described
// in the metainfo file, which can be verified by a SHA1 hash.
// A block is a portion of data that a client may request from at least one peer.
// Two or more blocks make up a whole piece, which may then be verified
type Piece struct {
	completed *bitarray.BitArray
	requested *bitarray.BitArray
	hash      string
	index     int
	length    int

	data []byte
}

func NewPiece(index, length int, hash string) *Piece {
	p := new(Piece)

	p.completed = bitarray.New(length / MaxBlockLength)
	p.requested = bitarray.New(p.completed.Len())

	p.hash = hash
	p.index = index
	p.length = length

	p.data = make([]byte, length)

	return p
}

func (p *Piece) SetBlock(begin int, block []byte) {
	index := begin / MaxBlockLength
	if p.completed.IsSet(index) {
		log.Warningf("Attempt to overwrite data at piece %v, offset %v", p.index, begin)
		return
	}

	end := begin + len(block)

	if end > len(p.data) || begin < 0 {
		log.Warningf("Len of block: %v", len(block))
		log.Warningf("Index: %v", index)
		log.Warningf("len(data): %v", len(p.data))
		log.Warningf("Begin: %v", begin)
		log.Warningf("End: %v", end)
	}

	copy(p.data[begin:end], block)
	p.completed.Set(index)
}

type BlockRequest struct {
	Begin  int
	Length int
}

func (p *Piece) NextBlock() *BlockRequest {
	if p.completed.Cardinality() == p.completed.Len() {
		return nil
	}

	indices := p.requested.Or(p.completed).UnsetIndices()
	if len(indices) == 0 {
		return nil
	}

	index := indices[0]
	p.requested.Set(index)
	lastLength := p.length % MaxBlockLength

	blockRequest := new(BlockRequest)
	blockRequest.Begin = index * MaxBlockLength

	if index == p.completed.Len()-1 && lastLength != 0 {
		blockRequest.Length = lastLength
	} else {
		blockRequest.Length = MaxBlockLength
	}

	return blockRequest
}

func (p *Piece) IsValid() bool {
	hash := sha1.New()
	hash.Write(p.data)

	return string(hash.Sum(nil)) == p.hash
}

func (p *Piece) Len() int {
	return p.length
}

func (p *Piece) Index() int {
	return p.index
}
