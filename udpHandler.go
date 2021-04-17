package main

import (
	"encoding/json"
	"log"
	"net"
	"strconv"
	"time"
)

func sendPlayerAllPositions(conn net.UDPConn, playerId int64) {
	for {
		for _, player := range playerList {
			if player.id == playerId {
				continue
			}
			if player.InVent {
				player.PlayerPosition.Y = -100
			}
			playerToSendJSON, err := json.Marshal(player)
			if err != nil {
				log.Println("Cant marshal location")
			}
			playerPacket, err := encapsulatePacketID(PositionBroadcast, playerToSendJSON)
			if err != nil {
				log.Println("Didn't encapsulate currUserJSON with ID")
			}
			conn.Write(stampPacketLength(playerPacket))
		}
		time.Sleep(20 * time.Millisecond) // sent 50 times a second
	}
}

func makeUdpConnection(udpConnection net.UDPConn, tcpConn net.Conn, currUser player) error {
	go handleUdpPlayer(udpConnection, tcpConn, currUser)
	return nil
}

func handleUdpPlayer(conn net.UDPConn, tcpConn net.Conn, currUser player) {
	log.Println("Accepted new connection.")
	defer conn.Close()
	defer log.Println("Closed connection.")

	for {
		// Read the first 4 bytes and see the packet length
		buf := make([]byte, 4)
		size, err := conn.Read(buf)
		if err != nil {
			log.Println("Cant read first 4 bytes!")
			err = nil
			return
		}
		data := buf[:size]
		readLength, err := strconv.Atoi(string(data))
		if err != nil {
			log.Println("Cant convent string size to int!")
			return
		}

		buf = make([]byte, readLength)
		size, err = conn.Read(buf)
		if err != nil {
			log.Println("Cant read the rest of the packet!")
			return
		}
		data = buf[:size]

		// Get the packet ID from the JSON
		var packet packetType
		err = json.Unmarshal(data, &packet)
		if err != nil {
			log.Println("Couldn't parse json player data! Skipping iteration!")
			continue
		}

		switch packet.ID {
		case UpdatePos:
			var newPosition position
			data, err := packet.dataToBytes()
			if err != nil {
				log.Println("Cant turn inteface to []byte!")
				return
			}
			err = json.Unmarshal(data, &newPosition)
			if err != nil {
				log.Println("Cant parse json init player data!")
			}
			playerList[currUser.id].PlayerPosition = newPosition
		case UpdatePitch:
			var pitch playerPitch
			data, err := packet.dataToBytes()
			if err != nil {
				log.Println("Cant turn inteface to []byte!")
				return
			}
			err = json.Unmarshal(data, &pitch)
			if err != nil {
				log.Println("Cant parse json init player data!")
			}
			playerList[currUser.id].Pitch = pitch.Pitch
		case UpdateRotation:
			var rotation playerRotation
			data, err := packet.dataToBytes()
			if err != nil {
				log.Println("Cant turn inteface to []byte!")
				return
			}
			err = json.Unmarshal(data, &rotation)
			if err != nil {
				log.Println("Cant parse json init player data!")
			}
			playerList[currUser.id].Rotation = rotation.Rotation
		default:
			sendErrorMsg(tcpConn, "Invalid packet type!")

		}

	}
}
