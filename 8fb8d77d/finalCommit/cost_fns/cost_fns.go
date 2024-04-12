package cost_fns

import (
	"Heis/driver-go/elevio"
	"Heis/elevator"
	"encoding/json"
	"fmt"
	"os/exec"
)

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

func RunCostFunc(elevMap map[string]elevator.Elevator) map[string]elevator.Elevator {
	commonHallCalls := elevator.OrHallCalls(elevMap)
	tempElevMap := elevMap
	for k, v := range tempElevMap {
		if v.Failure {
			fmt.Println("CostFunc: Elevator ", k, " has failure")
			delete(tempElevMap, k)
		} else if v.Floor == -1 {
			fmt.Println("CostFunc: Elevator ", k, " has floor -1")
			delete(tempElevMap, k)
		}
	}
	input := inputToCost(commonHallCalls, tempElevMap)
	newHRAs := getCostOutput(input)
	for k := range newHRAs {
		elevMap[k] = mergeHallAndRequests(elevMap[k], newHRAs[k])
	}
	return elevMap
}

func elevToHRAElevState(elev elevator.Elevator) HRAElevState {
	return HRAElevState{
		Behavior:    elevBehaviourToString(elev.Behaviour),
		Floor:       elev.Floor,
		Direction:   motorDirnToString(elev.Dirn),
		CabRequests: elevator.GetCabCalls(elev),
	}
}

func inputToCost(commonHallCalls [][2]bool, elevMap map[string](elevator.Elevator)) HRAInput {
	stateMap := map[string]HRAElevState{}
	for k, v := range elevMap {
		stateMap[k] = elevToHRAElevState(v)
	}
	input := HRAInput{
		HallRequests: commonHallCalls,
		States:       stateMap,
	}
	return input
}

func hra_funcs(input HRAInput, output *map[string][][2]bool, hraExecutable string) {

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
		return
	}

	ret, err := exec.Command("./hall_request_assigner/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
		return
	}

	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		return
	}

}
func getCostOutput(input HRAInput) map[string][][2]bool {
	hraExecutable := "hall_request_assigner"
	output := new(map[string][][2]bool)

	hra_funcs(input, output, hraExecutable)

	return *output
}

func mergeHallAndRequests(elev elevator.Elevator, halls [][2]bool) elevator.Elevator {
	for f := 0; f < elevio.NumFloors; f++ {
		for b := 0; b < elevio.NumButtons-1; b++ {
			elev.Requests[f][b] = halls[f][b]
		}
	}
	return elev
}

func elevBehaviourToString(elevBehaviour elevator.ElevatorBehaviour) string {
	switch elevBehaviour {
	case 0:
		return "idle"
	case 1:
		return "doorOpen"
	case 2:
		return "moving"
	default:
		return "Unknown"
	}
}

func motorDirnToString(elevDirn elevio.MotorDirection) string {
	switch elevDirn {
	case 0:
		return "stop"
	case 1:
		return "up"
	case -1:
		return "down"
	default:
		return "Unknown"
	}
}
