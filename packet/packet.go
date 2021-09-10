package packet

import helpers "aroundUsServer/utils"

// Client side packets
const (
	InitUser       = iota + 1 // TCP
	KilledPlayer              // TCP
	GameInit                  // TCP
	StartGame                 // TCP
	UpdatePos                 // UDP
	UpdatePitch               // UDP
	UpdateRotation            // UDP
	UdpDial                   // UDP
)

// Server side packets
const (
	UsersInGame         = iota + 1 // TCP
	IsUserManager                  // TCP
	NewPlayerConnected             // TCP
	ClientSpawnPosition            // TCP
	UserDisconnected               // TCP
	GameOver                       // TCP
	PlayerDied                     // TCP
	UserId                         // TCP
	PositionBroadcast              // UDP
)

type PacketType struct {
	ID   int
	Data interface{}
}

type PacketError struct {
	Msg string
}

type GameInitData struct {
	Imposters    []string
	TaskCount    uint8
	PlayerSpeed  uint8
	KillCooldown uint8
	Emergencies  uint8
}

func (dataPacket PacketType) DataToBytes() ([]byte, error) {
	buf, err := helpers.GetBytes(dataPacket.Data)
	return buf, err
}
