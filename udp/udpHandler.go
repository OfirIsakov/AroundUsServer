package udp

import (
	"aroundUsServer/globals"
	helpers "aroundUsServer/utils"
	"encoding/json"
	"log"
	"net"

	"github.com/enriquebris/goconcurrentqueue"
)

type PacketData struct {
	ID   int
	Data string
}

var queue *goconcurrentqueue.FIFO

func ListenUDP(host string, port int) {
	log.Println("Starting UDP listening")

	queue = goconcurrentqueue.NewFIFO()

	//Basic variables
	adreesss := "127.0.0.1:8080"
	protocol := "udp"

	//Build the address
	udpAddr, err := net.ResolveUDPAddr(protocol, adreesss)
	if err != nil {
		log.Println("Wrong Address")
		return
	}

	//Create the connection
	udpConn, err := net.ListenUDP(protocol, udpAddr)
	if err != nil {
		log.Println(err)
	}

	// create queue readers
	for i := 0; i < globals.QueueReaders; i++ {
		go handleIncomingUdpData()
	}

	//Keep calling this function
	for {
		quit := make(chan struct{})
		for i := 0; i < 1; i++ {
			go getIncomingUdp(udpConn, quit)
		}
		<-quit // hang until an error
	}
}

func getIncomingUdp(conn *net.UDPConn, quit chan struct{}) {
	var buf []byte
	err := error(nil)

	for err == nil {
		buf = make([]byte, 1024)

		size, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Println("Cant read packet!", err)
			continue
		}
		data := buf[:size]

		queue.Enqueue(data)
	}

	log.Println("Listener failed - restarting!", err)
	quit <- struct{}{}
}

func handleIncomingUdpData() {
	var err error = nil

	for err == nil {
		data, err := queue.DequeueOrWaitForNextElement()
		if err != nil {
			log.Println("Couldn't dequeue!")
			continue
		}

		bytesData, err := helpers.GetBytes(data)
		if err != nil {
			log.Println("Couldn't turn data to byte array!")
			continue
		}

		// Get the packet ID from the JSON
		var packet PacketData
		err = json.Unmarshal(bytesData, &packet)
		if err != nil {
			log.Println("Couldn't parse json player data! Skipping iteration!")
			continue
		}

		log.Println("from", packet.ID, ":", packet.Data)
	}
}

// func handleUdpData(conn net.UDPConn) {
// 	log.Println("Accepted new connection.")
// 	defer conn.Close()
// 	defer log.Println("Closed connection.")

// 	for {
// 		// Read the first 4 bytes and see the packet length
// 		buf := make([]byte, 4)
// 		size, err := conn.Read(buf)
// 		if err != nil {
// 			log.Println("Cant read first 4 bytes!")
// 			err = nil
// 			return
// 		}
// 		data := buf[:size]
// 		readLength, err := strconv.Atoi(string(data))
// 		if err != nil {
// 			log.Println("Cant convent string size to int!")
// 			return
// 		}

// 		buf = make([]byte, readLength)
// 		size, err = conn.Read(buf)
// 		if err != nil {
// 			log.Println("Cant read the rest of the packet!")
// 			return
// 		}
// 		data = buf[:size]

// 		// Get the packet ID from the JSON
// 		var packet packetType
// 		err = json.Unmarshal(data, &packet)
// 		if err != nil {
// 			log.Println("Couldn't parse json player data! Skipping iteration!")
// 			continue
// 		}

// 		switch packet.ID {
// 		case UpdatePos:
// 			var newPosition position
// 			data, err := packet.dataToBytes()
// 			if err != nil {
// 				log.Println("Cant turn inteface to []byte!")
// 				return
// 			}
// 			err = json.Unmarshal(data, &newPosition)
// 			if err != nil {
// 				log.Println("Cant parse json init player data!")
// 			}
// 			playerList[currUser.id].PlayerPosition = newPosition
// 		case UpdatePitch:
// 			var pitch playerPitch
// 			data, err := packet.dataToBytes()
// 			if err != nil {
// 				log.Println("Cant turn inteface to []byte!")
// 				return
// 			}
// 			err = json.Unmarshal(data, &pitch)
// 			if err != nil {
// 				log.Println("Cant parse json init player data!")
// 			}
// 			playerList[currUser.id].Pitch = pitch.Pitch
// 		case UpdateRotation:
// 			var rotation playerRotation
// 			data, err := packet.dataToBytes()
// 			if err != nil {
// 				log.Println("Cant turn inteface to []byte!")
// 				return
// 			}
// 			err = json.Unmarshal(data, &rotation)
// 			if err != nil {
// 				log.Println("Cant parse json init player data!")
// 			}
// 			playerList[currUser.id].Rotation = rotation.Rotation
// 		default:
// 			sendErrorMsg(tcpConn, "Invalid packet type!")

// 		}

// 	}
// }
