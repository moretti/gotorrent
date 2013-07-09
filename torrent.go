package gotorrent

import (
	log "code.google.com/p/tcgl/applog"
	"github.com/moretti/gotorrent/bitarray"
	"github.com/moretti/gotorrent/metainfo"
	"net"
	"os"
	"strings"
	"time"
)

type Torrent struct {
	ClientId     ClientId
	Port         int
	DownloadPath string

	Downloaded int
	Uploaded   int

	//Pieces []Piece
	Peers   []*Peer
	Tracker *Tracker

	IsComplete bool
	BitField   *bitarray.BitArray

	Announce     string
	InfoHash     string
	CreationDate int
	Name         string
	Length       int
	PieceHashes  string
	PieceLength  int
	PieceCount   int
}

func NewTorrent(clientId ClientId, port int, torrent string, downloadPath string) *Torrent {
	t := new(Torrent)
	t.ClientId = clientId
	t.Port = port
	t.DownloadPath = downloadPath
	t.Downloaded = 0
	t.Uploaded = 0
	t.IsComplete = false

	metaInfo := readTorrent(torrent)

	t.Announce = metaInfo.Announce
	t.InfoHash = metaInfo.InfoHash
	t.CreationDate = metaInfo.CreationDate
	t.Name = metaInfo.Info.Name
	t.Length = metaInfo.Info.Length
	t.PieceHashes = metaInfo.Info.Pieces
	t.PieceLength = metaInfo.Info.PieceLength
	t.PieceCount = t.Length / t.PieceLength

	t.BitField = bitarray.New(t.PieceCount)
	t.Tracker = NewTracker(t.Announce)

	log.Debugf("File Length: %v", t.Length)
	log.Debugf("Piece Length: %v", t.PieceLength)
	log.Debugf("Piece Count: %v", t.PieceCount)
	return t
}

func readTorrent(torrent string) *metainfo.MetaInfo {
	if strings.HasPrefix(torrent, "http:") {
		panic("Not implemented")
	} else if strings.HasPrefix(torrent, "magnet:") {
		panic("Not implemented")
	} else {

		log.Debugf("Opening: %v", torrent)

		file, err := os.Open(torrent)
		if err != nil {
			panic(err)
		}

		defer func() {
			if err := file.Close(); err != nil {
				panic(err)
			}
		}()

		metaInfo, err := metainfo.Read(file)
		if err != nil {
			panic(err)
		}

		return metaInfo
	}
}

func (torrent *Torrent) UpdatePeers() {
	trackerResponse, err := torrent.Tracker.Peers(
		torrent.InfoHash,
		torrent.ClientId,
		torrent.Port,
		torrent.Uploaded,
		torrent.Downloaded,
		torrent.Length,
	)

	if err != nil {
		log.Errorf("Unable to retrieve the peers: %v", err)
	}

	addresses := make(map[string]bool)
	for _, peer := range torrent.Peers {
		addresses[peer.Addr.String()] = true
	}

	for _, peerAddr := range trackerResponse.PeerAddresses {
		if _, ok := addresses[peerAddr.String()]; !ok {
			log.Debugf("Found Peer: %v", peerAddr)
			torrent.AddPeer(peerAddr)
		}
	}
}

func (torrent *Torrent) AddPeer(peerAddr net.TCPAddr) {
	peer := Peer{
		Addr:        peerAddr,
		PieceLength: uint32(torrent.PieceLength),
		PieceCount:  uint32(torrent.PieceCount),
		BitField:    bitarray.New(torrent.PieceCount),
	}

	torrent.Peers = append(torrent.Peers, &peer)
}

func (torrent *Torrent) Test() (err error) {
	torrent.UpdatePeers()

	for _, peer := range torrent.Peers {
		go test(peer, torrent)
	}

	time.Sleep(240 * time.Second)

	return
}

func test(peer *Peer, torrent *Torrent) (err error) {
	log.Debugf("Fired!")
	err = peer.Connect()
	if err != nil || peer.Conn == nil {
		return
	}

	err = peer.SendHandshake(torrent.InfoHash, string(torrent.ClientId))
	if err != nil {
		log.Errorf("%v", err)
	}
	err = peer.SendUnchoke()
	if err != nil {
		log.Errorf("%v", err)
	}
	err = peer.SendInterested()
	if err != nil {
		log.Errorf("%v", err)
	}
	err = peer.RequestBlock(300, 0, uint32(PieceChunkLenght))
	if err != nil {
		log.Errorf("%v", err)
	}
	err = peer.ReadHandshake()

	log.Debugf("Receiving Data...")
	err = peer.ReceiveData()
	if err != nil {
		log.Errorf("%v", err)
	}
	peer.ParseData()
	return
}
