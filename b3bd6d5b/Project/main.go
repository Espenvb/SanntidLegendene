package main

import (
	"Project/localElevator/elevio"
	"Project/localElevator/localElevatorHandler"
	"Project/network/messages"
	"Project/network/tcp"
	"Project/network/udpBroadcast"
	"Project/network/udpBroadcast/udpNetwork/peers"
	"Project/roleHandler/roleDistributor"
	"Project/roleHandler/roleFSM"
	"net"
)

func main() {
	const masterPort = "20050"
	elevio.Init("localhost:15657", elevio.N_FLOORS)
	buttonsCh := make(chan elevio.ButtonEvent)
	floorsCh := make(chan int)
	obstrCh := make(chan bool)
	peerUpdateToRoleDistributorCh := make(chan peers.PeerUpdate)
	roleAndSortedAliveElevsCh := make(chan roleDistributor.RoleAndSortedAliveElevs, 5)
	isMasterCh := make(chan bool, 5)
	editMastersConnMapCh := make(chan tcp.EditConnMap, 5)
	masterIPCh := make(chan net.IP)
	masterConnCh := make(chan net.Conn)
	sendNetworkMsgCh := make(chan tcp.SendNetworkMsg, 5)
	incommingNetworkMsgCh := make(chan []byte, 15)
	toSingleElevFSMCh := make(chan []byte, 5)
	toRoleFSMCh := make(chan []byte, 5)
	visibleOnNetwork := make(chan bool)

	
	if elevio.GetFloor() == -1 {
		localElevatorHandler.OnInitBetweenFloors()
	}
	localElevatorHandler.InitLights()
	
	go elevio.PollRequestButtons(buttonsCh)
	go elevio.PollFloorSensor(floorsCh)
	go elevio.PollObstructionSwitch(obstrCh)
	go udpBroadcast.StartPeerBroadcasting(peerUpdateToRoleDistributorCh, visibleOnNetwork)
	go roleDistributor.RoleDistributor(peerUpdateToRoleDistributorCh, roleAndSortedAliveElevsCh, masterIPCh)
	go roleFSM.RoleFSM(roleAndSortedAliveElevsCh, toRoleFSMCh, sendNetworkMsgCh, isMasterCh, editMastersConnMapCh)
	go tcp.EstablishMainListener(isMasterCh, masterPort, editMastersConnMapCh, incommingNetworkMsgCh)
	go tcp.EstablishConnectionAndListen(masterIPCh, masterPort, masterConnCh, incommingNetworkMsgCh)
	go tcp.SendMessage(sendNetworkMsgCh)
	go messages.DistributeMessages(incommingNetworkMsgCh, toSingleElevFSMCh, toRoleFSMCh)
	go localElevatorHandler.LocalElevatorHandler(buttonsCh, floorsCh, obstrCh, masterConnCh, sendNetworkMsgCh, toSingleElevFSMCh, visibleOnNetwork)
	
	for {
		select {}
	}
}
