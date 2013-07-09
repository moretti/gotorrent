package metainfo

import (
	"bytes"
	"code.google.com/p/bencode-go"
	log "code.google.com/p/tcgl/applog"
	"crypto/sha1"
	"io"
	"os"
)

// https://wiki.theory.org/BitTorrentSpecification#Metainfo_File_Structure
type MetaInfo struct {
	Info         InfoDict
	InfoHash     string "info hash"
	Announce     string
	AnnounceList string "announce-list"
	CreationDate int    "creation date"
	Comment      string
	CreatedBy    string "created by"
	Encoding     string
}

type InfoDict struct {
	PieceLength int "piece length"
	Pieces      string
	Private     int
	// Single File Mode
	Name   string
	Length int
	Md5sum string
	// Multiple File mode
	Files []FileDict
}

type FileDict struct {
	Length int
	Path   []string
	Md5sum string
}

func Read(file *os.File) (metaInfo *MetaInfo, err error) {
	metaInfo = new(MetaInfo)
	if err = bencode.Unmarshal(file, metaInfo); err != nil {
		return
	}

	if _, err = file.Seek(0, 0); err != nil {
		return
	}

	metaInfo.InfoHash = calculateInfoHash(file)
	log.Debugf("Info hash: %v", metaInfo.InfoHash)

	if err != nil {
		return
	}

	return
}

func calculateInfoHash(file io.Reader) (infoHash string) {
	result, err := bencode.Decode(file)
	if err != nil {
		panic(err)
	}

	metaInfoMap, ok := result.(map[string]interface{})
	if !ok {
		panic("Couldn't parse torrent file")
	}

	infoMap, ok := metaInfoMap["info"]
	if !ok {
		panic("Couldn't parse torrent info file.")
	}

	var b bytes.Buffer
	err = bencode.Marshal(&b, infoMap)

	if err != nil {
		panic(err)
	}

	hash := sha1.New()
	hash.Write(b.Bytes())

	infoHash = string(hash.Sum(nil))

	return
}
