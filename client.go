package gotorrent

import (
//log "code.google.com/p/tcgl/applog"
)

type Client struct {
	Id           ClientId
	Torrents     []*Torrent
	DownloadPath string
	Port         int
}

func NewClient() *Client {
	c := new(Client)
	c.Id = NewClientId()
	c.Torrents = []*Torrent{}

	// TODO: Read these options from the command line
	c.Port = 6881
	c.DownloadPath = "."

	return c
}

func (client *Client) AddTorrent(path string) *Torrent {
	torrent := NewTorrent(
		client.Id,
		client.Port,
		path,
		client.DownloadPath,
	)

	client.Torrents = append(client.Torrents, torrent)
	return torrent
}

func (client *Client) RemoveTorrent(torrent Torrent) {
	panic("Not implemented")
}
