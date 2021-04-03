package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

/*
** DISCLAIMER! **
This server is not designed to check the users inputs!
This server is quick and dirty to be able to play with friends a game and most calculations
get calculated in the client so the server is highly trusting the clients.
Its not designed to be released to the wild and shouldn't be trusted with random users.
If you use this server in the wild cheating will be SO ez.
*/

// Constants
// const host = "10.0.0.4" // "127.0.0.1"
// const port = 27403

// Client side packets
const (
	InitUser = iota + 1
	UpdatePos
	UpdatePitch
	UpdateRotation
	KilledPlayer
	GameInit
	StartGame
)

// Server sside packets
const (
	UsersInGame = iota + 1
	IsUserManager
	NewPlayerConnected
	PositionBroadcast
	ClientSpawnPosition
	UserDisconnected
	GameOver
	PlayerDied
)

// Global vars
var playerList = make([]player, 0)
var spawnPosition = make([]position, 0)
var tasksDone = 0
var isInLobby = true
var removeLock sync.Mutex

// Structs
type packetType struct {
	ID   int
	Data interface{}
}

type packetError struct {
	Msg string
}

type position struct {
	X float32
	Y float32
	Z float32
}

type playerPitch struct {
	Pitch float32
}

type gameInitData struct {
	imposters    []string
	taskCount    uint8
	speed        uint8
	killCooldown uint8
	emergencies  uint8
}

type playerRotation struct {
	Rotation float32
}

type player struct {
	Name           string   // The name of the player, can contain anything
	Color          int8     // The index of the color in the color list held in the client
	isManager      bool     // Whether the player is the game manager or not, he can start the game
	IsImposter     bool     // Sent on the round start to tell the client if hes an imposter or crew
	InVent         bool     // If true the server shouldnt send the player locations until hes leaving the vent
	IsDead         bool     // If the player is dead the server shouldnt send his location
	GotReported    bool     // If the player didnt get reported yet tell the client to show a body on the ground
	PlayerPosition position // The position of the player in Unity world cordinates
	Pitch          float32  // Should be -90 <= pitch <= 90, represents the head pitch(Up and down)
	Rotation       float32  // Should be 0 <= rotation <= 360, represents the body rotation
	index          int      // The current index of the player in the slice
	connection     net.Conn // The player connection socket
}

func main() {
	var host = flag.String("ip", "127.0.0.1", "Server local IP")
	var port = flag.Int("port", 27403, "Server port")
	flag.Parse()
	l, err := net.Listen("tcp", *host+":"+strconv.Itoa(*port))
	if err != nil {
		log.Panicln(err)
	}
	log.Println("Listening to connections at " + l.Addr().String())
	defer l.Close()

	initSpawnPosition()
	go consoleCommands()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handlePlayer(conn)
	}
}

func (packet packetType) dataToBytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	err := enc.Encode(packet.Data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p player) isInFilter(filter []string) bool {
	for _, name := range filter {
		if name == p.Name {
			return true
		}
	}
	return false
}

func (p player) unduplicateUsername() {
	var nextNumber int8
	for i := 0; i < len(playerList); i++ {
		if playerList[i].Name == p.Name {
			i = 0
			nextNumber++
		}
	}
	if nextNumber != 0 {
		p.Name = p.Name + strconv.Itoa(int(nextNumber))
	}
}

func (p player) removePlayer() {
	for {
		if p.index+1 <= len(playerList) {
			break
		}
	}
	removeLock.Lock()
	playerList = append(playerList[:p.index], playerList[p.index+1:]...)
	for i := 0; i < len(playerList); i++ {
		playerList[i].index = i // Update all players index
	}
	removeLock.Unlock()

	currUserJSON, err := json.Marshal(p) // Get all the players before adding the current user
	if err != nil {
		sendErrorMsg(p.connection, "Error while Marshaling the user for remove, brotha tell Ofir!")
		return
	}

	currUserJSON, err = encapsulatePacketID(UserDisconnected, currUserJSON)
	if err != nil {
		log.Println("Didn't encapsulate currUserJSON with ID")
		return
	}
	sendEveryoneData([]byte(currUserJSON), []string{p.Name})
}

func consoleCommands() {
	for {
		var command string
		fmt.Scanln(&command)
		commands := strings.Split(command, " ")
		switch {
		case commands[0] == "help":
			log.Println("help")
			log.Println("list")
			log.Println("dc [number]")
		case commands[0] == "list" || commands[0] == "ls":
			for i, client := range playerList {
				log.Println(fmt.Sprintf("%v) %v", i, client))
			}
		case commands[0] == "dc":
			position, err := strconv.Atoi(commands[1])
			if err != nil {
				log.Println("Cant convert to number position")
			}
			playerList[position].connection.Close()
		default:
			log.Println("Unknown command")
		}
	}
}

func handlePlayer(conn net.Conn) {
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
			sendEveryoneData([]byte(currUserJSON), []string{currUser.Name})

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

			log.Println("Started position update thread")
			go sendPlayerAllPositions(conn, currUser.index)
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
			playerList[currUser.index].PlayerPosition = newPosition
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
			playerList[currUser.index].Pitch = pitch.Pitch
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
			playerList[currUser.index].Rotation = rotation.Rotation
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
			playerList[currUser.index].Rotation = rotation.Rotation
		default:
			sendErrorMsg(conn, "Invalid packet type!")

		}

	}
}

func initSpawnPosition() {
	for i := 0; i < 6; i++ {
		spawnPosition = append(spawnPosition, position{-4, 1.75, float32(14 - i)})
	}
	for i := 0; i < 6; i++ {
		spawnPosition = append(spawnPosition, position{-6, 1.75, float32(14 - i)})
	}
}

func sendPlayerAllPositions(conn net.Conn, playerIndex int) {
	for {
		if playerIndex >= 0 && playerIndex <= len(playerList) && playerIndex+1 <= len(playerList) {
			playersToSend := make([]player, len(playerList))
			idexesCopied := copy(playersToSend, playerList)
			if idexesCopied > 0 {
				playersToSend = append(playersToSend[:playerIndex], playersToSend[+1:]...)
				clientJSON, err := json.Marshal(playersToSend)
				if err != nil {
					log.Println("Cant marshal location")
				}
				clientJSON, err = encapsulatePacketID(PositionBroadcast, clientJSON)
				if err != nil {
					log.Println("Didn't encapsulate currUserJSON with ID")
				}
				conn.Write(stampPacketLength(clientJSON))
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func encapsulatePacketID(ID int, JSON []byte) ([]byte, error) {
	errorJSON, err := json.Marshal(packetType{ID, JSON})
	return errorJSON, err
}

func stampPacketLength(data []byte) []byte {
	packet := make([]byte, 0, 4+len(data))
	packet = append(packet, []byte(fmt.Sprintf("%04d", len(data)))...)
	packet = append(packet, data...)
	return packet
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

func sendEveryoneData(data []byte, filter []string) {
	for _, client := range playerList {
		if !client.isInFilter(filter) {
			log.Println("Sending data to everyone(Filtered) " + string(data))
			client.connection.Write(stampPacketLength(data))
		}
	}
}

func genInitPlayerByData(data []byte, conn net.Conn) player {
	var newPlayer player
	err := json.Unmarshal(data, &newPlayer)
	if err != nil {
		log.Println("Cant parse json init player data!")
	}

	newPlayer.Color = -1        // Set the color to -1 as the player doesnt have a color yet
	newPlayer.connection = conn // Set the player connection socket
	if len(playerList) == 0 {
		newPlayer.isManager = true // If he is the first one in the lobby, set the player to be the game manager
	}
	newPlayer.index = len(playerList)
	newPlayer.PlayerPosition = position{0, 2, 0} // spawnPosition[newPlayer.index]

	log.Println("newPlayer got generated", newPlayer)

	return newPlayer
}
