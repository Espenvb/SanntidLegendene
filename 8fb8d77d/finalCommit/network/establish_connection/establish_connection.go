package establish_connection

import (
	"Heis/network/peers"
	"fmt"
	"net"
)

const (
	CONN_TYPE = "tcp"
)

func EstablishConnToMaster(port string, id string, masterIp string) (net.Conn, error) {
	addr, err := net.ResolveTCPAddr(CONN_TYPE, masterIp+":"+port)
	if err != nil {
		fmt.Println("Error resolving address:", err)
		return nil, err
	}
	conn, err := net.Dial(CONN_TYPE, addr.String())
	if err != nil {
		fmt.Printf("[error] Failed to Dial: %v\n", err)
		return nil, err
	}
	conn.Write([]byte(id))
	return conn, err
}


func EstablishConnToSlaves(id string, port string, masterConnCh chan<- net.Conn, connectionsCh chan<- map[string]net.Conn, isMasterCh chan bool) (net.Conn, error) {
	isMaster := false
	connections := make(map[string]net.Conn)
	buffer := make([]byte, 1024)
	var listener net.Listener 
	for {
		select {
		case c := <-isMasterCh:
			isMaster = c
			if isMaster {
				masterIp := peers.ExtractIpFromPeer(id)
				addr, err := net.ResolveTCPAddr("tcp", masterIp+":"+port)
				if err != nil {
					fmt.Println("Error resolving address:", err)
					return nil, err
				}
				listener, err = net.ListenTCP("tcp", addr)
				if err != nil {
					fmt.Println("Error creating listener:", err)
					return nil, err
				}
				defer listener.Close()
				fmt.Println("Server listening on", addr.String())
				connections = make(map[string]net.Conn)
				buffer = make([]byte, 1024)
			}
		default:
			if isMaster {
				fmt.Println("[EstablishConnToSlaves] Waiting for slaves trying to connect")
				masterConn, err := listener.Accept()
				if err != nil {
					fmt.Println("Error accepting connection:", err)
					continue
				}
				fmt.Println("Accepted connection on port: " + port)
				k, err := masterConn.Read(buffer)
				if err != nil {
					fmt.Printf("[error] Failed to read: %v\n", err)
					return nil, err
				}
				id := string(buffer[0:k])
				connections[id] = masterConn
				masterConnCh <- masterConn
				connectionsCh <- connections
			}
		}
	}
}

