package main

import (
	"encoding/json"
	"log"
	"net"
	"strconv"
)

/*
** DISCLAIMER! **
This server is not designed to check the users inputs!
This server is quick and dirty to be able to play with friends a game and most calculations
get calculated in the client so the server is highly trusting the clients.
Its not designed to be released to the wild and shouldn't be trusted with random users.
*/

// Constants
const host = "127.0.0.1"
const port = 27403

// Packets type enum
const (
	InitName = iota + 1
	NewPlayerConnected
)

// Global vars
var playerList = make([]player, 0)

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

type player struct {
	Name           string   // The name of the player, can contain anything
	Color          int8     // The index of the color in the color list held in the client
	IsImposter     bool     // Sent on the round start to tell the client if hes an imposter or crew
	InVent         bool     // If true the server shouldnt send the player locations until hes leaving the vent
	IsDead         bool     // If the player is dead the server shouldnt send his location
	GotReported    bool     // If the player didnt get reported yet tell the client to show a body on the ground
	PlayerPosition position // The position of the player in Unity world cordinates
	Pitch          int      // Should be -90 <= pitch <= 90, represents the head pitch(Up and down)
	Rotation       int      // Should be 0 <= rotation <= 360, represents the body rotation
	isManager      bool     // Wheather the player is the game manager or not, he can start the game
	index          int      // The current index of the player in the slice
	connection     net.Conn // The player connection socket
}

func main() {
	l, err := net.Listen("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		log.Panicln(err)
	}
	log.Println("Listening to connections at '"+host+"' on port", strconv.Itoa(port))
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handlePlayer(conn)
	}
}

func (t player) isInFilter(filter []string) bool {
	for _, name := range filter {
		if name == t.Name {
			return true
		}
	}
	return false
}

func (t player) removePlayerByName() {
	// var index int
	// for _, client := range playerList {
	// 	if client.Name == t.Name {
	// 		index = client.index
	// 	}
	// }
	playerList = append(playerList[:t.index], playerList[t.index+1:]...)
	for i := 0; i < len(playerList); i++ {
		playerList[i].index = i // Update all players index
	}
}

func handlePlayer(conn net.Conn) {
	log.Println("Accepted new connection.")
	defer conn.Close()
	defer log.Println("Closed connection.")

	var currUser player

	for {
		// Get 1KB data, the client shouldnt send more then that to the server
		buf := make([]byte, 1024)
		size, err := conn.Read(buf)
		if err != nil {
			return
		}
		data := buf[:size]

		// Get the packet ID from the JSON
		var packet packetType
		err = json.Unmarshal(data, &packet)
		if err != nil {
			log.Println("Couldnt parse json player data! Skipping iteration!")
			continue
		}

		switch packet.ID {
		case InitName: // {"ID":1, "name":"bro"}
			currUser = genInitPlayerByData(data, conn)

			conenctedUsersJSON, err := json.Marshal(playerList) // Get all the players before adding the current user
			if err != nil {
				sendErrorMsg(conn, "Error while Marshaling the current connected users, disconnecting the user")
				return
			}

			playerList = append(playerList, currUser) // Add the current user to the player list

			defer currUser.removePlayerByName()

			// Tell old users that a user connected
			currUserJSON, err := json.Marshal(currUser) // Get all the players before adding the current user
			if err != nil {
				sendErrorMsg(conn, "Error while Marshaling the current user, other users dont know of your existance!")
			}

			sendEveryoneData([]byte(currUserJSON), []string{currUser.Name})

			// Tell the current user who is already connected
			log.Println(string(conenctedUsersJSON))
			conn.Write(conenctedUsersJSON)

			log.Println(playerList)
		default:
			conn.Write([]byte("Invalid packet type!"))
		}

	}
}

func sendErrorMsg(conn net.Conn, msg string) {
	log.Println(msg)
	errorJSON, err := json.Marshal(packetError{msg + " Bruh tell the developer about this..."})
	if err != nil {
		log.Println("Error while Marshaling error msg!")
		return
	}
	conn.Write([]byte(errorJSON))
}

func sendEveryoneData(data []byte, filter []string) {
	for _, client := range playerList {
		if !client.isInFilter(filter) {
			log.Println("Sending data to everyone(Filtered) " + string(data))
			client.connection.Write(data)
		}
	}
}

func genInitPlayerByData(data []byte, conn net.Conn) player {
	var newPlayer player
	err := json.Unmarshal(data, &newPlayer)
	if err != nil {
		log.Println("Couldnt parse json player data!")
	}

	newPlayer.Color = -1        // Set the color to -1 as the player doesnt have a color yet
	newPlayer.connection = conn // Set the player connection socket
	if len(playerList) == 0 {
		newPlayer.isManager = true // If he is the first one in the lobby, set the player to be the gae manager
	}

	log.Println("newPlayer got generated", newPlayer)

	return newPlayer
}
