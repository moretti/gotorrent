package gotorrent

import (
	//log "code.google.com/p/tcgl/applog"
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
