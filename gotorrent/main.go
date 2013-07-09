package main

import (
	"fmt"
	"github.com/moretti/gotorrent"
	"os"
)

func main() {
	args := os.Args

	if len(args) != 2 {
		fmt.Println("Usage: gotorrent torrent_file")
	}

	torrentPath := args[1]
	client := gotorrent.NewClient()
	torrent := client.AddTorrent(torrentPath)
	torrent.Test()
}
