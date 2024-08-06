package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

func getPeers(torrentMeta TorrentMeta) []string {
	tracker := fromTorrentMeta(torrentMeta)
	return tracker.Peers
}

func peerHandshake(peerUrl string, infoHash []byte) (net.Conn, error) {
	conn, err := net.Dial("tcp", peerUrl)
	if err != nil {
		fmt.Println("error establishing connection to peer")
		return nil, err
	}
	err = conn.SetReadDeadline(time.Now().Add(peerHadshakeTimeout))
	if err != nil {
		fmt.Println("set deadline failed")
		return nil, err
	}

	handshakeMsg := createHandshakeMessage(infoHash)

	_, err = conn.Write(handshakeMsg)
	if err != nil {
		fmt.Println("error sending handshake to peer")
		return nil, err
	}

	buf := make([]byte, 68)
	_, err = conn.Read(buf)
	if err != nil {
		fmt.Println("error reading handshake from peer")
		return nil, err
	}

	return conn, nil
}

func exchangePeerMessages(conn net.Conn, peer string) error {
	peerMessage, err := receiveMessageFromPeer(conn)
	if err != nil {
		return fmt.Errorf("[%s] Error receiving bitfield message", peer)
	}

	if peerMessage.id != bitfield {
		return fmt.Errorf("[%s] Expected bitfield message", peer)
	}

	err = sendMessageToPeer(conn, []byte{0, 0, 0, 1, interested})
	if err != nil {
		return fmt.Errorf("[%s] Error sending interested message to peer", peer)
	}

	peerMessage, err = receiveMessageFromPeer(conn)
	if err != nil {
		return fmt.Errorf("[%s] Error receiving peer unchoke message", peer)
	}

	if peerMessage.id != unchoke {
		return fmt.Errorf("[%s] Expected unchoke message", peer)
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

func receiveMessageFromPeer(conn net.Conn) (PeerMessage, error) {
	peerMessage := PeerMessage{}

	buf := make([]byte, 4)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println(err)
		return peerMessage, err
	}

	peerMessage.lengthPrefix = binary.BigEndian.Uint32(buf)

	// Message was a keep alive so we ignore it and read the next one
	if peerMessage.lengthPrefix == 0 {
		receiveMessageFromPeer(conn)
	}

	payloadBuf := make([]byte, peerMessage.lengthPrefix)
	_, err = conn.Read(payloadBuf)
	if err != nil {
		fmt.Println(err)
		return peerMessage, err
	}
	peerMessage.id = payloadBuf[0]

	return peerMessage, nil
}

func receiveDataMessageFromPeer(conn net.Conn) ([]byte, error) {
	buf := make([]byte, 4)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	lengthPrefix := binary.BigEndian.Uint32(buf)

	// Message was a keep alive so we ignore it and read the next one
	if lengthPrefix == 0 {
		receiveMessageFromPeer(conn)
	}

	payloadBuf := make([]byte, lengthPrefix)
	_, err = io.ReadFull(conn, payloadBuf)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	messageId := payloadBuf[0]
	if messageId != piece {
		return nil, fmt.Errorf("expected piece message")
	}

	return payloadBuf[9:], nil
}
