package gotorrent

import (
	"github.com/moretti/gotorrent/bitarray"
)

const (
	PieceChunkLenght = 1024 * 16
)

type Piece struct {
	Complete  *bitarray.BitArray
	Requested *bitarray.BitArray
	File      *File
	Hash      string
	Index     int
	Offset    int
	Length    int
}

func NewPiece(index, offset, length int, hash string, file *File) *Piece {
	p := new(Piece)

	p.Complete = bitarray.New(length / PieceChunkLenght)
	p.Requested = bitarray.New(p.Complete.Len())

	p.File = file
	p.Hash = hash
	p.Index = index
	p.Offset = offset
	p.Length = length

	return p
}
