package main

import (
	"fmt"
	"net"
)

func main() {
	udp_broadcast_addr := net.UDPAddr{IP: nil, Port: 30000, Zone: ""}
	udp_listen_conn, err := net.ListenUDP("udp", &udp_broadcast_addr)
	if err != nil {
		fmt.Println("Error on net.ListenUDP:", err)
	}
	defer udp_listen_conn.Close()

	buf := make([]byte, 1024)
	for {
		n, addr, err := udp_listen_conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error on conn.ReadFromUDP: ", err)
		}

		fmt.Println("Received ", string(buf[0:n]), " from ", addr)
	}
}
