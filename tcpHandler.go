package main

import (
	"encoding/json"
	"log"
	"net"
	"strconv"
)

func handleTcpPlayer(conn net.Conn) {
	log.Println("Accepted new connection.")
	defer conn.Close()
	defer log.Println("Closed connection.")

	if !isInLobby {
		sendErrorMsg(conn, "Game has already started!")
	}

	var currUser player

	for {
		// Read the first 4 bytes and see the packet length
		buf := make([]byte, 4)
		size, err := conn.Read(buf)
		if err != nil {
			log.Println("Cant first read 4 bytes!")
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
		case InitUser: // {"ID":1, "Data":{"name":"bro"}}
			data, err := packet.dataToBytes()
			if err != nil {
				log.Println("Cant turn inteface to []byte!")
				return
			}
			currUser = genInitPlayerByData(data, conn)

			conenctedUsersJSON, err := json.Marshal(playerList) // Get all the players before adding the current user
			if err != nil {
				sendErrorMsg(conn, "Error while Marshaling the current connected users, disconnecting the user")
				return
			}
			currUser.unduplicateUsername()

			playerList = append(playerList, currUser) // Add the current user to the player list

			defer currUser.removePlayer()

			// Tell old users that a user connected
			currUserJSON, err := json.Marshal(currUser) // Get all the players before adding the current user
			if err != nil {
				sendErrorMsg(conn, "Error while Marshaling the current user, other users dont know of your existance!")
			}

			currUserJSON, err = encapsulatePacketID(NewPlayerConnected, currUserJSON)
			if err != nil {
				log.Println("Didn't encapsulate currUserJSON with ID")
			}
			sendEveryoneTcpData([]byte(currUserJSON), []string{currUser.Name})

			// Tell the current user where to spawn at
			ClientSpawnPositionJSON, err := json.Marshal(currUser.PlayerPosition) // Get all the players before adding the current user
			if err != nil {
				sendErrorMsg(conn, "Error while Marshaling the current user position")
			}
			ClientSpawnPositionJSON, err = encapsulatePacketID(ClientSpawnPosition, ClientSpawnPositionJSON)
			if err != nil {
				log.Println("Didn't encapsulate currUserJSON with ID")
			}
			conn.Write([]byte(stampPacketLength(ClientSpawnPositionJSON)))

			// Tell the current user who is already connected
			conenctedUsersJSON, err = encapsulatePacketID(UsersInGame, conenctedUsersJSON)
			if err != nil {
				log.Println("Didn't encapsulate currUserJSON with ID")
			}
			conn.Write([]byte(stampPacketLength(conenctedUsersJSON)))

			// Tell the user if he is manager
			conenctedUsersJSON, err = encapsulatePacketID(IsUserManager, []byte(strconv.FormatBool(currUser.isManager)))
			if err != nil {
				log.Println("Didn't encapsulate currUserJSON with ID")
			}
			conn.Write([]byte(stampPacketLength(conenctedUsersJSON)))

			// Tell the his ID
			conenctedUsersJSON, err = encapsulatePacketID(UserId, []byte(strconv.FormatInt(currUser.id, 10)))
			if err != nil {
				log.Println("Didn't encapsulate currUserJSON with ID")
			}
			conn.Write([]byte(stampPacketLength(conenctedUsersJSON)))

			log.Println("Started position update thread")
			go sendPlayerAllPositions(currUser.udpConnection, currUser.id)
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
		case StartGame:
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
			sendErrorMsg(conn, "Invalid packet type!")

		}

	}
}

func genInitPlayerByData(data []byte, tcpConnection net.Conn) player {
	var newPlayer player
	err := json.Unmarshal(data, &newPlayer)
	if err != nil {
		log.Println("Cant parse json init player data!")
	}

	newPlayer.Color = -1                    // Set the color to -1 as the player doesnt have a color yet
	newPlayer.tcpConnection = tcpConnection // Set the player TCP connection socket
	if len(playerList) == 0 {
		newPlayer.isManager = true // If he is the first one in the lobby, set the player to be the game manager
	}
	newPlayer.id = currId
	currId++
	newPlayer.PlayerPosition = position{0, 2, 0} // spawnPosition[newPlayer.index]

	log.Println("newPlayer got generated", newPlayer)

	return newPlayer
}

func sendEveryoneTcpData(data []byte, filter []string) {
	for _, client := range playerList {
		if !client.isInFilter(filter) {
			log.Println("Sending data to everyone(Filtered) " + string(data))
			client.tcpConnection.Write(stampPacketLength(data))
		}
	}
}

func sendErrorMsg(conn net.Conn, msg string) {
	log.Println(msg)
	errorJSON, err := json.Marshal(packetError{msg})
	if err != nil {
		log.Println("Error while Marshaling error msg!")
		return
	}
	conn.Write(stampPacketLength([]byte(errorJSON)))
}
