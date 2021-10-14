package main

import (
	"fmt"
	"github.com/olivercullimore/geo-energy-data/server"
)

func main() {
	// Run server
	fmt.Println("                ___                          ___       _")
	fmt.Println(" __ _ ___ ___  | __|_ _  ___ _ _ __ _ _  _  |   \\ __ _| |_ __ _")
	fmt.Println("/ _` / -_) _ \\ | _|| ' \\/ -_) '_/ _` | || | | |) / _` |  _/ _` |")
	fmt.Println("\\__, \\___\\___/ |___|_||_\\___|_| \\__, |\\_, | |___/\\__,_|\\__\\__,_|")
	fmt.Println("|___/                           |___/ |__/")
	fmt.Println("----------------------------------------------------------------")
	server.Run()
}
