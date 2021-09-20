package packet

import (
	"encoding/json"
	"fmt"
	"net"
)

// Client -> Server packets
const (
	InitUser       = iota + 1 // TCP
	KilledPlayer              // TCP
	GameInit                  // TCP
	StartGame                 // TCP
	DialAddr                  // UDP
	UpdatePos                 // UDP
	UpdateRotation            // UDP
)

// Server -> Client packets
const (
	UsersInGame         = iota + 1 // TCP
	IsUserManager                  // TCP
	NewPlayerConnected             // TCP
	ClientSpawnPosition            // TCP
	UserDisconnected               // TCP
	GameOver                       // TCP
	PlayerDied                     // TCP
	UserId                         // TCP
	Error                          // TCP
	PositionBroadcast              // UDP
)

type ClientPacket struct {
	PlayerID int                    `json:"playerID"`
	Type     int8                   `json:"type"`
	Data     map[string]interface{} `json:"data"`
}

type ServerPacket struct {
	Type int8        `json:"type"`
	Data interface{} `json:"data"`
}

type GameInitData struct {
	Imposters    []string `json:"imposters"`
	TaskCount    uint8    `json:"taskCount"`
	PlayerSpeed  uint8    `json:"playerSpeed"`
	KillCooldown uint8    `json:"killCooldown"`
	Emergencies  uint8    `json:"emergencies"`
}

func (dataPacket *ClientPacket) DataToBytes() ([]byte, error) {
	// buf, err := helpers.GetBytes(dataPacket.Data)
	jsonString, err := json.Marshal(dataPacket.Data)
	if err != nil {
		return nil, fmt.Errorf("error while turning packet data to bytes")
	}
	return []byte(jsonString), nil
}

func StampPacket(data interface{}, packetType int8) ServerPacket {
	return ServerPacket{Type: packetType, Data: data}
}

func (packet *ServerPacket) SendTcpStream(tcpConnection net.Conn) (int, error) {
	packetJson, err := json.Marshal(*packet)
	if err != nil {
		return 0, fmt.Errorf("error while marshaling TCP packet")
	}
	return tcpConnection.Write([]byte(packetJson))
}

func (packet *ServerPacket) SendUdpStream(udpConnection *net.UDPConn, udpAddress *net.UDPAddr) (int, error) {
	packetJson, err := json.Marshal(*packet)
	if err != nil {
		return 0, fmt.Errorf("error while marshaling UDP packet")
	}
	return udpConnection.WriteToUDP([]byte(packetJson), udpAddress)
}
