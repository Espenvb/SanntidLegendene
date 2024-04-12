package master

import (
	"Heis/elevator"
	"Heis/network/peers"
	"Heis/network/tcp"
	"fmt"
	"net"
	"sync"
)

/*
--------------------------------------------------------------------------
--! @file
--! @brief This file contain functions and variables regarding the master
--------------------------------------------------------------------------
*/

var mutex sync.Mutex


func SendAndReceiveToSlaves(id string, peerCh chan peers.PeerUpdate, masterConnCh <-chan net.Conn, connectionsCh <-chan map[string]net.Conn,
	sendMapToSlavesCh <-chan map[string]elevator.Elevator, getElevFromSlave chan elevator.Elevator) {
	var connections map[string]net.Conn
	for {
		select {
		case c := <-peerCh:
			peers.PrintUpdatedPeers(c)
			if len(c.Lost) != 0 {
				mutex.Lock()
				for i := 0; i < len(c.Lost); i++ {
					for k := range connections {
						if k == c.Lost[i] {
							delete(connections, k)
						}
					}
				}
				mutex.Unlock()
			}
		case c := <-connectionsCh:
			mutex.Lock()
			connections = c
			fmt.Println()
			mutex.Unlock()
		case c := <-masterConnCh:
			go tcp.Receive(c, getElevFromSlave)
		case c := <-sendMapToSlavesCh:
			mutex.Lock()
			for _, v := range connections {
				tcp.Transmit(v, c) 
			}
			mutex.Unlock()
		}
	}
}
