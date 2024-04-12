package main

import (
	"flag"
	"fmt"
	"project/alive"
	"project/backup"
	"project/communication"
	"project/connection"
	"project/distribution"
	"project/elevio"
	"project/execution"
	"project/primary"
	"project/primary_reconfig"
	"project/process_pair"
	pt "project/project_types"
	"time"
)

// Flags
// TODO: define non-identifying default values
var isSupervisor = flag.Bool("is-supervisor", false, "Used internally for process pair. Default is false.")

var groupId = flag.Int("group-id", 100, "Change the group id used for messages. Default is 100.")
var broadcastPort = flag.Int("bport", 10000, "Change broadcast port used for sending alive messages. Default is 10000.")
var communicationPort = flag.Int("cport", 10001, "Change local port used for non broadcast communication. Default is 10001.")

var elevatorServerIP = flag.String("elevator-server-ip", "localhost", "Set elevator server IP/Host that elevio shall use. Default is localhost.")
var elevatorServerPort = flag.Int("eport", 15657, "Set elevator server port that elevio shall use. Default is 15657.")
var m_numFloors = flag.Int("m", 4, "Set number of floors of the elevator system. Default is 4.")

var hraExecutablePath = flag.String("hra", "../Project-resources/cost_fns/hall_request_assigner/hall_request_assigner", "Set the path for the Hall Request Assigner executable.")

// channels for connecting driver/execution
var aliveMessageChanTx = make(chan pt.AliveMessage)        // blocking to make sure, it is actually sent before proceding
var aliveMessageChanRx = make(chan pt.AliveMessage, 1)     // non-blocking to not loose other incomming messages
var nodeUpdateChanToReconfig = make(chan pt.NodeUpdate, 3) // not fully sure if should be blocking

var orderChanRx = make(chan pt.Order, 1)              // non-blocking to not loose other incomming messages
var elevioFloorUpdateChan = make(chan int)            // blocking to make sure a passed floor is actually recognized by the state machine
var obstructionSwitch = make(chan bool)               // blocking to make sure no obstruction change is missed
var obstructedChan = make(chan bool)                  // blocking to make sure it is actually done
var elevatorStateChanTx = make(chan pt.Elevator)      // blocking to make sure, it is actually sent before proceding
var orderAckChanTx = make(chan pt.OrderAck)           // blocking to make sure, it is actually sent before proceding
var floorServicedChanTx = make(chan pt.FloorServiced) // blocking to make sure, it is actually sent before proceding

var buttonPressChanToFwd = make(chan elevio.ButtonEvent) // blocking to make sure a button press is actually recognized by forwarder

// channels for conecting distribution
var systemOrderChanToDistrBackup = make(chan pt.SystemOrder) // blocking because only one button press at a time can be processed (send to backup before eventually lighting up)
var nodeUpdateChanToDistrBackup = make(chan pt.NodeUpdate)   // blocking because it is second required input before starting to distribute to the (alive) backups
var systemOrderAckChanRx = make(chan pt.SystemOrderAck, 1)   // non-blocking to not loose other incomming messages
var successfulChanDistrBackup = make(chan bool, 1)           // non-blocking to allow primary early exit and BackupDistribution then can still take new distribution tasks from primary

var ordersChanToDistrOrders = make(chan map[pt.Node]pt.Order) // blocking to be sure sending is actually started and the task before has been either successsful or timed-out
var orderAckChanRx = make(chan pt.OrderAck, 1)                // non-blocking to not loose other incomming messages
var successfulChanDistrOrder = make(chan bool, 1)             // non-blocking to make OrderDistribution to still taking new tasks on a primary restart

var systemOrderChanToDistrLight = make(chan pt.SystemOrder, 1) // non-blocking because sometimes we do not wait until successful
var nodeUpdateChanToDistrLight = make(chan pt.NodeUpdate, 1)   // non-blocking because sometimes we do not wait until successful
var lightOrderAckChanRx = make(chan pt.LightOrderAck, 1)       // non-blocking to not loose other incomming messages
var successfulChanDistrLight = make(chan bool, 1)              // non-blocking to make LightDistribution to still taking new tasks on a primary restart

var systemOrderChanRx = make(chan pt.SystemOrder)          // non-blocking to not loose other incomming messages
var systemOrderAckChanTx = make(chan pt.SystemOrderAck)    // blocking to make sure, it is actually sent before proceding
var systemOrderRecoveryChan = make(chan pt.SystemOrder, 1) // non-blocking because backup is continously feeding into this channel, but primary just picks up value when needed

var primaryChangeChan = make(chan pt.Node)            // blocking to ensure transmitter is sending to new primary
var stopButtonChan = make(chan bool)                  // blocking to ensure stop button press is recognized by execution
var lightOrderChanRx = make(chan pt.LightOrder, 1)    // non-blocking to not loose other incomming messages
var lightOrderAckChanTx = make(chan pt.LightOrderAck) // blocking to make sure, it is actually sent before proceding

var buttonPressChanTx = make(chan pt.ButtonPress) // blocking to make sure, it is actually sent before proceding

var buttonPressChanRx = make(chan pt.ButtonPress, 1)      // non-blocking to not loose other incomming messages
var elevatorStateChanRx = make(chan pt.Elevator, 1)       // non-blocking to not loose other incomming messages
var floorServicedChanRx = make(chan pt.FloorServiced, 1)  // non-blocking to not loose other incomming messages
var nodeUpdateChanToPrimary = make(chan pt.NodeUpdate, 1) // non-blocking because pause should handle it, but not fully sure
var pausePrimaryChan = make(chan bool)                    // blocking to make sure primary will acutally stop

var primaryAnnounceChanRx = make(chan pt.PrimaryAnnounce, 1) // non-blocking to not loose other incomming messages
var primaryAnnounceChanTx = make(chan pt.PrimaryAnnounce)    // blocking to make sure, it is actually sent before proceding
// Channels
var pairMessageChanRx = make(chan string, 1)
var pairMessageChanTx = make(chan string)

// var nodeTypeUpdate = make(chan node.Type)               // blocking to make sure, it is actually sent before proceding
// var aliveSendChan = make(chan alive.AliveMessage)       // blocking to make sure, it is actually sent before proceding
// var aliveReceiveChan = make(chan alive.AliveMessage, 1) // non-blocking to not loose messages
// var nodeUpdateChan = make(chan alive.NodeUpdate, 1)     // non-blocking to not loose messages

// var targetFloorUpdate = make(chan execution.TargetFloor, 1) // non-blocking to not loose other messages in the receiver
// var floorUpdate = make(chan int)                            // blocking to make sure a passed floor is actually registered by the state machine
// var obstruction = make(chan bool)                           // blocking to make sure no obstruction change is missed
// var buttonPressChan = make(chan elevio.ButtonEvent)         // blocking to make sure a button press is actually sent by transmitter
// var hallRequestsChan = make(chan [][2]bool, 1)              // non-blocking to not loose other messages in the receiver
// var cabRequestsChan = make(chan []bool, 1)                  // non-blocking to not loose other messages in the receiver

func supervisorProgram() {
	communication.Init(*communicationPort, *broadcastPort)

	go communication.Receiver(*communicationPort+100, pairMessageChanRx)
	go process_pair.ObserverProcessPair("main-program", pairMessageChanRx)

	// go communication.TransmitterUnicast(*communicationPort, "localhost", *communicationPort+200, pairMessageChanTx)
	mainProgramProcesspairNodeChan := make(chan pt.Node)
	go communication.TransmitterUnicast(mainProgramProcesspairNodeChan, pairMessageChanTx)
	mainProgramProcesspairNodeChan <- pt.Node{
		GroupId: *groupId,
		IP:      "localhost",
		Port:    *communicationPort + 200,
	}
	go process_pair.SendProcessPairAliveMessages(pairMessageChanTx)

	for {
		time.Sleep(5 * time.Second)
	}
}

func mainProgram() {
	communication.Init(*communicationPort, *broadcastPort)
	go communication.Receiver(*communicationPort+200, pairMessageChanRx)
	go process_pair.ObserverProcessPair("supervisor", pairMessageChanRx)

	// setup local data
	myNode := pt.Node{
		GroupId: *groupId,
		IP:      connection.GetLocalIP(),
		Port:    *communicationPort,
	}

	// go communication.TransmitterUnicast(*communicationPort, "localhost", *communicationPort+100, pairMessageChanTx)
	mainProgramProcesspairNodeChan := make(chan pt.Node)
	go communication.TransmitterUnicast(mainProgramProcesspairNodeChan, pairMessageChanTx)
	mainProgramProcesspairNodeChan <- pt.Node{
		GroupId: *groupId,
		IP:      "localhost",
		Port:    *communicationPort + 100,
	}
	go process_pair.SendProcessPairAliveMessages(pairMessageChanTx)

	elevServerAddress := fmt.Sprintf("%s:%d", *elevatorServerIP, *elevatorServerPort)

	// initialize packages
	// communication.Init(*communicationPort, *broadcastPort)
	alive.Init(myNode)
	backup.Init(myNode, *m_numFloors)
	elevio.Init(elevServerAddress, *m_numFloors)
	execution.Init(myNode, *m_numFloors)
	primary.Init(*m_numFloors, *hraExecutablePath)
	primary_reconfig.Init(myNode)

	// setup goroutines

	// BROADCAST:
	go communication.Receiver(*broadcastPort,
		aliveMessageChanRx,
		primaryAnnounceChanRx,
	)
	go communication.TransmitterBroadcast(
		aliveMessageChanTx,
		primaryAnnounceChanTx,
	)

	// UNICAST:
	go communication.Receiver(*communicationPort,
		systemOrderChanRx,
		systemOrderAckChanRx,
		orderAckChanRx,
		lightOrderAckChanRx,
		lightOrderChanRx,
		orderChanRx,
		buttonPressChanRx,
		elevatorStateChanRx,
		floorServicedChanRx,
	)
	go communication.TransmitterUnicast(primaryChangeChan,
		systemOrderAckChanTx,
		lightOrderAckChanTx,
		buttonPressChanTx,
		elevatorStateChanTx,
		orderAckChanTx,
		floorServicedChanTx,
	)

	go alive.Transmitter(obstructedChan, aliveMessageChanTx)
	go alive.Receiver(aliveMessageChanRx, nodeUpdateChanToReconfig)

	go backup.BackupService(systemOrderChanRx, systemOrderAckChanTx, systemOrderRecoveryChan)

	// setup goroutines for primary and distribution
	go distribution.BackupDistribution(systemOrderChanToDistrBackup, nodeUpdateChanToDistrBackup, systemOrderAckChanRx, successfulChanDistrBackup)
	go distribution.OrderDistribution(ordersChanToDistrOrders, orderAckChanRx, successfulChanDistrOrder)
	go distribution.LightDistribution(systemOrderChanToDistrLight, nodeUpdateChanToDistrLight, lightOrderAckChanRx, successfulChanDistrLight)

	go elevio.PollButtons(buttonPressChanToFwd)
	go elevio.PollFloorSensor(elevioFloorUpdateChan)
	go elevio.PollStopButton(stopButtonChan)
	go elevio.PollObstructionSwitch(obstructionSwitch)

	go execution.ButtonLightService(lightOrderChanRx, lightOrderAckChanTx)
	go execution.ButtonPressForwarder(buttonPressChanToFwd, buttonPressChanTx)
	go execution.ExecutionFloor(orderChanRx, elevioFloorUpdateChan, obstructionSwitch, elevatorStateChanTx, orderAckChanTx, floorServicedChanTx, obstructedChan)

	go primary.PrimaryService(
		// input from driver/exection:
		buttonPressChanRx,
		elevatorStateChanRx,
		floorServicedChanRx,
		// input from primary_reconfig:
		nodeUpdateChanToPrimary,
		pausePrimaryChan,
		// input from backup:
		systemOrderRecoveryChan,
		// input from distribution:
		successfulChanDistrBackup,
		successfulChanDistrOrder,
		successfulChanDistrLight,
		// output to distribution:
		systemOrderChanToDistrBackup,
		nodeUpdateChanToDistrBackup,
		ordersChanToDistrOrders,
		systemOrderChanToDistrLight,
		nodeUpdateChanToDistrLight,
	)

	go primary_reconfig.ReconfigurationService(nodeUpdateChanToReconfig, primaryAnnounceChanRx, nodeUpdateChanToPrimary, primaryAnnounceChanTx, pausePrimaryChan, primaryChangeChan)

	// Plug-in debugging:
	//go alive.DebugAlive(nodeUpdateToReconfigChan, obstructedChan)
	//go execution.DebugServiceFloor(targetFloorUpdate)
	//go execution.DebugLightService(hallRequestsChan, cabRequestsChan)
	//go execution.DebugButtonPressForwarding(buttonPressChanToFwd)
	//go execution.DebugExecutionFloor(orderChanRx, elevioFloorUpdateChan, obstructionSwitch, ownNodeChangedChan, elevatorChanTx, orderAckChanTx, floorServicedChanTx)
	for {
		time.Sleep(5 * time.Second)
	}
}

func deferredSleep() {
	for {
		time.Sleep(5 * time.Second)
	}
}

func main() {
	flag.Parse()
	process_pair.Init(*isSupervisor, *groupId, *broadcastPort, *communicationPort, *elevatorServerIP, *elevatorServerPort, *m_numFloors, *hraExecutablePath)

	fmt.Println("--is-supervisor =", *isSupervisor)

	if *isSupervisor {
		supervisorProgram()
	} else {
		mainProgram()
	}
}
