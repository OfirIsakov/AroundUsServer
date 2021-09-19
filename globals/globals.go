package globals

import (
	"aroundUsServer/player"
	"sync"
)

var PlayerList = make(map[int]*player.Player, 10)           // holds the players, maximum 10
var PlayerListLock sync.Mutex                               // used when updating/accessing the player list
var Colors [12]bool                                         // holds the colors, index indicated the color and the value if its taken or not
var SpawnPositionsStack = make([]player.PlayerPosition, 10) // holds where the players spawn when respawning after a meeting, functions as a stack
var TasksToWin = 0                                          // total tasks are needed to win
var TasksDone = 0                                           // how many tasks have been finished
var ImpostorsAlive = 0                                      // how many impostors are left
var CrewAlive = 0                                           // how many crew are left
var IsInLobby = true                                        // whether the game started or not
var CurrId int                                              // the next player id when joining
var QueueReaders int = 5                                    // amount of UDP queue reader threads

// var LobbyPositions = make([]player.PlayerPosition, 0) // holds where the players spawn when joining the lobby
