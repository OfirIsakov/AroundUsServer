package player

import "net"

type Player struct {
	Name           string         // The name of the player, can contain anything
	Color          int8           // The index of the color in the color list held in the client
	isManager      bool           // Whether the player is the game manager or not, he can start the game
	IsImposter     bool           // Sent on the round start to tell the client if hes an imposter or crew
	InVent         bool           // If true the server shouldnt send the player locations until hes leaving the vent
	IsDead         bool           // If the player is dead the server shouldnt send his location
	GotReported    bool           // If the player didnt get reported yet tell the client to show a body on the ground
	PlayerPosition PlayerPosition // The position of the player in Unity world cordinates
	Pitch          float32        // Should be -90 <= pitch <= 90, represents the head pitch(Up and down)
	Rotation       float32        // Should be 0 <= rotation <= 360, represents the body rotation
	id             int            // Id of the player
	tcpConnection  net.Conn       // The player TCP connection socket
	udpAddress     *net.UDPAddr   // The player UDP address socket
}

type PlayerPosition struct {
	X float32
	Y float32
	Z float32
}
