package gotorrent

import (
	"bytes"
	log "code.google.com/p/tcgl/applog"
	"encoding/binary"
	"github.com/moretti/gotorrent/bitarray"
	"github.com/moretti/gotorrent/messages"
	"io"
	"net"
	"time"
)

type Peer struct {
	Addr     net.TCPAddr
	Conn     net.Conn
	Data     []byte
	BitField *bitarray.BitArray

	PieceCount  uint32
	PieceLength uint32

	//Handshake bool
	//Interested bool
	//Choked bool
}

func (peer *Peer) Connect() (err error) {
	log.Debugf("Connecting to %s...", peer.Addr.String())
	peer.Conn, err = net.DialTimeout("tcp", peer.Addr.String(), time.Second*5)
	if peer.Conn == nil || err != nil {
		log.Debugf("Unable to connect to %s", peer.Addr.String())
	}
	log.Debugf("Connected to %s", peer.Addr.String())
	return
}

func (peer *Peer) SendUnchoke() (err error) {
	unchoke := messages.NewUnchoke()
	err = binary.Write(peer.Conn, binary.BigEndian, unchoke)
	return
}

func (peer *Peer) SendInterested() (err error) {
	interested := messages.NewInterested()
	err = binary.Write(peer.Conn, binary.BigEndian, interested)
	return
}

func (peer *Peer) SendHandshake(infoHash, peerId string) (err error) {
	hand := messages.NewHandshake(infoHash, peerId)
	err = binary.Write(peer.Conn, binary.BigEndian, hand)
	return
}

func (peer *Peer) ReadHandshake() (err error) {
	buf := make([]byte, messages.HandshakeLength)
	n, err := peer.Conn.Read(buf)

	if err != nil || n != messages.HandshakeLength {
		log.Errorf("Cannot read the handshake: %v", err)
		return err
	}

	var hand messages.Handshake
	err = binary.Read(bytes.NewBuffer(buf), binary.BigEndian, &hand)
	if err != nil {
		log.Errorf("%v", err)
	}

	log.Debugf("Handshake from peer %v", peer.Addr.String())
	log.Debugf(hand.String())
	return err
}

func (peer *Peer) RequestBlock(pieceIndex, blockOffset, blockLenght uint32) (err error) {
	req := messages.NewRequest(pieceIndex, blockOffset, blockLenght)
	err = binary.Write(peer.Conn, binary.BigEndian, req)
	return
}

func (peer *Peer) ReceiveData() (err error) {
	nBytes := 0
	buf := make([]byte, 65536)
	for {
		peer.Conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, err := peer.Conn.Read(buf)

		readed := buf[:n]
		nBytes += n

		if err == io.EOF || n == 0 {
			break
		}
		if err != nil {
			log.Errorf("Error in reading: %v", err)
			return err
		}
		peer.Data = append(peer.Data, readed...)
	}
	log.Debugf("Readed %v, byte(s)", nBytes)
	//log.Debugf("%v", peer.Data)
	return
}

func (peer *Peer) ParseData() {
	for {
		if len(peer.Data) < 4 {
			break
			/*peer.SendUnchoke()
			peer.SendInterested()
			log.Debugf("Trying to request piece #%v", i)
			peer.RequestBlock(300, 0, uint32(PieceChunkLenght))
			peer.ReceiveData()
			i++*/
		}

		var message messages.Message
		err := binary.Read(bytes.NewBuffer(peer.Data[:4]), binary.BigEndian, &message.Length)
		if err != nil {
			panic(err)
		}

		if message.Length == 0 {
			log.Debugf("Keep alive")
			peer.Data = peer.Data[4:]

		} else {
			message.Id = peer.Data[4]
			data := peer.Data[:4+message.Length]

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
				var haveMsg messages.Have
				err = binary.Read(bytes.NewBuffer(data), binary.BigEndian, &haveMsg)
				log.Debugf("Peer %v - Have piece #%v", peer.Addr.String(), haveMsg.PieceIndex)

				if haveMsg.PieceIndex > peer.PieceCount {
					log.Errorf("Invalid piece index, piece index: %v, piece count: %v", haveMsg.PieceIndex, peer.PieceCount)
					break
				}

				peer.BitField.Set(int(haveMsg.PieceIndex), true)
				log.Debugf("Peer %v - BitField updated", peer.Addr.String())
			case messages.BitFieldId:
				log.Debugf("BitField")

				var bitMsg messages.BitArray
				err = binary.Read(bytes.NewBuffer(data), binary.BigEndian, &bitMsg)
				bitMsg.BitField = data[:message.Length-1]

				bitCount := (message.Length - 1) * 8
				if peer.PieceCount > bitCount {
					log.Errorf("Invalid BitField, bit count: %v, piece count: %v", bitCount, peer.PieceCount)
					break
				}
				peer.BitField = bitarray.NewFromBytes(bitMsg.BitField, int(peer.PieceCount))
				log.Debugf("Peer %v - New BitField:", peer.Addr.String())
			case messages.RequestId:
				log.Debugf("Request")
			case messages.PieceId:
				log.Debugf("Piece")
				var pieceMsg messages.Piece
				err = binary.Read(bytes.NewBuffer(data), binary.BigEndian, &pieceMsg)
				pieceMsg.BlockData = data[:message.Length-9]

				log.Debugf("Peer %v - Found a new block - PieceIndex: %v BlockOffset: %v", peer.Addr.String(), pieceMsg.PieceIndex, pieceMsg.BlockOffset)

			case messages.CancelId:
				log.Debugf("Cancel")
			case messages.PortId:
				log.Debugf("Port")
			}
			peer.Data = peer.Data[4+message.Length:]
		}
	}
	return
}
