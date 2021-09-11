package globals

import (
	"aroundUsServer/player"
)

var PlayerList = make(map[int]*player.Player, 10)
var SpawnPosition = make([]player.PlayerPosition, 0)
var TasksDone = 0
var IsInLobby = true
var CurrId int
var Players int
var QueueReaders int
