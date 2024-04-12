package elevator

import (
	"Heis/driver-go/elevio"
	"fmt"
	"os"
	"strings"
)

func InitElev() Elevator {

	return Elevator{
		Floor:     -1,
		Dirn:      elevio.MD_Stop,
		Requests:  initializeRequests(),
		Behaviour: EB_Idle,
		ElevID:    "",
		Failure:   false,
	}
}

func OnInitBetweenFloors(elev Elevator) Elevator {
	elevio.SetMotorDirection(elevio.MD_Down)
	elev.Behaviour = EB_Moving
	elev.Dirn = elevio.MD_Down
	return elev
}

func checkAndLoadCabCalls() []bool {
	fileInfo, err := os.Stat("localBackup.txt")
	if os.IsNotExist(err) || (fileInfo != nil && fileInfo.Size() == 0) {
		cabCalls := make([]bool, elevio.NumFloors)
		for i := range cabCalls {
			cabCalls[i] = false
		}
		return cabCalls
	} else {
		file, err := os.ReadFile("localBackup.txt")
		if err != nil {
			fmt.Println(err)
		}
		var cabCalls []bool
		fileContent := strings.TrimSpace(string(file))
		for _, char := range fileContent {
			if char == 't' {
				cabCalls = append(cabCalls, true)
			} else if char == 'f' {
				cabCalls = append(cabCalls, false)
			}
		}
		return cabCalls
	}
}

func initializeRequests() [elevio.NumFloors][elevio.NumButtons]bool {
	initializeCabCalls := checkAndLoadCabCalls()
	requests := [elevio.NumFloors][elevio.NumButtons]bool{}

	for floor := range requests {
		requests[floor][0] = false
		requests[floor][1] = false
		requests[floor][2] = initializeCabCalls[floor]
	}
	return requests
}
