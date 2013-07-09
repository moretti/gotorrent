package gotorrent

import (
	"bytes"
	"code.google.com/p/bencode-go"
	log "code.google.com/p/tcgl/applog"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

type Tracker struct {
	Announce string
}

// https://wiki.theory.org/BitTorrentSpecification#Tracker_Response
type TrackerResponse struct {
	FailureReason  string "failure reason"
	WarningMessage string "warning message"
	Interval       int
	MinInterval    int    "min interval"
	TrackerId      string "tracker id"
	Complete       int
	Incomplete     int
	BinaryPeers    string "peers"
	PeerAddresses  []net.TCPAddr
}

func NewTracker(announce string) *Tracker {
	t := new(Tracker)
	t.Announce = announce
	return t
}

func (tracker Tracker) Peers(infoHash string, clientId ClientId, port, uploaded, downloaded, left int) (trackerResponse *TrackerResponse, err error) {
	v := url.Values{}

	v.Set("info_hash", infoHash)
	v.Add("peer_id", string(clientId))
	v.Add("port", strconv.FormatInt(int64(port), 10))
	v.Add("uploaded", strconv.FormatInt(int64(uploaded), 10))
	v.Add("downloaded", strconv.FormatInt(int64(downloaded), 10))
	v.Add("left", strconv.FormatInt(int64(left), 10))
	v.Add("compact", strconv.FormatInt(1, 10))

	query := v.Encode()
	uri := tracker.Announce + "?" + query

	log.Debugf("Contacting: %v", tracker.Announce)
	log.Debugf("%v", uri)

	httpResp, err := http.Get(uri)
	defer func() {
		if err := httpResp.Body.Close(); err != nil {
			return
		}
	}()
	if err != nil {
		return
	}

	if httpResp.StatusCode != 200 {
		//buf := new(bytes.Buffer)
		//buf.ReadFrom(resp.Body)
		//response := buf.String()
		err = fmt.Errorf("Unable to contact the tracker. Status code: %v", httpResp.StatusCode)
		return
	}

	trackerResponse = new(TrackerResponse)
	err = bencode.Unmarshal(httpResp.Body, trackerResponse)
	if err != nil {
		return
	}
	trackerResponse.PeerAddresses, err = parsePeerAddresses(trackerResponse.BinaryPeers)

	if err != nil {
		return
	}

	return
}

// binaryPeers is a string consisting of multiples of 6 bytes.
// First 4 bytes are the IP address and last 2 bytes are the port number.
// All in network (big endian) notation.
func parsePeerAddresses(binaryPeers string) (peers []net.TCPAddr, err error) {
	peersCount := len(binaryPeers) / 6
	peers = make([]net.TCPAddr, peersCount, peersCount)

	for i := 0; i < len(binaryPeers); i += 6 {
		peer := binaryPeers[i : i+6]

		ip := net.IPv4(
			peer[0],
			peer[1],
			peer[2],
			peer[3])

		var port uint16
		err = binary.Read(bytes.NewBufferString(peer[4:]), binary.BigEndian, &port)
		if err != nil {
			return
		}

		peers[i/6] = net.TCPAddr{ip, int(port), ""}
	}

	return
}
