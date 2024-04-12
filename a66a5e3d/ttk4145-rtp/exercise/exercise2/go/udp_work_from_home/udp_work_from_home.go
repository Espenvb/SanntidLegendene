package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

func sendResponse(conn *net.UDPConn, addr *net.UDPAddr) {
	//_, err := conn.WriteToUDP([]byte("From server: Hello I got your message "), addr)
	conn.WriteToUDP([]byte("From server: Hello I got your message "), addr)
}

func talkToServer(ipaddress, port string) {
	p := make([]byte, 2048)
	conn, err := net.Dial("udp", ipaddress+":"+port)
	if err != nil {
		fmt.Printf("Some error %v", err) // TOFIX: read udp 127.0.0.1:54762->127.0.0.1:20035: read: connection refused (probably firewall/network-config issues)
		return
	}
	fmt.Fprintf(conn, "Hi UDP Server, How are you doing?")
	_, err = bufio.NewReader(conn).Read(p)
	if err == nil {
		fmt.Printf("%s\n", p)
	} else {
		fmt.Printf("Some error %v\n", err)
	}
	conn.Close()
}

func listenToBroadcast() {
	fmt.Printf("hello Broadcast")
	receive_buffer := make([]byte, 1024)
	//ipaddress, _ := net.ResolveIPAddr("wlp0s20f3", "10.22.70.219")
	addr := net.UDPAddr{
		Port: 30000,
		//IP: net.ParseIP(ipaddress.String()),
		//IP: net.ParseIP("10.22.70.255"),
		//IP:   net.ParseIP("#.#.#.255"), // QUESTION: How to use best?
		IP: nil,
	}
	server, _ := net.ListenUDP("udp", &addr)

	// Listen on UDP IP address announcement port 30000
	for {
		n, remoteaddr, _ := server.ReadFromUDP(receive_buffer)
		fmt.Printf("Broadcast: Received from %v the %d bytes message %s\n", remoteaddr, n, receive_buffer)
	}
}

func main() {
	go listenToBroadcast()

	go talkToServer("127.0.0.1", "20035")

	for {
		time.Sleep(time.Second * 10)
	}
}
