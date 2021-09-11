package player

import "net"

type Player struct {
	Name           string         `json:"name"`           // The name of the player, can contain anything
	Color          int8           `json:"color"`          // The index of the color in the color list held in the client
	IsManager      bool           `json:"-"`              // Whether the player is the game manager or not, he can start the game
	IsImposter     bool           `json:"isImposter"`     // Sent on the round start to tell the client if hes an imposter or crew
	InVent         bool           `json:"inVent"`         // If true the server shouldnt send the player locations until hes leaving the vent
	IsDead         bool           `json:"isDead"`         // If the player is dead the server shouldnt send his location
	GotReported    bool           `json:"gotReported"`    // If the player didnt get reported yet tell the client to show a body on the ground
	PlayerPosition PlayerPosition `json:"playerPosition"` // The position of the player in Unity world cordinates
	Rotation       PlayerRotation `json:"rotation"`       // Pitch: -90 <= pitch <= 90(head up and down), Yaw: 0 <= rotation <= 360(body rotation)
	Id             int            `json:"-"`              // Id of the player
	TcpConnection  net.Conn       `json:"-"`              // The player TCP connection socket
	UdpAddress     *net.UDPAddr   `json:"-"`              // The player UDP address socket
}

type PlayerPosition struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}

type PlayerRotation struct {
	Pitch float32 `json:"rotation"`
	Yaw   float32 `json:"yaw"`
}
