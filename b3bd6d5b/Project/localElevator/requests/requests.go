package requests

import (
	"Project/localElevator/elevator"
	"Project/localElevator/elevio"
)

type DirnBehaviourPair struct {
	Dirn  elevio.MotorDirection
	State elevator.Behaviour
}

func requestsAbove(elev elevator.Elevator) bool {
	for floor := elev.Floor + 1; floor < elevio.N_FLOORS; floor++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			if elev.Requests[floor][btn] {
				return true
			}
		}
	}
	return false
}

func requestsBelow(elev elevator.Elevator) bool {
	for floor := 0; floor < elev.Floor; floor++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			if elev.Requests[floor][btn] {
				return true
			}
		}
	}
	return false
}

func requestsHere(elev elevator.Elevator) bool {
	for btn := 0; btn < elevio.N_BUTTONS; btn++ {
		if elev.Requests[elev.Floor][btn] {
			return true
		}
	}
	return false
}

func ChooseDirection(elev elevator.Elevator) DirnBehaviourPair {
	switch elev.Dirn {
	case elevio.Up:
		if requestsAbove(elev) {
			return DirnBehaviourPair{elevio.Up, elevator.Moving}
		} else if requestsHere(elev) {
			return DirnBehaviourPair{elevio.Down, elevator.DoorOpen}
		} else if requestsBelow(elev) {
			return DirnBehaviourPair{elevio.Down, elevator.Moving}
		} else {
			return DirnBehaviourPair{elevio.Stop, elevator.Idle}
		}
	case elevio.Down:
		if requestsBelow(elev) {
			return DirnBehaviourPair{elevio.Down, elevator.Moving}
		} else if requestsHere(elev) {
			return DirnBehaviourPair{elevio.Up, elevator.DoorOpen}
		} else if requestsAbove(elev) {
			return DirnBehaviourPair{elevio.Up, elevator.Moving}
		} else {
			return DirnBehaviourPair{elevio.Stop, elevator.Idle}
		}
	case elevio.Stop:
		if requestsHere(elev) {
			return DirnBehaviourPair{elevio.Stop, elevator.DoorOpen}
		} else if requestsAbove(elev) {
			return DirnBehaviourPair{elevio.Up, elevator.Moving}
		} else if requestsBelow(elev) {
			return DirnBehaviourPair{elevio.Down, elevator.Moving}
		} else {
			return DirnBehaviourPair{elevio.Stop, elevator.Idle}
		}
	default:
		return DirnBehaviourPair{elevio.Stop, elevator.Idle}
	}
}

func ShouldStop(elev elevator.Elevator) bool {
	switch elev.Dirn {
	case elevio.Down:
		return elev.Requests[elev.Floor][elevio.HallDown] || elev.Requests[elev.Floor][elevio.Cab] || !requestsBelow(elev)
	case elevio.Up:
		return elev.Requests[elev.Floor][elevio.HallUp] || elev.Requests[elev.Floor][elevio.Cab] || !requestsAbove(elev)
	default:
		return true
	}
}

func ShouldClearRequest(elev elevator.Elevator) []bool {
	return elev.Requests[elev.Floor]
}

func ShouldClearImmediately(
	elev elevator.Elevator,
	btn_floor int,
	btn_type elevio.ButtonType,
) bool {
	switch elev.Config.ClearRequestVariant {
	case elevator.CV_All:
		return elev.Floor == btn_floor
	case elevator.CV_InDirn:
		return elev.Floor == btn_floor && ((elev.Dirn == elevio.Up && btn_type == elevio.HallUp) ||
			(elev.Dirn == elevio.Down && btn_type == elevio.HallDown) ||
			elev.Dirn == elevio.Stop || btn_type == elevio.Cab)
	default:
		return false
	}
}

func ClearAtCurrentFloor(elev elevator.Elevator) (elevator.Elevator, [2]bool) {
	removingHallButtons := [2]bool{false, false}
	switch elev.Config.ClearRequestVariant {
	case elevator.CV_All:
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			elev.Requests[elev.Floor][btn] = false
		}
	case elevator.CV_InDirn:
		elev.Requests[elev.Floor][elevio.Cab] = false
		switch elev.Dirn {
		case elevio.Up:
			if !requestsAbove(elev) && !elev.Requests[elev.Floor][elevio.HallUp] {
				elev.Requests[elev.Floor][elevio.HallDown] = false
				removingHallButtons[elevio.HallDown] = true
			}
			elev.Requests[elev.Floor][elevio.HallUp] = false
			removingHallButtons[elevio.HallUp] = true
		case elevio.Down:
			if !requestsBelow(elev) && !elev.Requests[elev.Floor][elevio.HallDown] {
				elev.Requests[elev.Floor][elevio.HallUp] = false
				removingHallButtons[elevio.HallUp] = true
			}
			elev.Requests[elev.Floor][elevio.HallDown] = false
			removingHallButtons[elevio.HallDown] = true
		case elevio.Stop:
		default:
			elev.Requests[elev.Floor][elevio.HallUp] = false
			elev.Requests[elev.Floor][elevio.HallDown] = false
			removingHallButtons[elevio.HallUp] = true
			removingHallButtons[elevio.HallDown] = true
		}
	default:
	}
	return elev, removingHallButtons
}
