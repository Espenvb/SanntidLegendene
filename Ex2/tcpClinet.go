package main

import (
	"log"
	"net"
	"fmt"
)

//portDel := "33546"


func main(){
	addr, err := net.ResolveTCPAddr("tcp", "10.100.23.129:34933")
	if err!=nil {
		log.Fatal(err)
	}

	conn, err := net.ListenTCP("tcp", addr)
	if err!=nil {
		log.Fatal(err)
	}

	defer conn.Close()

	buffer := make([]byte, 1024)

	n, _, err := conn.Read
	if err!=nil {
		log.Fatal(err)
	}

	message := string(buffer[:n])
	fmt.Printf("melding:", message)


	
}



