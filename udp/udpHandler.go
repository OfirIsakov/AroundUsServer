package udp

import (
	"aroundUsServer/globals"
	"aroundUsServer/packet"
	"aroundUsServer/player"
	"aroundUsServer/tcp"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/enriquebris/goconcurrentqueue"
)

var packetsQueue *goconcurrentqueue.FIFO

type udpPacket struct {
	Address *net.UDPAddr
	Data    []byte
}

func ListenUDP(host string, port int) {
	log.Println("Starting UDP listening")

	packetsQueue = goconcurrentqueue.NewFIFO()

	//Basic variables
	addresss := fmt.Sprintf("%s:%d", host, port)
	protocol := "udp"

	//Build the address
	udpAddr, err := net.ResolveUDPAddr(protocol, addresss)
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
	err := error(nil)

	for err == nil {
		buffer := make([]byte, 1024)

		size, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Println("Cant read packet!", err)
			continue
		}
		data := buffer[:size]

		packetsQueue.Enqueue(udpPacket{Address: addr, Data: data})
	}

	log.Println("Listener failed - restarting!", err)
	quit <- struct{}{}
}

func handleIncomingUdpData() {
	for {
		dequeuedRawPacket, err := packetsQueue.DequeueOrWaitForNextElement()
		if err != nil {
			log.Println("Couldn't dequeue!")
			continue
		}

		dequeuedPacket, ok := dequeuedRawPacket.(udpPacket)
		if !ok {
			log.Println("Couldn't turn udp data to udpPacket!")
			continue
		}

		// Get the packet ID from the JSON
		var dataPacket packet.ClientPacket
		err = json.Unmarshal(dequeuedPacket.Data, &dataPacket)
		if err != nil {
			log.Println("Couldn't parse json player data! Skipping iteration!")
			continue
		}

		log.Println("from", dataPacket.PlayerID, ":", dataPacket.Data)
		err = handleUdpData(dequeuedPacket.Address, dataPacket)
		if err != nil {
			log.Println("Error while handling UDP packet: " + err.Error())
			continue
		}
	}
}

func handleUdpData(userAddress *net.UDPAddr, clientPacket packet.ClientPacket) error {
	dataPacket, err := clientPacket.DataToBytes()
	if err != nil {
		return err
	}
	switch clientPacket.Type {
	case packet.DialAddr:
		//TODO here we want to save the player addr
		return nil
	case packet.UpdatePos:
		var newPosition player.PlayerPosition
		err := json.Unmarshal([]byte(dataPacket), &newPosition)
		if err != nil {
			return fmt.Errorf("cant parse position player data")
		}
		globals.PlayerList[clientPacket.PlayerID].PlayerPosition = newPosition
	case packet.UpdateRotation:
		var newRotation player.PlayerRotation
		err := json.Unmarshal([]byte(dataPacket), &newRotation)
		if err != nil {
			return fmt.Errorf("cant parse rotation player data")
		}
		globals.PlayerList[clientPacket.PlayerID].Rotation = newRotation
	default:
		tcp.SendErrorMsg(globals.PlayerList[clientPacket.PlayerID].TcpConnection, "Invalid UDP packet type!")

	}

	return nil
}
