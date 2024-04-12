package main

import (
	"net"
	"time"
)

func main() {

	serverAddr, _ := net.ResolveUDPAddr("udp", "10.100.23.129:20006")

	conn, _ := net.DialUDP("udp", nil, serverAddr)
	defer conn.Close()

	for {
		msg := "Hello server from place 06"
		buf := []byte(msg)
		conn.Write(buf)
		time.Sleep(time.Second * 1)
	}
}
