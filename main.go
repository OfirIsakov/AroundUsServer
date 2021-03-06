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
)

/*
** DISCLAIMER! **
This server is not designed to check the users inputs!
This server is quick and dirty to be able to play with friends a game and most calculations
get calculated in the client so the server is highly trusting the clients.
Its not designed to be released to the wild and shouldn't be trusted with random users.
If you use this server in the wild cheating will be SO ez.
The unity game client I built wont be released as I respect the developers of Among Us.
*/

// Constants
// const host = "10.0.0.4" // "127.0.0.1"
// const port = 27403

// Client side packets
const (
	InitUser       = iota + 1 // TCP
	UpdatePos                 // UDP
	UpdatePitch               // UDP
	UpdateRotation            // UDP
	KilledPlayer              // TCP
	GameInit                  // TCP
	StartGame                 // TCP
)

// Server sside packets
const (
	UsersInGame         = iota + 1 // TCP
	IsUserManager                  // TCP
	NewPlayerConnected             // TCP
	PositionBroadcast              // UDP
	ClientSpawnPosition            // TCP
	UserDisconnected               // TCP
	GameOver                       // TCP
	PlayerDied                     // TCP
	UserId                         // TCP
)

// Global vars
var playerList = make([]player, 0)
var spawnPosition = make([]position, 0)
var tasksDone = 0
var isInLobby = true
var currId int64
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
	Name           string      // The name of the player, can contain anything
	Color          int8        // The index of the color in the color list held in the client
	isManager      bool        // Whether the player is the game manager or not, he can start the game
	IsImposter     bool        // Sent on the round start to tell the client if hes an imposter or crew
	InVent         bool        // If true the server shouldnt send the player locations until hes leaving the vent
	IsDead         bool        // If the player is dead the server shouldnt send his location
	GotReported    bool        // If the player didnt get reported yet tell the client to show a body on the ground
	PlayerPosition position    // The position of the player in Unity world cordinates
	Pitch          float32     // Should be -90 <= pitch <= 90, represents the head pitch(Up and down)
	Rotation       float32     // Should be 0 <= rotation <= 360, represents the body rotation
	id             int64       // Id of the player
	tcpConnection  net.Conn    // The player TCP connection socket
	udpConnection  net.UDPConn // The player UDP connection socket
}

func main() {
	var host = flag.String("ip", "127.0.0.1", "Server local IP")
	var port = flag.Int("port", 27403, "Server port")
	flag.Parse()
	tcpListener, err := net.Listen("tcp", *host+":"+strconv.Itoa(*port))
	if err != nil {
		log.Panicln(err)
	}
	log.Println("Listening to connections at " + tcpListener.Addr().String())
	defer tcpListener.Close()

	initSpawnPosition()
	go consoleCommands()
	for {
		tcpConnection, err := tcpListener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleTcpPlayer(tcpConnection)
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
	for i := 0; i < len(playerList); i++ {
		if playerList[i].id == p.id {

			removeLock.Lock()
			playerList = append(playerList[:i], playerList[i:]...)
			removeLock.Unlock()

			currUserJSON, err := json.Marshal(p) // Get all the players before adding the current user
			if err != nil {
				sendErrorMsg(p.tcpConnection, "Error while Marshaling the user for remove, brotha tell Ofir!")
				return
			}

			currUserJSON, err = encapsulatePacketID(UserDisconnected, currUserJSON)
			if err != nil {
				log.Println("Didn't encapsulate currUserJSON with ID")
				return
			}
			sendEveryoneTcpData([]byte(currUserJSON), []string{p.Name})
		}
	}
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
			playerList[position].tcpConnection.Close()
		default:
			log.Println("Unknown command")
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
