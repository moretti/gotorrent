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

	if message.Length == 0 {
		log.Debugf("Keep alive")
	} else {
		switch message.Id {
		case messages.ChokeId:
			log.Debugf("Choke")
		case messages.UnchokeId:
			log.Debugf("Unchoke")
		case messages.InterestedId:
			log.Debugf("Interested")
		case messages.NotInterestedId:
			log.Debugf("NotInterested")
		case messages.HaveId:
			log.Debugf("Have")
			/*var haveMsg messages.Have
			err = binary.Read(bytes.NewBuffer(data), binary.BigEndian, &haveMsg)
			if err != nil
			SendChannelMessage(haveMsg)

				log.Debugf("Peer %v - Have piece #%v", pc.Addr.String(), haveMsg.PieceIndex)

				if haveMsg.PieceIndex > pc.PieceCount {
					log.Errorf("Invalid piece index, piece index: %v, piece count: %v", haveMsg.PieceIndex, pc.PieceCount)
					break
				}

				pc.BitField.Set(int(haveMsg.PieceIndex), true)
				log.Debugf("Peer %v - BitField updated", pc.Addr.String())*/
		case messages.BitFieldId:
			log.Debugf("BitField")
			/*

				var bitMsg messages.BitArray
				err = binary.Read(bytes.NewBuffer(data), binary.BigEndian, &bitMsg)
				bitMsg.BitField = data[:message.Length-1]

				bitCount := (message.Length - 1) * 8
				if pc.PieceCount > bitCount {
					log.Errorf("Invalid BitField, bit count: %v, piece count: %v", bitCount, pc.PieceCount)
					break
				}
				pc.BitField = bitarray.NewFromBytes(bitMsg.BitField, int(pc.PieceCount))
				log.Debugf("Peer %v - New BitField:", pc.Addr.String())*/
		case messages.RequestId:
			log.Debugf("Request")
		case messages.PieceId:
			log.Debugf("Piece")
			/*var pieceMsg messages.Piece
			err = binary.Read(bytes.NewBuffer(data), binary.BigEndian, &pieceMsg)
			pieceMsg.BlockData = data[:message.Length-9]

			log.Debugf("Peer %v - Found a new block - PieceIndex: %v BlockOffset: %v", pc.Addr.String(), pieceMsg.PieceIndex, pieceMsg.BlockOffset)*/
		case messages.CancelId:
			log.Debugf("Cancel")
		case messages.PortId:
			log.Debugf("Port")
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
