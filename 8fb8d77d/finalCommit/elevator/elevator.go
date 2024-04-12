package elevator

import (
	"Heis/driver-go/elevio"
)

type ElevatorBehaviour int

const (
	EB_Idle     ElevatorBehaviour = 0
	EB_DoorOpen ElevatorBehaviour = 1
	EB_Moving   ElevatorBehaviour = 2
)

type Elevator struct {
	Floor     int
	Dirn      elevio.MotorDirection
	Requests  [elevio.NumFloors][elevio.NumButtons]bool
	Behaviour ElevatorBehaviour
	ElevID    string
	Failure   bool
}

func SetAllLights(elev Elevator, mapOfElevs map[string]Elevator) {
	HallCalls := OrHallCalls(mapOfElevs)

	for f := 0; f < elevio.NumFloors; f++ {
		for b := 0; b < elevio.NumButtons-1; b++ {
			elevio.SetButtonLamp(elevio.ButtonType(b), f, HallCalls[f][b])
		}
		elevio.SetButtonLamp(elevio.BT_Cab, f, elev.Requests[f][elevio.BT_Cab])
	}
	elevio.SetDoorOpenLamp(elev.Behaviour == EB_DoorOpen)

}

func OrHallCalls(allElevs map[string]Elevator) [][2]bool {
	result := make([][2]bool, elevio.NumFloors)
	for _, elev := range allElevs {
		for j, row := range getHallCalls(elev) {
			result[j][0] = result[j][0] || row[0]
			result[j][1] = result[j][1] || row[1]
		}
	}
	return result
}

func GetCabCalls(elev Elevator) []bool {
	cabRequests := []bool{false, false, false, false}
	for f := 0; f < elevio.NumFloors; f++ {
		cabRequests[f] = elev.Requests[f][elevio.NumButtons-1]
	}
	return cabRequests
}

func getHallCalls(elev Elevator) [][2]bool {
	HallCalls := [][2]bool{{false, false},
		{false, false},
		{false, false},
		{false, false}}

	for f := 0; f < elevio.NumFloors; f++ {
		for b := 0; b < (elevio.NumButtons - 1); b++ {
			HallCalls[f][b] = elev.Requests[f][b]
		}
	}
	return HallCalls
}

func MergeHallAndCabCall(cabs []bool, halls [][2]bool) [elevio.NumFloors][elevio.NumButtons]bool {
	requests := [elevio.NumFloors][elevio.NumButtons]bool{{false, false, false},
		{false, false, false},
		{false, false, false},
		{false, false, false}}
	for f := 0; f < elevio.NumFloors; f++ {
		for b := 0; b < (elevio.NumButtons - 1); b++ {
			requests[f][b] = halls[f][b]
		}
		requests[f][elevio.NumButtons-1] = cabs[f]
	}
	return requests
}
