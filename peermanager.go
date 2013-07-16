package gotorrent

import (
	log "code.google.com/p/tcgl/applog"
	"github.com/moretti/gotorrent/messages"
	"net"
)

type PeerManager struct {
	Torrent *Torrent
	Peers   map[string]*Peer

	PiecesCount []int

	AddPeerAddr chan net.TCPAddr
	RemovePeer  chan Peer

	Errors     chan PeerError
	InMessages chan PeerMessage

	Quit <-chan bool
}

func NewPeerManager(torrent *Torrent) *PeerManager {
	pm := new(PeerManager)
	pm.Torrent = torrent
	pm.Peers = make(map[string]*Peer)

	pm.AddPeerAddr = make(chan net.TCPAddr)
	pm.RemovePeer = make(chan Peer)

	pm.Errors = make(chan PeerError)
	pm.InMessages = make(chan PeerMessage)

	pm.Quit = make(<-chan bool)

	go pm.manage()

	return pm
}

func (pm *PeerManager) manage() {
	for {
		select {
		case peerAddr := <-pm.AddPeerAddr:
			pm.addPeer(peerAddr)
		case <-pm.Quit:
			log.Debugf("Quitting...")
			return
		case message := <-pm.InMessages:
			pm.processMessage(message)
		case err := <-pm.Errors:
			log.Errorf("Peer %v: %v", err.Addr.String(), err.Err)
			/*case <-pm.removePeer:
			panic("Not implemented")*/
		}
	}
}

func (pm *PeerManager) addPeer(peerAddr net.TCPAddr) {
	strAddr := peerAddr.String()
	if _, ok := pm.Peers[strAddr]; !ok {
		log.Debugf("Found Peer: %v", peerAddr)

		peer := NewPeer(
			peerAddr,
			pm.Torrent,
			pm.Errors,
			pm.InMessages)

		pm.Peers[strAddr] = peer

		peer.Connect()
	}
}

func (pm *PeerManager) processMessage(peerMessage PeerMessage) {
	message := peerMessage.Message
	data := peerMessage.Message.Data

	peer, ok := pm.Peers[peerMessage.Addr.String()]
	if !ok {
		log.Errorf("Unable to find peer %v", peerMessage.Addr)
	}

	if message.Header.Length == 0 {
		peer.SetKeepAlive()
	} else {
		switch message.Header.Id {
		case messages.ChokeId:
			peer.SetChoke()
		case messages.UnchokeId:
			peer.SetUnchoke()
		case messages.InterestedId:
		case messages.NotInterestedId:
		case messages.HaveId:
			peer.SetHave(data)
		case messages.BitFieldId:
			peer.SetBitField(data)
		case messages.RequestId:
		case messages.PieceId:
			log.Debugf("Piece")
			/*var pieceMsg messages.Piece
			err = binary.Read(bytes.NewBuffer(data), binary.BigEndian, &pieceMsg)
			pieceMsg.BlockData = data[:message.Length-9]

			log.Debugf("Peer %v - Found a new block - PieceIndex: %v BlockOffset: %v", pc.Addr.String(), pieceMsg.PieceIndex, pieceMsg.BlockOffset)*/
		case messages.CancelId:
		case messages.PortId:
		}
	}
}

func (pm *PeerManager) UpdatePeers(addresses []net.TCPAddr) {
	for _, peerAddr := range addresses {
		pm.AddPeerAddr <- peerAddr
	}
}

func (PeerManager *PeerManager) DownloadPiece(number int) (err error) {
	panic("Not implemented")
}
