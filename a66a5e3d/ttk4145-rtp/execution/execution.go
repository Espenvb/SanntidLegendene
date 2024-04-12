package execution

import (
	"fmt"
	"project/elevio"
	pt "project/project_types"
	"time"
)

const _doorTimeOut = 3 * time.Second

var _initialized = false
var _node pt.Node
var _m_numFloors = 4 // underscore to highlight that it should be used as read-only

func Init(node pt.Node, numFloors int) {
	if _initialized {
		panic("trying to re-initialize execution package!")
	}
	_initialized = true
	_node = node
	_m_numFloors = numFloors
}

func initElevator() pt.Elevator {
	initHallRequests := pt.EmptyHallRequests(_m_numFloors)
	initHallRequests[1][int(elevio.BT_HallUp)] = true
	return pt.Elevator{
		Floor:    -1,
		Dir:      elevio.MD_Stop,
		Behavior: pt.EB_Idle,
		Order: pt.Order{
			HallRequests: initHallRequests,
			CabRequests:  make([]bool, _m_numFloors),
		},
		Node: _node,
	}
}

func requestsAbove(e pt.Elevator) bool {
	for f := e.Floor + 1; f < _m_numFloors; f++ {
		if e.Order.HallRequests[f][int(elevio.BT_HallDown)] || e.Order.HallRequests[f][int(elevio.BT_HallUp)] || e.Order.CabRequests[f] {
			return true
		}
	}
	return false
}

func requestsBelow(e pt.Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		if e.Order.HallRequests[f][int(elevio.BT_HallDown)] || e.Order.HallRequests[f][int(elevio.BT_HallUp)] || e.Order.CabRequests[f] {
			return true
		}
	}
	return false
}

func requestsHere(e pt.Elevator) bool {
	if e.Order.HallRequests[e.Floor][int(elevio.BT_HallDown)] || e.Order.HallRequests[e.Floor][int(elevio.BT_HallUp)] || e.Order.CabRequests[e.Floor] {
		return true
	}
	return false
}

func shouldStop(e pt.Elevator) bool {
	switch e.Dir {
	case elevio.MD_Down:
		return e.Order.HallRequests[e.Floor][int(elevio.BT_HallDown)] || e.Order.CabRequests[e.Floor] || !requestsBelow(e)
	case elevio.MD_Up:
		return e.Order.HallRequests[e.Floor][int(elevio.BT_HallUp)] || e.Order.CabRequests[e.Floor] || !requestsAbove(e)
	case elevio.MD_Stop:
		fallthrough
	default:
		return true
	}
}

func requestsChooseDirection(e pt.Elevator) pt.Elevator {
	switch e.Dir {
	case elevio.MD_Up:
		if requestsAbove(e) {
			e.Dir = elevio.MD_Up
			e.Behavior = pt.EB_Moving
			return e
		} else if requestsHere(e) {
			e.Dir = elevio.MD_Stop
			e.Behavior = pt.EB_DoorOpen
			return e
		} else if requestsBelow(e) {
			e.Dir = elevio.MD_Down
			e.Behavior = pt.EB_Moving
			return e
		} else {
			e.Dir = elevio.MD_Stop
			e.Behavior = pt.EB_Idle
			return e
		}
	case elevio.MD_Down:
		if requestsBelow(e) {
			e.Dir = elevio.MD_Down
			e.Behavior = pt.EB_Moving
			return e
		} else if requestsHere(e) {
			e.Dir = elevio.MD_Stop
			e.Behavior = pt.EB_DoorOpen
			return e
		} else if requestsAbove(e) {
			e.Dir = elevio.MD_Up
			e.Behavior = pt.EB_Moving
			return e
		} else {
			e.Dir = elevio.MD_Stop
			e.Behavior = pt.EB_Idle
			return e
		}
	case elevio.MD_Stop:
		if requestsHere(e) {
			e.Dir = elevio.MD_Stop
			e.Behavior = pt.EB_DoorOpen
			return e
		} else if requestsAbove(e) {
			e.Dir = elevio.MD_Up
			e.Behavior = pt.EB_Moving
			return e
		} else if requestsBelow(e) {
			e.Dir = elevio.MD_Down
			e.Behavior = pt.EB_Moving
			return e
		} else {
			e.Dir = elevio.MD_Stop
			e.Behavior = pt.EB_Idle
			return e
		}
	default:
		e.Dir = elevio.MD_Stop
		e.Behavior = pt.EB_Idle
		return e
	}
}

func clearCurrentFloor(
	eIn pt.Elevator,
) (
	e pt.Elevator, clearDown bool, clearUp bool, clearCab bool,
) {
	e = eIn
	clearDown, clearUp, clearCab = false, false, false
	if e.Order.CabRequests[e.Floor] {
		clearCab = true
	}

	switch e.Dir {
	case elevio.MD_Up:
		if !requestsAbove(e) && !e.Order.HallRequests[e.Floor][int(elevio.BT_HallUp)] {
			if e.Order.HallRequests[e.Floor][int(elevio.BT_HallDown)] {
				clearDown = true
			}

		}
		if e.Order.HallRequests[e.Floor][int(elevio.BT_HallUp)] {
			clearUp = true
		}

	case elevio.MD_Down:
		if !requestsBelow(e) && !e.Order.HallRequests[e.Floor][elevio.BT_HallDown] {
			if e.Order.HallRequests[e.Floor][int(elevio.BT_HallUp)] {
				clearUp = true
			}

		}
		if e.Order.HallRequests[e.Floor][int(elevio.BT_HallDown)] {
			clearDown = true
		}

	case elevio.MD_Stop:
		fallthrough
	default:
		if e.Order.HallRequests[e.Floor][int(elevio.BT_HallUp)] {
			clearUp = true
		}
		if e.Order.HallRequests[e.Floor][int(elevio.BT_HallDown)] {
			clearDown = true
		}

	}
	return e, clearDown, clearUp, clearCab
}

func ExecutionFloor(
	// input:
	orderChan <-chan pt.Order,
	floorUpdate <-chan int,
	// TODO: WONTFIX: add StopButton-handling
	obstructionSwitch <-chan bool,
	// output:
	elevatorChan chan<- pt.Elevator,
	orderAckChan chan<- pt.OrderAck,
	floorServicedChan chan<- pt.FloorServiced,
	obstructedChan chan<- bool,
) {
	if !_initialized {
		panic("trying to start un-initialized execution package!")
	}
	clearUp := false
	clearDown := false
	clearCab := false
	obstructed := false
	doorTimeElapsed := false
	// elevator := initElevator()
	initHallRequests := pt.EmptyHallRequests(_m_numFloors)
	initHallRequests[1][int(elevio.BT_HallUp)] = true
	elevator := pt.Elevator{
		Floor:    0,
		Dir:      elevio.MD_Up,
		Behavior: pt.EB_Moving,
		Order: pt.Order{
			HallRequests: initHallRequests,
			CabRequests:  make([]bool, _m_numFloors),
		},
		Node: _node,
	}
	doorOpenTimer := time.NewTimer(_doorTimeOut)

	elevio.SetDoorOpenLamp(false)
	elevio.SetMotorDirection(elevator.Dir)
	for {
		switch elevator.Behavior {
		case pt.EB_Idle:
			fmt.Println("driver: idle, MD:", elevator.Dir, "Floor:", elevator.Floor, "Order:", elevator.Order)
			select {
			case elevator.Order = <-orderChan:
				orderAckChan <- pt.OrderAck{
					Order: elevator.Order,
					Node:  _node,
				}
				elevator = requestsChooseDirection(elevator)
				elevio.SetMotorDirection(elevator.Dir)
				elevio.SetDoorOpenLamp(elevator.Behavior == pt.EB_DoorOpen)
				if !doorOpenTimer.Stop() {
					<-doorOpenTimer.C
				}
				doorOpenTimer.Reset(_doorTimeOut) // does not hurt to reset, even if not going to door open state
				elevatorChan <- elevator
			case obstructed = <-obstructionSwitch:
			case elevator.Floor = <-floorUpdate: // maybe draining channel required, but just PollingGoroutine should spin
				// if shouldStop(elevator) {
				// 	elevio.SetMotorDirection(elevio.MD_Stop) // immediate stop before making further/complex calcs to avoid missing the floor/sensor and eventually running out of range
				// 	elevator = requestsChooseDirection(elevator)
				// 	// elevio.SetMotorDirection(elevator.Dir)
				// 	elevio.SetDoorOpenLamp(elevator.Behavior == pt.EB_DoorOpen)
				// 	if !doorOpenTimer.Stop() {
				// 		<-doorOpenTimer.C
				// 	} else {
				// 		// elevio.SetMotorDirection(elevator.Dir)
				// 	}
				// 	doorOpenTimer.Reset(_doorTimeOut)
				// }
				// elevio.SetFloorIndicator(elevator.Floor)
				// elevatorChan <- elevator
			}
		case pt.EB_Moving:
			fmt.Println("driver: moving, MD:", elevator.Dir, "Floor:", elevator.Floor, "Order:", elevator.Order)
			select {
			case elevator.Order = <-orderChan:
				orderAckChan <- pt.OrderAck{
					Order: elevator.Order,
					Node:  _node,
				} // do nothing with updated order until reached next floor
			case elevator.Floor = <-floorUpdate:
				//elevator.Floor = newFloor
				if shouldStop(elevator) {
					elevio.SetMotorDirection(elevio.MD_Stop) // immediate stop before making further/complex calcs to avoid missing the floor/sensor and eventually running out of range
					elevator = requestsChooseDirection(elevator)
					// elevio.SetMotorDirection(elevator.Dir)
					elevio.SetDoorOpenLamp(elevator.Behavior == pt.EB_DoorOpen)
					if !doorOpenTimer.Stop() {
						<-doorOpenTimer.C
					}
					doorOpenTimer.Reset(_doorTimeOut)
				} else {
					// elevio.SetMotorDirection(elevator.Dir)
				}
				elevio.SetFloorIndicator(elevator.Floor)
				elevatorChan <- elevator
			case obstructed = <-obstructionSwitch:
			}
		case pt.EB_DoorOpen:
			fmt.Println("driver: door open, MD:", elevator.Dir, "Floor:", elevator.Floor, "Order:", elevator.Order)
			select {
			case <-doorOpenTimer.C:
				doorTimeElapsed = true // to ignore obstruction until door like to close

				elevator, clearDown, clearUp, clearCab = clearCurrentFloor(elevator)
				floorServiced := pt.FloorServiced{
					Floor: elevio.ButtonEvent{
						Floor: elevator.Floor,
					},
					Node: _node,
				}
				if clearDown {
					elevator.Order.HallRequests[elevator.Floor][int(elevio.BT_HallDown)] = false
					clearDown = false
					floorServiced.Floor.Button = elevio.BT_HallDown
					floorServicedChan <- floorServiced
				}
				if clearUp {
					elevator.Order.HallRequests[elevator.Floor][int(elevio.BT_HallUp)] = false
					clearUp = false
					floorServiced.Floor.Button = elevio.BT_HallUp
					floorServicedChan <- floorServiced
				}
				if clearCab {
					elevator.Order.CabRequests[elevator.Floor] = false
					clearCab = false
					floorServiced.Floor.Button = elevio.BT_Cab
					floorServicedChan <- floorServiced
				}
				if !obstructed {
					// elevio.SetDoorOpenLamp(false) // TODO: move under decision and set depending on ouput of requestsChooseDirection
					elevator = requestsChooseDirection(elevator)
					elevio.SetMotorDirection(elevator.Dir)
					elevio.SetDoorOpenLamp(elevator.Behavior == pt.EB_DoorOpen)
					doorOpenTimer.Reset(_doorTimeOut)
					elevatorChan <- elevator
				} else {
					obstructedChan <- true // TODO: WONTFIX: consider moving out of else branch by just fwding value`obstructedChan <- obstructed`
				}
			case obstructed = <-obstructionSwitch:
				if doorTimeElapsed && !obstructed {
					doorTimeElapsed = false
					obstructedChan <- false
					elevator = requestsChooseDirection(elevator)
					elevio.SetMotorDirection(elevator.Dir)
					elevio.SetDoorOpenLamp(elevator.Behavior == pt.EB_DoorOpen)
					if !doorOpenTimer.Stop() {
						<-doorOpenTimer.C
					}
					doorOpenTimer.Reset(_doorTimeOut)
					elevatorChan <- elevator
				}
			case elevator.Order = <-orderChan:
				orderAckChan <- pt.OrderAck{
					Order: elevator.Order,
					Node:  _node,
				} // do nothing with updated order until doorTimeElapsed and no obstruction
			case elevator.Floor = <-floorUpdate: // maybe draining channel required, but just PollingGoroutine should spin
			}
		}
	}
}


// This is a whole new way to implement the State Machine: 










