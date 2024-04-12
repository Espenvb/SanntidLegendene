package main

import (
	"encoding/binary"
	. "fmt"
	"log"
	"net"
	"os/exec"
	t "time"
)

var counter uint64
var port = 9999
var buf = make([]byte, 16)

func startBackup() {
	(exec.Command("gnome-terminal", "-x", "sh", "-c", "go run ProcessPair.go")).Run()
	//(exec.Command("osascript", "-e", "tell app \"Terminal\" to do script \"go run Desktop/ex06/phoenix.go\"")).Run()

	Println("New backup up and running!")
}

func main() {
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:9999")
	isPrimary := false
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Println("Error, something went wrong...")
	}
	
	log.Println("Hi, I'm the backup!")

	// backup loop
	for !(isPrimary) {
		conn.SetReadDeadline(t.Now().Add(2 * t.Second))
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			isPrimary = true
		} else {
			counter = binary.BigEndian.Uint64(buf[:n])
		}
	}
	conn.Close()

	Println("addr: ", addr)
	startBackup()
	Println("I'm now the primary!")
	bcastConn, _ := net.DialUDP("udp", nil, addr)

	// primary loop
	for {
		if counter%10 == 0 {
			Println("\t*---------------*")
			Println("\t| Number: ", counter, "\t|")
		} else {
			Println("\t| Number: ", counter, "\t|")
		}
		counter++
		binary.BigEndian.PutUint64(buf, counter)
		_, _ = bcastConn.Write(buf)
		t.Sleep(50 * t.Millisecond)
	}
}
//we are done!