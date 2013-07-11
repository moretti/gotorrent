package gotorrent

// The File struct is used by Torrent struct to read/write from files.
// The torrent will create one for every file in the torrent upon initialization.
type File struct {
}

func (file *File) Write(pieceOffset int, data []byte) {
	panic("Not implemented")
}
