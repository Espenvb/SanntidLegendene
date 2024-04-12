package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

func getBytes(msg string) []byte {
	return append([]byte(msg), 0)
}

func getLocalIP() string {
	addresses, _ := net.InterfaceAddrs()

	for _, address := range addresses {
		ipnet, ok := address.(*net.IPNet) // kind of a cast
		//fmt.Println(address.String(), ipnet.Network(), ipnet.IP.IsLoopback(), ipnet.IP.DefaultMask(), ipnet.IP.IsGlobalUnicast())

		if ok && // check type to be *net.IPNet
			ipnet.IP.DefaultMask() != nil && // filter IPv6 addresses
			ipnet.IP.IsLoopback() == false { // filter local only addresses
			return strings.Split(address.String(), "/")[0]
		}
	}

	return ""
}

func getServerIP() string {
	return ""
}

func sendConnectRequestViaUDP(server_ip string, udp_port string, local_ip string, local_port string, tries int) {
	server_addr, _ := net.ResolveUDPAddr("udp", server_ip+":"+udp_port)
	connection, err := net.DialUDP("udp", nil, server_addr)
	if err != nil {
		fmt.Println("Error in sendConnectRequestViaUDP() while dialing:", err)
	}
	defer connection.Close()

	for i := 0; i < tries; i++ {
		msg := "Connect to: " + local_ip + ":" + local_port
		_, err = connection.Write(getBytes(msg))
		if err != nil {
			fmt.Println("Error in sendConnectRequestViaUDP() while sending:", err)
		}
		time.Sleep(time.Second * 1)
	}
}

func sendConnectRequestViaTCP(server_ip, tcp_port, local_ip, local_port string) {

}

func handleTCPConnection(connection net.Conn) {
	for {
		buffer := make([]byte, 2048)
		n, _ := connection.Read(buffer)
		received_msg := string(buffer[:n])

		fmt.Println(connection.RemoteAddr().String(), received_msg)

		thanks := "Thanks!"
		connection.Write(getBytes(thanks + " Server, you said " + received_msg))
		if strings.Count(received_msg, thanks) >= 2 {
			connection.Close()
		}
		time.Sleep(time.Second * 1)
	}
}

func listenTCP(port string) {
	//local, _ := net.ResolveTCPAddr("tcp", my_ip)
	//accept_socket, _ := net.ListenTCP("tcp", local)

	//accept_socket.AcceptTCP() // accepting any connection on that address

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error in listenTCP while creating Listener:", err)
	}
	defer listener.Close()

	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("Error in listenTCP while accepting connection:", err)
		}
		go handleTCPConnection(connection)
	}
}

func main() {
	local_ip := getLocalIP()
	fmt.Println(local_ip)
	tcp_port := "8080"

	server_ip := "10.100.23.129"
	udp_port := "20006"

	go listenTCP(tcp_port)
	go sendConnectRequestViaUDP(server_ip, udp_port, local_ip, tcp_port, 3)

	for {
		time.Sleep(time.Second * 10)
	}
}
