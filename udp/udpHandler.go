package udp

import (
	"aroundUsServer/globals"
	"aroundUsServer/packet"
	"aroundUsServer/player"
	"aroundUsServer/tcp"
	"aroundUsServer/utils"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/enriquebris/goconcurrentqueue"
)

var packetsQueue *goconcurrentqueue.FIFO
var udpConnection *net.UDPConn

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
	udpConnection, err = net.ListenUDP(protocol, udpAddr)
	if err != nil {
		log.Println(err)
	}

	// create queue readers
	for i := 0; i < globals.QueueReaders; i++ {
		go handleIncomingUdpData()
	}

	// reate position updater
	go updatePlayerPosition()

	//Keep calling this function
	for {
		quit := make(chan struct{})
		for i := 0; i < 1; i++ {
			go getIncomingUdp(quit)
		}
		<-quit // hang until an error
	}
}

func getIncomingUdp(quit chan struct{}) {
	err := error(nil)

	for err == nil {
		buffer := make([]byte, 1024)

		size, addr, err := udpConnection.ReadFromUDP(buffer)
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

		var dataPacket packet.ClientPacket
		err = json.Unmarshal(dequeuedPacket.Data, &dataPacket)
		if err != nil {
			log.Println("Couldn't parse json player data! Skipping iteration!")
			continue
		}

		err = handleUdpData(dequeuedPacket.Address, dataPacket)
		if err != nil {
			log.Println("Error while handling UDP packet: " + err.Error())
			continue
		}
	}
}

func handleUdpData(userAddress *net.UDPAddr, clientPacket packet.ClientPacket) error {
	if clientPacket.Type == packet.DialAddr { // {"type":5, "id": 0}
		if user, ok := globals.PlayerList[clientPacket.PlayerID]; ok {
			user.UdpAddress = userAddress
		}
		return nil
	}

	dataPacket, err := clientPacket.DataToBytes()
	if err != nil {
		return err
	}

	switch clientPacket.Type {
	case packet.UpdatePos: // {"type":6, "id": 0, "data":{"x":0, "y":2, "z":69}}
		var newPosition player.PlayerPosition
		err := json.Unmarshal([]byte(dataPacket), &newPosition)
		if err != nil {
			return fmt.Errorf("cant parse position player data")
		}
		globals.PlayerList[clientPacket.PlayerID].PlayerPosition = newPosition
	case packet.UpdateRotation: // {"type":7, "id": 0, "data":{"pitch":42, "yaw":11}}
		var newRotation player.PlayerRotation
		_ = json.Unmarshal([]byte(dataPacket), &newRotation)
		if err != nil {
			return fmt.Errorf("cant parse rotation player data")
		}
		globals.PlayerList[clientPacket.PlayerID].Rotation = newRotation
	default:
		if user, ok := globals.PlayerList[clientPacket.PlayerID]; ok {
			tcp.SendErrorMsg(user.TcpConnection, "Invalid UDP packet type!")
		}

	}

	return nil
}
func updatePlayerPosition() {
	for {
		if len(globals.PlayerList) > 1 {
			for _, user := range globals.PlayerList {
				//TODO send name or id as well
				//BUG where only one recieves
				BroadcastUDP(user.PlayerPosition, packet.PositionBroadcast, []int{user.Id})
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// function wont send the message for players in the filter
func BroadcastUDP(data interface{}, packetType int8, userFilter []int) error {
	packetToSend := packet.StampPacket(data, packetType)
	for _, user := range globals.PlayerList {
		if !utils.IntInArray(user.Id, userFilter) && user.UdpAddress != nil {
			_, err := packetToSend.SendUdpStream(udpConnection, user.UdpAddress)
			if err != nil {
				log.Println(err)
			}
		}
	}
	return nil
}
