package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

const peerHadshakeTimeout time.Duration = time.Duration(5 * time.Second)
const (
	piece      = 7
	request    = 6
	bitfield   = 5
	interested = 2
	unchoke    = 1
)

type PeerDataMessage struct {
	id   uint8
	data []byte
}

type Bitfield []byte

func (bf Bitfield) hasPiece(index int) bool {
	byteIndex := index / 8
	offset := index % 8
	return bf[byteIndex]>>(7-offset)&1 != 0
}

func getPeers(torrentMeta TorrentMeta) []string {
	tracker := fromTorrentMeta(torrentMeta)
	return tracker.Peers
}

func peerHandshake(peerUrl string, infoHash []byte) (net.Conn, error) {
	conn, err := net.Dial("tcp", peerUrl)
	if err != nil {
		return nil, errors.New("error establishing connection to peer")
	}
	err = conn.SetReadDeadline(time.Now().Add(peerHadshakeTimeout))
	if err != nil {
		return nil, errors.New("set deadline failed")
	}

	handshakeMsg := createHandshakeMessage(infoHash)

	_, err = conn.Write(handshakeMsg)
	if err != nil {
		return nil, errors.New("error sending handshake to peer")
	}

	buf := make([]byte, 68)
	_, err = conn.Read(buf)
	if err != nil {
		return nil, errors.New("error reading handshake from peer")
	}

	return conn, nil
}

func exchangePeerMessages(conn net.Conn, piece int) error {
	// receive bitfield message
	bitfield, err := receiveBitfieldMessageFromPeer(conn)
	if err != nil {
		return errors.New("error receiving bitfield message")
	}

	if !bitfield.hasPiece(piece) {
		return errors.New("peer doesn't have requested piece")
	}

	// send interested message
	err = sendMessageToPeer(conn, []byte{0, 0, 0, 1, interested})
	if err != nil {
		return errors.New("error sending interested message to peer")
	}

	// receive unchoke message
	peerMessage, err := receiveMessageFromPeer(conn)
	if err != nil {
		return errors.New("error receiving peer unchoke message")
	}

	if peerMessage.id != unchoke {
		return errors.New("expected unchoke message")
	}
	return nil
}

func createHandshakeMessage(infoHash []byte) []byte {
	pstrlen := byte(19)
	pstr := []byte("BitTorrent protocol")
	reserved := make([]byte, 8)
	handshake := append([]byte{pstrlen}, pstr...)
	handshake = append(handshake, reserved...)
	handshake = append(handshake, infoHash...)
	handshake = append(handshake, []byte{0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9}...)
	return handshake
}

func sendMessageToPeer(conn net.Conn, message []byte) error {
	_, err := conn.Write(message)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func receiveMessageFromPeer(conn net.Conn) (PeerDataMessage, error) {
	peerMessage := PeerDataMessage{}

	payloadBuf, err := readMessage(conn)
	if err != nil {
		return peerMessage, err
	}

	peerMessage.id = payloadBuf[0]

	return peerMessage, nil
}

func receivePieceMessageFromPeer(conn net.Conn) (PeerDataMessage, error) {
	peerMessage := PeerDataMessage{}

	payloadBuf, err := readMessage(conn)
	if err != nil {
		return peerMessage, err
	}

	peerMessage.id = payloadBuf[0]
	peerMessage.data = payloadBuf[9:]

	return peerMessage, nil
}

func receiveBitfieldMessageFromPeer(conn net.Conn) (Bitfield, error) {
	peerMessage := PeerDataMessage{}

	payloadBuf, err := readMessage(conn)
	if err != nil {
		return nil, err
	}

	peerMessage.id = payloadBuf[0]
	if peerMessage.id != bitfield {
		return nil, errors.New("expected bitfield message")
	}

	return payloadBuf[1:], nil
}

func readMessage(conn net.Conn) ([]byte, error) {
	buf := make([]byte, 4)
	_, err := conn.Read(buf)
	if err != nil {
		return nil, errors.New("failed to read length prefix")
	}

	lengthPrefix := binary.BigEndian.Uint32(buf)

	// Message was a keep alive so we ignore it and read the next one
	if lengthPrefix == 0 {
		receiveMessageFromPeer(conn)
	}

	payloadBuf := make([]byte, lengthPrefix)
	_, err = io.ReadFull(conn, payloadBuf)
	if err != nil {
		return nil, errors.New("failed to read payload")
	}

	return payloadBuf, nil
}
