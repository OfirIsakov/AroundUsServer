package udp

import (
	"aroundUsServer/globals"
	"aroundUsServer/packet"
	"aroundUsServer/player"
	helpers "aroundUsServer/utils"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/enriquebris/goconcurrentqueue"
)

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
	var buffer []byte
	err := error(nil)

	for err == nil {
		buffer = make([]byte, 1024)

		size, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Println("Cant read packet!", err)
			continue
		}
		data := buffer[:size]

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
		var dataPacket packet.PacketType
		err = json.Unmarshal(bytesData, &dataPacket)
		if err != nil {
			log.Println("Couldn't parse json player data! Skipping iteration!")
			continue
		}

		log.Println("from", dataPacket.ID, ":", dataPacket.Data)
		err = handleUdpData(dataPacket)
		if err != nil {
			log.Println("Error while handling UDP packet: " + err.Error())
			continue
		}
	}
}

func handleUdpData(dataPacket packet.PacketType) error {
	data, err := dataPacket.DataToBytes()
	if err != nil {
		return fmt.Errorf("cant turn inteface to []byte")
	}

	switch dataPacket.Type {
	case packet.UpdatePos:
		var newPosition player.PlayerPosition
		err = json.Unmarshal(data, &newPosition)
		if err != nil {
			return fmt.Errorf("cant parse position player data")
		}
		globals.PlayerList[dataPacket.ID].PlayerPosition = newPosition
	case packet.UpdateRotation:
		var newRotation player.PlayerRotation
		err = json.Unmarshal(data, &newRotation)
		if err != nil {
			return fmt.Errorf("cant parse rotation player data")
		}
		globals.PlayerList[dataPacket.ID].Rotation = newRotation
	default:
		// sendErrorMsg(tcpConn, "Invalid packet type!")

	}

	return nil
}
