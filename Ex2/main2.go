package main

import (
	"log"
	"net"
	"fmt"
)

func main(){
	port := 30000

	addr, err := net.ResolveUDPAddr("udp",fmt.Sprintf(":%d",port))
	if err!=nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp",addr)
	if err!=nil {
		log.Fatal(err)
	}

	defer conn.Close()

	
	buffer := make([]byte,1024)

	n, _, err := conn.ReadFromUDP(buffer)
	if err!=nil {
		log.Fatal(err)
		return
	}

	message := string(buffer[:n])
	fmt.Printf("melding:", message)
	
	
}
	