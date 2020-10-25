package main

import (
	"log"
	"net"
	"strconv"
)

const port = 27403
const host = "127.0.0.1"

type position struct {
	x float32
	y float32
	z float32
}

type player struct {
	name           string   // The name of the player, can contain anything
	color          int8     // The index of the color in the color list held in the client
	isImposter     bool     // Sent on the round start to tell the client if hes an imposter or crew
	inVent         bool     // If true the server shouldnt send the player locations until hes leaving the vent
	isDead         bool     // If the player is dead the server shouldnt send his location
	gotReported    bool     // If the player didnt get reported yet tell the client to show a body on the ground
	playerPosition position // The position of the player in Unity world cordinates
	pitch          int      // Should be -90 <= pitch <= 90, represents the head pitch(Up and down)
	rotation       int      // Should be 0 <= rotation <= 360, represents the body rotation
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
			log.Panicln(err)
		}

		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	log.Println("Accepted new connection.")
	defer conn.Close()
	defer log.Println("Closed connection.")

	for {
		buf := make([]byte, 1024)
		size, err := conn.Read(buf)
		if err != nil {
			return
		}
		data := buf[:size]
		log.Println("Read new data from connection", data)
		conn.Write(data)
	}
}
