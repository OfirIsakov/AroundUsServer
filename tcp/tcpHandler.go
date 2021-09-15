package tcp

import (
	"aroundUsServer/globals"
	"aroundUsServer/packet"
	"aroundUsServer/player"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

var initializePlayerLock sync.Mutex

func ListenTCP(host string, port int) {
	tcpListener, err := net.Listen("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		log.Panicln(err)
	}
	log.Println("Starting TCP listening")
	defer tcpListener.Close()

	for {
		tcpConnection, err := tcpListener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleTcpPlayer(tcpConnection)
	}
}

func handleTcpPlayer(conn net.Conn) {
	log.Println("Accepted new connection.")
	defer conn.Close()
	defer log.Println("Closed connection.")

	if !globals.IsInLobby {
		SendErrorMsg(conn, "Game has already started!")
		return
	}

	var currUser *player.Player

	for {
		// Max packet is 1024 bytes long
		buf := make([]byte, 1024)
		size, err := conn.Read(buf)
		if err != nil {
			SendErrorMsg(conn, "Error while reading the packet!\n"+err.Error())
			log.Println(string(buf))
			return
		}
		data := []byte(strings.TrimSpace(string(buf[:size])))

		log.Println(string(data))

		// Get the packet ID from the JSON
		var clientPacket packet.ClientPacket
		err = json.Unmarshal(data, &clientPacket)
		if err != nil {
			log.Println("Couldn't parse json player data! Skipping iteration!")
			continue
		}

		jsonString, err := json.Marshal(clientPacket.Data)
		if err != nil {
			SendErrorMsg(conn, "Cant turn inteface to json!\n"+err.Error())
			continue
		}

		packetData := []byte(jsonString)
		// packetData, err := clientPacket.DataToBytes()
		// if err != nil {
		// 	log.Println("Cant turn inteface to []byte!")
		// 	return
		// }
		// log.Println(string(packetData))

		switch clientPacket.Type {
		case packet.InitUser: // example: {"type":1, "data":{"name":"bro", "color": 1}}
			initializePlayerLock.Lock()

			currUser, err = initializePlayer(packetData, conn)
			if err != nil {
				SendErrorMsg(conn, "error while making a user: "+err.Error())
				return
			}

			initializePlayerLock.Unlock()

			globals.PlayerList[currUser.Id] = currUser

			// conenctedUsersJSON, err := json.Marshal(playerList) // Get all the players before adding the current user
			// if err != nil {
			// 	sendErrorMsg(conn, "Error while Marshaling the current connected users, disconnecting the user")
			// 	return
			// }
			// currUser.unduplicateUsername()

			// playerList[currUser.id] = currUser // Add the current user to the player map

			// defer currUser.removePlayer()

			// // Tell old users that a user connected
			// currUserJSON, err := json.Marshal(currUser) // Get all the players before adding the current user
			// if err != nil {
			// 	sendErrorMsg(conn, "Error while Marshaling the current user, other users dont know of your existance!")
			// }

			// currUserJSON, err = encapsulatePacketID(NewPlayerConnected, currUserJSON)
			// if err != nil {
			// 	log.Println("Didn't encapsulate currUserJSON with ID")
			// }
			// sendEveryoneTcpData([]byte(currUserJSON), []string{currUser.Name})

			// // Tell the current user where to spawn at
			// ClientSpawnPositionJSON, err := json.Marshal(currUser.PlayerPosition) // Get all the players before adding the current user
			// if err != nil {
			// 	sendErrorMsg(conn, "Error while Marshaling the current user position")
			// }
			// ClientSpawnPositionJSON, err = encapsulatePacketID(ClientSpawnPosition, ClientSpawnPositionJSON)
			// if err != nil {
			// 	log.Println("Didn't encapsulate currUserJSON with ID")
			// }
			// conn.Write([]byte(stampPacketLength(ClientSpawnPositionJSON)))

			// // Tell the current user who is already connected
			// conenctedUsersJSON, err = encapsulatePacketID(UsersInGame, conenctedUsersJSON)
			// if err != nil {
			// 	log.Println("Didn't encapsulate currUserJSON with ID")
			// }
			// conn.Write([]byte(stampPacketLength(conenctedUsersJSON)))

			// // Tell the user if he is manager
			// conenctedUsersJSON, err = encapsulatePacketID(IsUserManager, []byte(strconv.FormatBool(currUser.isManager)))
			// if err != nil {
			// 	log.Println("Didn't encapsulate currUserJSON with ID")
			// }
			// conn.Write([]byte(stampPacketLength(conenctedUsersJSON)))

			// // Tell the his ID
			// conenctedUsersJSON, err = encapsulatePacketID(UserId, []byte(strconv.FormatInt(int64(currUser.id), 10)))
			// if err != nil {
			// 	log.Println("Didn't encapsulate currUserJSON with ID")
			// }
			// conn.Write([]byte(stampPacketLength(conenctedUsersJSON)))

		// case StartGame:
		// 	var rotation playerRotation
		// 	data, err := packet.dataToBytes()
		// 	if err != nil {
		// 		log.Println("Cant turn inteface to []byte!")
		// 		return
		// 	}
		// 	err = json.Unmarshal(data, &rotation)
		// 	if err != nil {
		// 		log.Println("Cant parse json init player data!")
		// 	}
		// 	playerList[currUser.id].Rotation = rotation.Rotation
		default:
			SendErrorMsg(conn, "Invalid packet type!")

		}

	}
}

func initializePlayer(data []byte, tcpConnection net.Conn) (*player.Player, error) {
	var newPlayer *player.Player
	err := json.Unmarshal(data, &newPlayer)
	if err != nil {
		log.Println("Cant parse json init player data!")
		return nil, err
	}

	newPlayer.TcpConnection = tcpConnection // Set the player TCP connection socket

	// check if the name is taken or invalid
	// we need to keep a counter so the name will be in the format `<name> <count>`
	var newNameCount int8
	for nameOk := false; !nameOk; {
		nameOk = true
		for _, player := range globals.PlayerList {
			if player.Name == newPlayer.Name || player.Name == fmt.Sprintf("%s %d", newPlayer.Name, newNameCount) {
				newNameCount++
				nameOk = false
				break
			}
		}
	}
	if newNameCount > 0 {
		newPlayer.Name = fmt.Sprintf("%s %d", newPlayer.Name, newNameCount)
	}

	// check if the color is taken or invalid, if it is assign next not taken color
	if int8(0) > newPlayer.Color || int8(len(globals.Colors)) <= newPlayer.Color || globals.Colors[newPlayer.Color] {
		for index, color := range globals.Colors {
			if !color {
				newPlayer.Color = int8(index)
			}
		}
	}

	globals.Colors[newPlayer.Color] = true // set player color as taken

	// check if he is the first one in the lobby, if true set the player to be the game manager
	if len(globals.PlayerList) == 0 {
		newPlayer.IsManager = true
	}

	// set player ID and increase to next one, theoretically this can roll back at 2^31-1
	newPlayer.Id = globals.CurrId
	globals.CurrId++

	// set player spawn position
	newPlayer.PlayerPosition = globals.SpawnPositionsStack[len(globals.SpawnPositionsStack)-1]   // peek at the last element
	globals.SpawnPositionsStack = globals.SpawnPositionsStack[:len(globals.SpawnPositionsStack)] // pop

	log.Println("New player got generated: \n", newPlayer)

	return newPlayer, nil
}

func SendErrorMsg(conn net.Conn, msg string) error {
	log.Println(msg)
	errorJSON, err := json.Marshal(packet.ServerPacket{Type: packet.Error, Data: msg})
	if err != nil {
		return fmt.Errorf("error while marshaling error msg")
	}
	_, err = conn.Write(errorJSON)
	return err
}

// func sendEveryoneTcpData(data []byte, filter []string) {
// 	for _, client := range playerList {
// 		if !client.isInFilter(filter) {
// 			log.Println("Sending data to everyone(Filtered) " + string(data))
// 			client.tcpConnection.Write(stampPacketLength(data))
// 		}
// 	}
// }
