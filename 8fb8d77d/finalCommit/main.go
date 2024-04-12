package main

import (
	"Heis/driver-go/elevio"
	"Heis/elevator"
	"Heis/master"
	"Heis/network/establish_connection"
	"Heis/network/peers"
	"Heis/process_pair"
	"Heis/single_elev"
	"Heis/slave"
	"Heis/timer"
	"net"
)

func main() {

	go process_pair.Primary(process_pair.Backup())

	const (
		CONN_PORT = "8070"
	)

	_numFloors := elevio.NumFloors

	elevio.Init("localhost:15657", _numFloors)

	// Elevator initialization
	elevId := peers.MakeId()

	buffer := 10

	// Channels to update the the FSM-functions:
	newOrderCh := make(chan map[string]elevator.Elevator, buffer)
	elevUpdateRealtimeCh := make(chan elevator.Elevator, buffer)

	// Channels for door-timer and light:
	doorTimerCh := make(chan bool, buffer)
	timedOut := make(chan int, buffer)
	lightsCh := make(chan int, buffer)

	// Channels for inputs:
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	// Master channels:
	masterConnCh := make(chan net.Conn, buffer)
	connectionsCh := make(chan map[string]net.Conn, buffer)
	masterIdToEstablishConnToSlaves := make(chan string, buffer)
	masterIdToAlertMasterCh := make(chan string, buffer)
	masterIdToSendAndReceiveToMasterCh := make(chan string, buffer)
	sendMapToSlavesCh := make(chan map[string]elevator.Elevator, buffer)
	getElevFromSlaveCh := make(chan elevator.Elevator, buffer)
	isMasterCh1 := make(chan bool, buffer)
	isMasterCh2 := make(chan bool, buffer)

	// Slave channels:
	slaveConnCh := make(chan net.Conn, buffer)
	sendMyselfToMasterTx := make(chan elevator.Elevator, buffer)
	receiveMapFromMasterCh := make(chan map[string]elevator.Elevator, buffer)

	// Channels for Heartbeat:
	peerUpdateCh1 := make(chan peers.PeerUpdate)
	peerUpdateCh2 := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	go peers.Transmitter(15623, elevId, peerTxEnable)
	go peers.Receiver(15623, peerUpdateCh1)
	go peers.PeerUpdates(elevId, peerUpdateCh1, peerUpdateCh2, isMasterCh1, isMasterCh2, masterIdToEstablishConnToSlaves, masterIdToAlertMasterCh)

	go establish_connection.EstablishConnToSlaves(elevId, CONN_PORT, masterConnCh, connectionsCh, isMasterCh1)
	go slave.AlertMaster(CONN_PORT, elevId, masterIdToAlertMasterCh, masterIdToSendAndReceiveToMasterCh, slaveConnCh)

	go single_elev.ButtonsAndRequests(elevId, elevUpdateRealtimeCh,
		drv_buttons, sendMapToSlavesCh, getElevFromSlaveCh, receiveMapFromMasterCh,
		newOrderCh, lightsCh, sendMyselfToMasterTx, isMasterCh2)

	go single_elev.OrderExecution(elevId, elevUpdateRealtimeCh, drv_floors,
		newOrderCh, doorTimerCh, timedOut, lightsCh)

	go timer.Timer(doorTimerCh, timedOut)

	go slave.SendAndReceiveToMaster(elevId, slaveConnCh, masterIdToSendAndReceiveToMasterCh, receiveMapFromMasterCh, sendMyselfToMasterTx)
	go master.SendAndReceiveToSlaves(elevId, peerUpdateCh2, masterConnCh, connectionsCh, sendMapToSlavesCh, getElevFromSlaveCh)

	select {}
}
