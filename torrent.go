package gotorrent

import (
	log "code.google.com/p/tcgl/applog"
	"github.com/moretti/gotorrent/bitarray"
	"github.com/moretti/gotorrent/metainfo"
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

	Pieces      []*Piece
	PeerManager *PeerManager
	Tracker     *Tracker

	ActivePieces    *bitarray.BitArray
	CompletedPieces *bitarray.BitArray

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

	metaInfo := readTorrent(torrent)

	t.Announce = metaInfo.Announce
	t.InfoHash = metaInfo.InfoHash
	t.CreationDate = metaInfo.CreationDate
	t.Name = metaInfo.Info.Name
	t.Length = metaInfo.Info.Length
	t.PieceHashes = metaInfo.Info.Pieces
	t.PieceLength = metaInfo.Info.PieceLength
	t.PieceCount = t.Length / t.PieceLength

	t.Pieces = make([]*Piece, t.PieceCount)
	for i := 0; i < t.PieceCount; i++ {
		hashIndex := i * 20
		t.Pieces[i] = NewPiece(i, t.PieceLength, t.PieceHashes[hashIndex:hashIndex+20])
	}

	t.ActivePieces = bitarray.New(t.PieceCount)
	t.CompletedPieces = bitarray.New(t.PieceCount)
	t.Tracker = NewTracker(t.Announce)
	t.PeerManager = NewPeerManager(t)

	log.Debugf("File Length: %v", t.Length)
	log.Debugf("Piece Length: %v", t.PieceLength)
	log.Debugf("Piece Count: %v", t.PieceCount)
	log.Debugf("Piece Hashes: %v", len(t.PieceHashes))

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

func (torrent *Torrent) Test() (err error) {
	trackerResponse, err := torrent.Tracker.Peers(
		torrent.InfoHash,
		torrent.ClientId,
		torrent.Port,
		torrent.Uploaded,
		torrent.Downloaded,
		torrent.Length,
	)

	log.Debugf("Len of addr: %v", len(trackerResponse.PeerAddresses))
	torrent.PeerManager.UpdatePeers(trackerResponse.PeerAddresses)

	time.Sleep(240 * time.Second)

	return
}
