package slave

import (
	"Heis/elevator"
	"Heis/network/establish_connection"
	"Heis/network/peers"
	"Heis/network/tcp"
	"fmt"
	"net"
)

func AlertMaster(port string, id string, masterIdToAlertMasterCh chan string, masterIdToSendAndReceiveCh chan string, slaveConnCh chan<- net.Conn) {
	var slaveConn net.Conn = nil
	var err error                
	select {
	case c := <-masterIdToAlertMasterCh:
		if id != c {
			masterIp := peers.ExtractIpFromPeer(c)
			slaveConn, err = establish_connection.EstablishConnToMaster(port, id, masterIp)
			masterIdToSendAndReceiveCh <- c
			slaveConnCh <- slaveConn
			if err != nil {
				fmt.Printf("[error] Failed to Dial: %v\n", err)
				return
			}
		}
	}
}

func SendAndReceiveToMaster(id string, slaveConnCh <-chan net.Conn, masterIdToSendAndReceiveToMasterCh chan string,
	receiveMapFromMasterCh chan map[string]elevator.Elevator, sendElevToMaster chan elevator.Elevator) {
	var elev elevator.Elevator
	masterId := ""
	var slaveConn net.Conn
	for {
		if masterId != id {
			select {
			case c := <-masterIdToSendAndReceiveToMasterCh:
				masterId = c
			case c := <-slaveConnCh:
				slaveConn = c
				go tcp.Receive(slaveConn, receiveMapFromMasterCh)
			case c := <-sendElevToMaster:
				elev = c
				tcp.Transmit(slaveConn, elev)
			}
		}
	}
}
