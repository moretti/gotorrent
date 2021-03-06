package gotorrent

import (
	log "code.google.com/p/tcgl/applog"
	"github.com/moretti/gotorrent/messages"
	"math/rand"
	"net"
)

type PeerManager struct {
	Torrent *Torrent
	Peers   map[string]*Peer

	AddPeerAddr         chan net.TCPAddr
	PeerReadyToDownload chan Peer

	Errors     chan PeerError
	InMessages chan PeerMessage

	Quit <-chan bool
}

func NewPeerManager(torrent *Torrent) *PeerManager {
	pm := new(PeerManager)
	pm.Torrent = torrent
	pm.Peers = make(map[string]*Peer)

	pm.AddPeerAddr = make(chan net.TCPAddr)
	pm.PeerReadyToDownload = make(chan Peer)

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
		case peerMessage := <-pm.InMessages:
			//log.Debugf("Receiving message %v from peer %v", peerMessage.Message.Header, peerMessage.Addr)
			pm.processMessage(peerMessage)
		case peerError := <-pm.Errors:
			pm.handleError(peerError)
		case <-pm.Quit:
			log.Debugf("Quitting...")
			return
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

func (pm *PeerManager) handleError(peerError PeerError) {
	log.Errorf("Peer %v: %v", peerError.Addr.String(), peerError.Err)
	delete(pm.Peers, peerError.Addr.String())
}

func (pm *PeerManager) processMessage(peerMessage PeerMessage) {
	message := peerMessage.Message

	peer, ok := pm.Peers[peerMessage.Addr.String()]
	if !ok {
		log.Errorf("Unable to find peer %v", peerMessage.Addr)
		return
	}

	if message.Header.Length == 0 {
		log.Debugf("Peer %v - Keep alive", peer.String())
		peer.SetKeepAlive()
	} else {
		switch message.Header.Id {
		case messages.ChokeId:
			log.Debugf("Peer %v - Chocked", peer.String())
			peer.IsChoked = true
		case messages.UnchokeId:
			peer.IsChoked = false
			log.Debugf("Peer %v - Unchocked", peer.String())
			pm.downloadPiece(peer)
		case messages.InterestedId:
		case messages.NotInterestedId:
		case messages.HaveId:
			peer.SetHave(message)
			pm.downloadPiece(peer)
		case messages.BitFieldId:
			peer.SetBitField(message)
			pm.downloadPiece(peer)
		case messages.RequestId:
		case messages.PieceId:
			peer.RequestsCount--
			pm.processPiece(message, peer)
		case messages.CancelId:
		case messages.PortId:
		default:
			log.Errorf("Peer %v - Unknown id: %v", peer.String(), message.Header.Id)
		}
	}
}

func (pm *PeerManager) downloadPiece(peer *Peer) {
	if peer.IsChoked || peer.RequestsCount > PeerMaxRequests {
		return
	}

	amInterested, pieceIndices := peer.AmInterested(pm.Torrent.ActivePieces, pm.Torrent.CompletedPieces)
	if !amInterested {
		return
	}

	// Choose a random piece that I don't have
	pieceIndex := randomChoice(pieceIndices)
	piece := pm.Torrent.Pieces[pieceIndex]

	peer.RequestsCount++
	log.Debugf("Requesting piece #%v to peer %v", pieceIndex, peer.String())
	peer.RequestPiece(piece)
}

func (pm *PeerManager) processPiece(message messages.Message, peer *Peer) {
	pieceMsg, err := message.ToPiece()

	if err != nil {
		log.Errorf("Peer %v - Unable to parse the piece message: %v", peer.String(), err)
		return
	}

	log.Debugf("Peer %v - Found a new block - PieceIndex: %v BlockOffset: %v", peer.String(), pieceMsg.PieceIndex, pieceMsg.BlockOffset)

	piece := pm.Torrent.Pieces[pieceMsg.PieceIndex]
	piece.SetBlock(int(pieceMsg.BlockOffset), pieceMsg.BlockData)
}

func randomChoice(slice []int) int {
	randomIndex := rand.Intn(len(slice))
	return slice[randomIndex]
}

func (pm *PeerManager) UpdatePeers(addresses []net.TCPAddr) {
	for _, peerAddr := range addresses {
		pm.AddPeerAddr <- peerAddr
	}
}
