package packet

import helpers "aroundUsServer/utils"

// Client -> Server packets
const (
	InitUser       = iota + 1 // TCP
	KilledPlayer              // TCP
	GameInit                  // TCP
	StartGame                 // TCP
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
	PlayerID int         `json:"playerID"`
	Type     int8        `json:"type"`
	Data     interface{} `json:"data"`
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
	buf, err := helpers.GetBytes(dataPacket.Data)
	return buf, err
}
