package main

import (
	//"log"
	"net"
	"fmt"
)

func main(){
	// Resolve UDP address to listen on
	addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", 12345))
		
	// Create UDP listener
	conn, _ := net.ListenUDP("udp", addr)
	defer conn.Close()
	a := 0
	dataReceived := make(chan bool)

go func(){
	buffer := make([]byte, 1024)
	for{
		n,_,err := conn.ReadFromUDP(buffer)
		if err != nil{
			fmt.Println("error",err)
			return
		}
		dataReceived <- true
		a
	}
	
	}

	
}
