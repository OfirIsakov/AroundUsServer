package cli

import (
	"aroundUsServer/globals"
	"fmt"
	"log"
	"strconv"
	"strings"
)

func ConsoleCLI() {
	for {
		var command string
		fmt.Scanln(&command)
		commands := strings.Split(strings.Trim(command, "\n\t /\\'\""), " ")
		switch commands[0] {
		case "help", "h":
			log.Println("help(h)")
			log.Println("list(ls)")
			log.Println("disconnet(dc) [id]")
		case "list", "ls":
			for id, player := range globals.PlayerList {
				log.Println(fmt.Sprintf("%v) %v", id, player))
			}
		case "disconnet", "dc":
			_, err := strconv.Atoi(commands[1])
			if err != nil {
				log.Println("Cant convert to number position")
			}
			// globals.PlayerList[id].TcpConnection.Close()
		default:
			log.Println("Unknown command")
		}
	}
}
