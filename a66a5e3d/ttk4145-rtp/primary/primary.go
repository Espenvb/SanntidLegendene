package primary

import (
	"fmt"
	"project/elevio"
	"project/hra"
	pt "project/project_types"
	"reflect"
)

// TODO: clean up: string-state to enum-like
type state string

const (
	s_pause state = "pause"
)

var _initialized = false
var _m_numFloors = 4
var _hraExecutablePath string

func Init(m_numFloors int, hraExecutablePath string) {
	_initialized = true
	_m_numFloors = m_numFloors
	_hraExecutablePath = hraExecutablePath
}

func noOrders(systemOrder pt.SystemOrder, aliveNodes pt.NodeUpdate) bool {
	for _, floor := range systemOrder.HallRequests {
		if floor[elevio.BT_HallUp] || floor[elevio.BT_HallDown] {
			return false
		}
	}
	for _, node := range aliveNodes.AliveNodes {
		for _, request := range systemOrder.CabRequests[node.String()] {
			if request {
				return false
			}
		}
	}
	return true
}

// does not change systemOrder even though containting reference types map and slice
// by copying when neccessary
func addBtnPress(systemOrder pt.SystemOrder, btnPress pt.ButtonPress) pt.SystemOrder {
	newSystemOrder := systemOrder
	emptyCabRequests := make([]bool, _m_numFloors)
	floor := btnPress.Button.Floor
	btn := btnPress.Button.Button
	id := btnPress.Node.String()
	if len(systemOrder.CabRequests[id]) == 0 {
		newSystemOrder.CabRequests[id] = emptyCabRequests
	}
	if btn == elevio.BT_Cab {
		copy(newSystemOrder.CabRequests[id], systemOrder.CabRequests[id])
		newSystemOrder.CabRequests[id][floor] = true
	} else {
		copy(newSystemOrder.HallRequests, systemOrder.HallRequests)
		newSystemOrder.HallRequests[floor][btn] = true
	}
	return systemOrder
}

// if missed aliveUpdate first reassign, else go to idle; always reset flag;
// this is to ensure that every alive-update is catched, as these are most important
// in order to fullfil button-light contract
// ommitting re-assignment when no new/lost nodes detected, because elevators
// should know its own remaining orders
func onLightDistrSuccessful(missedAliveUpdateIn bool) (state string, missedAliveUpdate bool) {
	if missedAliveUpdateIn {
		return "reassign_orders", false
	}
	return "idle", false
}

func onReassignOrders(
	systemOrder pt.SystemOrder,
	elevatorStatesIn map[pt.Node]pt.Elevator,
	assignment map[pt.Node]pt.Order,
	aliveNodes pt.NodeUpdate,
) (
	pendingAssignment map[pt.Node]pt.Order,
	newSystemOrder pt.SystemOrder,
	elevatorStates map[pt.Node]pt.Elevator,
	state string,
) {
	elevatorStates = elevatorStatesIn
	if noOrders(systemOrder, aliveNodes) {
		newSystemOrder = initEmptyCabRequests(systemOrder, aliveNodes)
		return emptyAssignment(aliveNodes), newSystemOrder, elevatorStates, "idle"
	}
	pendingAssignment, newSystemOrder, elevatorStates = hra.Assigner(_hraExecutablePath, systemOrder, elevatorStates, aliveNodes)
	if reflect.DeepEqual(assignment, pendingAssignment) {
		return assignment, newSystemOrder, elevatorStates, "idle"
	}
	return pendingAssignment, newSystemOrder, elevatorStates, "assign_orders"
}

func initEmptyCabRequests(systemOrder pt.SystemOrder, aliveNodes pt.NodeUpdate) pt.SystemOrder {
	newSystemOrder := systemOrder
	for _, node := range aliveNodes.AliveNodes {
		if _, contained := systemOrder.CabRequests[node.String()]; !contained {
			newSystemOrder.CabRequests[node.String()] = pt.EmptyCabRequests(_m_numFloors)
		}
	}
	return newSystemOrder
}

func emptyAssignment(aliveNodes pt.NodeUpdate) map[pt.Node]pt.Order {
	assignment := make(map[pt.Node]pt.Order, len(aliveNodes.AliveNodes))
	for _, node := range aliveNodes.AliveNodes {
		assignment[node] = pt.Order{
			HallRequests: pt.EmptyHallRequests(_m_numFloors),
			CabRequests:  pt.EmptyCabRequests(_m_numFloors),
		}
	}
	return assignment
}

func shouldPauseOnIdle(pause bool) string {
	if pause {
		return "pause"
	}
	return "idle"
}

// does not change systemOrder even though containting reference types map and slice
// by copying when neccessary
func removeServicedFloor(systemOrder pt.SystemOrder, floorServiced pt.FloorServiced) pt.SystemOrder {
	newSystemOrder := systemOrder
	btn := floorServiced.Floor.Button
	floor := floorServiced.Floor.Floor
	id := floorServiced.Node.String()
	if btn == elevio.BT_Cab {
		copy(newSystemOrder.CabRequests[id], systemOrder.CabRequests[id])
		newSystemOrder.CabRequests[id][floor] = false
	} else {
		copy(newSystemOrder.HallRequests, systemOrder.HallRequests)
		newSystemOrder.HallRequests[floor][btn] = false
	}
	return newSystemOrder
}

// Inputs
// button presses (hall and cab requests)
// alive list
// states of the elevators, i.e. behaviour, floor, direction
// floor serviced
// Outputs
// Order packages for all elevator drivers (hall and cab request execution)
// Order backup (only hall and cab requests)
// Button light config
func PrimaryService(
	// input from driver/exection:
	buttonPressChan <-chan pt.ButtonPress,
	elevatorStateChan <-chan pt.Elevator,
	floorServicedChan <-chan pt.FloorServiced,
	// input from primary_reconfig:
	nodeUpdateChan <-chan pt.NodeUpdate,
	pauseChan <-chan bool,
	// input from backup:
	systemOrderRecoveryChan <-chan pt.SystemOrder, // MUST be non-blocking channel // TODO: in backup continously send received backup to channel
	// input from distribution:
	successfulChanDistrBackup <-chan bool,
	successfulChanDistrOrder <-chan bool,
	successfulChanDistrLight <-chan bool,
	// output to distribution:
	systemOrderChanToBackupDistr chan<- pt.SystemOrder,
	nodeUpdateChanToBackupDistr chan<- pt.NodeUpdate,
	ordersChanToOrderDistr chan<- map[pt.Node]pt.Order,
	systemOrderChanToLightDistr chan<- pt.SystemOrder,
	nodeUpdateChanToLightDistr chan<- pt.NodeUpdate,
) {
	if !_initialized {
		panic("starting un-initialized PrimaryService")
	}
	systemOrder := pt.EmptySystemOrder(_m_numFloors)
	pendingSystemOrder := pt.EmptySystemOrder(_m_numFloors)
	elevatorStates := make(map[pt.Node]pt.Elevator)
	assignment := map[pt.Node]pt.Order{} // contains the _successfully_ distributed orders
	pendingAssignment := map[pt.Node]pt.Order{}
	aliveNodes := pt.NodeUpdate{} // TODO: consider initializing with _node, just to be sure??
	state := "pause"              // idle | backup_distr | order_accepted | remove_light | assign_orders | reassign_orders | order_distr | pause
	missedAliveUpdate := false

	for {
		switch state {
		case "pause":
			select {
			case pause := <-pauseChan:
				if !pause {
					state = "start"
				}
			case <-buttonPressChan:
			case <-elevatorStateChan:
			case <-floorServicedChan:
			case <-nodeUpdateChan: // enforcing channel feed order for resuming primary
			case <-systemOrderRecoveryChan: // ignore, allows backup to "update" content of channel, so it might be picked up on state start
			case <-successfulChanDistrBackup:
			case <-successfulChanDistrOrder:
			case <-successfulChanDistrLight:
			}
		case "start":
			systemOrder = <-systemOrderRecoveryChan // is beeing feeded continously by backup
			aliveNodes = <-nodeUpdateChan           // is beeing feeded directly after pauseChan by reconfig
			state = "reassign_orders"
		case "idle":
			select {
			case btnPress := <-buttonPressChan:
				pendingSystemOrder = addBtnPress(systemOrder, btnPress)
				fmt.Println("primary: idle: pendingSystemOrder:", pendingSystemOrder)
				if noOrders(pendingSystemOrder, aliveNodes) {
					state = "idle"
					break
				}
				systemOrderChanToBackupDistr <- pendingSystemOrder
				nodeUpdateChanToBackupDistr <- aliveNodes
				state = "backup_distr"
			case newElevState := <-elevatorStateChan: // TODO: consider larger channel in order to not loose elevator state updates
				elevatorStates[newElevState.Node] = newElevState
			case floorServiced := <-floorServicedChan: // TODO consider larger channel in order to not loose floorServiced messages when (stucking) in distribution phases
				fmt.Println("primary received: floor serviced", floorServiced)
				// floorServiced-check just here, as is should provide enough consistency
				systemOrder = removeServicedFloor(systemOrder, floorServiced)
				// omitting backup of changed systemorder, to save bandwidth, as not crutial for fulfilment of requirements (A elevator not ONE elev. shall arrive)
				systemOrderChanToLightDistr <- systemOrder
				nodeUpdateChanToLightDistr <- aliveNodes
				state = "remove_light"
			case aliveNodes = <-nodeUpdateChan:
				if _, keyExists := systemOrder.CabRequests[aliveNodes.New.String()]; !keyExists {
					systemOrder.CabRequests[aliveNodes.New.String()] = pt.EmptyCabRequests(_m_numFloors)
				}
				state = "reassign_orders"
			case pause := <-pauseChan:
				state = shouldPauseOnIdle(pause)
				fmt.Println("PrimaryService: idle: onPauseChannel =", pause, "new state =", state)
			case <-systemOrderRecoveryChan: // ignore in this state
			case <-successfulChanDistrBackup:
			case <-successfulChanDistrOrder:
			case <-successfulChanDistrLight:
			}
		case "backup_distr":
			select {
			case successful := <-successfulChanDistrBackup:
				if successful {
					state = "order_accepted"
				} else {
					state = "idle" // do nothing when backup distribution was not successful in order to fulfil button-light contract
				}
			case <-buttonPressChan:
			case newElevState := <-elevatorStateChan:
				elevatorStates[newElevState.Node] = newElevState
			case <-floorServicedChan:
			case aliveNodes = <-nodeUpdateChan:
				if _, keyExists := systemOrder.CabRequests[aliveNodes.New.String()]; !keyExists {
					systemOrder.CabRequests[aliveNodes.New.String()] = pt.EmptyCabRequests(_m_numFloors)
				}
				state = "reassign_orders" // early exit of backup-phase: reassignment is more important than confirming a button press in order to fulfill button-light contract
			case pause := <-pauseChan:
				if pause {
					// TODO: add stop channel to BackupDistribution
					//stopBackupDistribution <- true
					state = "pause"
				}
			case <-systemOrderRecoveryChan: // ignore in this state
			case <-successfulChanDistrOrder:
			case <-successfulChanDistrLight:
			}
		case "order_accepted":
			systemOrder = pendingSystemOrder
			// omit waiting for lightOrderAck because not crutial for button-light contract
			systemOrderChanToLightDistr <- systemOrder
			nodeUpdateChanToLightDistr <- aliveNodes
			state = "reassign_orders"
		case "remove_light": // not waiting for aliveUpdateChan because it is crutial to first sucessfully switch off lights (or panic); an eventual nodeUpdate is handled in idle state then
			select {
			case successful := <-successfulChanDistrLight:
				if !successful {
					panic("could not switch off light - probably a node died right after kicking off LightDistribution")
				}
				state, missedAliveUpdate = onLightDistrSuccessful(missedAliveUpdate)
				fmt.Println("primary: state:", state, "(after remove_light)")
			case <-buttonPressChan:
			case newElevState := <-elevatorStateChan:
				elevatorStates[newElevState.Node] = newElevState
			case <-floorServicedChan:
			case aliveNodes = <-nodeUpdateChan:
				if _, keyExists := systemOrder.CabRequests[aliveNodes.New.String()]; !keyExists {
					systemOrder.CabRequests[aliveNodes.New.String()] = pt.EmptyCabRequests(_m_numFloors)
				}
				missedAliveUpdate = true
			case pause := <-pauseChan:
				if pause {
					// TODO: add stop channel to LightDistribution
					//stopLightDistribution <- true
					state = "pause"
				}
			case <-systemOrderRecoveryChan: // ignore in this state
			case <-successfulChanDistrBackup:
			case <-successfulChanDistrOrder:
			}
		case "reassign_orders":
			pendingAssignment, systemOrder, elevatorStates, state = onReassignOrders(systemOrder, elevatorStates, assignment, aliveNodes)
			fmt.Println("primary:", state, "pendingAssignment:", pendingAssignment)
		case "assign_orders":
			ordersChanToOrderDistr <- pendingAssignment
			state = "order_distr" // could also be fallthrough
		case "order_distr":
			select {
			case successful := <-successfulChanDistrOrder:
				fmt.Println("order distribution successful =", successful)
				if successful {
					assignment = pendingAssignment
					state = "idle"
				} else {
					state = "assign_orders" // try it again // TODO: WONTFIX: define maximum number of retries
				}
			case <-buttonPressChan:
			case newElevState := <-elevatorStateChan:
				elevatorStates[newElevState.Node] = newElevState
			case <-floorServicedChan:
			case aliveNodes = <-nodeUpdateChan:
				if _, keyExists := systemOrder.CabRequests[aliveNodes.New.String()]; !keyExists {
					systemOrder.CabRequests[aliveNodes.New.String()] = pt.EmptyCabRequests(_m_numFloors)
				}
				state = "reassign_orders"
			case pause := <-pauseChan:
				if pause {
					// TODO: add stop channel to OrderDistribution
					//stopLightDistribution <- true
					state = "pause"
				}
			case <-systemOrderRecoveryChan: // ignore in this state
			case <-successfulChanDistrBackup:
			case <-successfulChanDistrLight:
			}
		}
	}
}
