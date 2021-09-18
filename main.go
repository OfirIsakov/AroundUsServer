package main

import (
	"aroundUsServer/cli"
	"aroundUsServer/globals"
	"aroundUsServer/player"
	"aroundUsServer/tcp"
	"aroundUsServer/udp"
	"flag"
	"log"
)

/*
** DISCLAIMER! **
This server is not designed to check the users inputs!
This server is ~quick~ and dirty to be able to play with friends a game and most calculations
get calculated in the client so the server is highly trusting the clients.
Its not designed to be released to the wild and shouldn't be trusted with random users.
If you use this server in the wild, cheating & crashing will be SO ez.
The unity game client I built wont be released as I respect the developers of "Among Us".
*/

func main() {
	// init variables
	initSpawnPosition()

	// get program flags
	var host = flag.String("ip", "127.0.0.1", "Server listen IP")
	var port = flag.Int("port", 27403, "Server listen port")
	flag.Parse()

	log.Printf("Starting listening on: %s:%d", *host, *port)

	// start listening
	go tcp.ListenTCP(*host, *port)
	go udp.ListenUDP(*host, *port)

	// block main thread with the console
	cli.ConsoleCLI()
}

// func (p player.Player) isInFilter(filter []string) bool {
// 	for _, name := range filter {
// 		if name == p.Name {
// 			return true
// 		}
// 	}
// 	return false
// }

// func (p player.Player) unduplicateUsername() {
// 	var nextNumber int8
// 	wasDuped := true
// 	criticalUseLock.Lock()
// 	for wasDuped {
// 		wasDuped = false
// 		for _, player := range playerList {
// 			if player.Name == p.Name {
// 				nextNumber++
// 				wasDuped = true
// 				break
// 			}
// 		}
// 	}
// 	criticalUseLock.Unlock()
// 	if nextNumber != 0 {
// 		p.Name = p.Name + strconv.Itoa(int(nextNumber))
// 	}
// }

// func (p player) removePlayer() {
// 	criticalUseLock.Lock()
// 	delete(playerList, p.id)
// 	criticalUseLock.Unlock()

// 	currUserJSON, err := json.Marshal(p) // Get all the players before adding the current user
// 	if err != nil {
// 		sendErrorMsg(p.tcpConnection, "Error while Marshaling the user for remove, brotha tell ofido!")
// 		return
// 	}

// 	currUserJSON, err = encapsulatePacketID(UserDisconnected, currUserJSON)
// 	if err != nil {
// 		log.Println("Didn't encapsulate currUserJSON with ID")
// 		return
// 	}
// 	sendEveryoneTcpData([]byte(currUserJSON), []string{p.Name})
// }

func initSpawnPosition() {
	for i := 5; i <= 0; i++ {
		globals.SpawnPositionsStack = append(globals.SpawnPositionsStack, player.PlayerPosition{X: -4, Y: 1.75, Z: float32(14 - i)})
		globals.SpawnPositionsStack = append(globals.SpawnPositionsStack, player.PlayerPosition{X: -6, Y: 1.75, Z: float32(14 - i)})
	}
}

// func encapsulatePacketID(ID int, JSON []byte) ([]byte, error) {
// 	errorJSON, err := json.Marshal(packetType{ID, JSON})
// 	return errorJSON, err
// }

// func stampPacketLength(data []byte) []byte {
// 	packet := make([]byte, 0, 4+len(data))
// 	packet = append(packet, []byte(fmt.Sprintf("%04d", len(data)))...)
// 	packet = append(packet, data...)
// 	return packet
// }
