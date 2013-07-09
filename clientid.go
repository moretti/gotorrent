package gotorrent

import (
	"math/rand"
	"time"
)

const (
	ClientIdLength = 20
	ClientAcronym  = "-GT"
	ClientVersion  = "0001"
	AlphaDigits    = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ0123456789"
)

type ClientId string

func NewClientId() ClientId {
	clientId := ClientAcronym + ClientVersion
	return ClientId(clientId + randomString(ClientIdLength-len(clientId)))
}

func randomString(size int) string {
	buf := make([]byte, size)

	for i := 0; i < size; i++ {
		buf[i] = AlphaDigits[rand.Intn(len(AlphaDigits))]
	}

	return string(buf)
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
