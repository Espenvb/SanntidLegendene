package main

import (
	"fmt"
	"net"
	"time"
)

func receiver(conn net.Conn) {
	for {
		buf := make([]byte, 1024)
		n, _ := conn.Read(buf)
		fmt.Printf("Received: %s\n", buf[:n])
	}
}

func sender(conn net.Conn) {
	time.Sleep(time.Second * 1)
	for {
		msg := "Hello from 6"
		buf := append([]byte(msg), 0) // append zero-byte \0 to signal end of tcp payload
		conn.Write(buf)

		time.Sleep(time.Second * 1)
	}
}

func main() {
	conn, _ := net.Dial("tcp", "10.100.23.129:33546")
	defer conn.Close()

	go receiver(conn)
	go sender(conn)

	for {
		time.Sleep(time.Second * 10)
	}
}
