package hra

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"project/elevio"
	pt "project/project_types"
	"reflect"
)

// Struct members must be public in order to be accessible by json.Marshal/.Unmarshal
// This means they must start with a capital letter, so we need to use field renaming struct tags to make them camelCase

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

const (
	DIR_Up   string = "up"
	DIR_Down string = "down"
	DIR_Stop string = "stop"
)

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

func convertMDtoString(md elevio.MotorDirection) string {
	switch md {
	case elevio.MD_Up:
		return DIR_Up
	case elevio.MD_Down:
		return DIR_Down
	case elevio.MD_Stop:
		return DIR_Stop
	default:
		panic("hra: convertMDtoString: md not a valid elevio.MotorDirection")
	}
}

func convertHRAOutput(
	hraOutput map[string][][2]bool,
	systemOrder pt.SystemOrder,
	elevatorStates map[pt.Node]pt.Elevator,
) (
	orders map[pt.Node]pt.Order,
) {
	orders = make(map[pt.Node]pt.Order)
	for id, hallRequests := range hraOutput {
		// get node from id
		//node := findNodeById(id, elevatorStates)
		node := pt.StringToNode(id)
		cabRequests := systemOrder.CabRequests[node.String()]
		order := pt.Order{
			HallRequests: hallRequests,
			CabRequests:  cabRequests,
		}
		orders[node] = order
	}
	return orders
}

func findNodeById(id string, elevatorStates map[pt.Node]pt.Elevator) pt.Node {
	var ip string
	var port int
	fmt.Sscanf(id, "%s:%d", &ip, &port)
	for node := range elevatorStates {
		if node.IP == ip && node.Port == port {
			return node
		}
	}
	panic(fmt.Sprintln("hra: findNodeById: 404 node ", id, " not found, elevatorState:", elevatorStates))
}

func printHRAOutput(hraOutput map[string][][2]bool) {
	fmt.Printf("hraAutput:\n")
	for k, v := range hraOutput {
		fmt.Printf("%6v :  %+v\n", k, v)
	}
}

// TODO use deep copy or remove
func removeLostElevators(es map[pt.Node]pt.Elevator, nu pt.NodeUpdate) map[pt.Node]pt.Elevator {
	emptyElevator := pt.Elevator{}
	for _, node := range nu.AliveNodes {
		if reflect.DeepEqual(es[node], emptyElevator) {
			delete(es, node)
		}
	}
	return es
}

func Assigner(
	hraExecutablePath string,
	systemOrder pt.SystemOrder,
	elevatorStatesIn map[pt.Node]pt.Elevator,
	aliveNodes pt.NodeUpdate,
) (
	orders map[pt.Node]pt.Order,
	newSystemOrder pt.SystemOrder,
	elevatorStates map[pt.Node]pt.Elevator, // even though pointer type, here explicit return to increase readability
) {
	newSystemOrder = systemOrder
	// copy(newSystemOrder.HallRequests, systemOrder.HallRequests)
	elevatorStates = elevatorStatesIn

	// set when primary has seen the elevators state never before although its node is alive
	guessedInitElevState := pt.Elevator{
		Floor:    1,
		Dir:      elevio.MD_Up,
		Behavior: pt.EB_Moving,
	}

	hraStates := make(map[string]HRAElevState, len(aliveNodes.AliveNodes))

	for _, node := range aliveNodes.AliveNodes {
		if _, ok := systemOrder.CabRequests[node.String()]; !ok {
			systemOrder.CabRequests[node.String()] = pt.EmptyCabRequests(len(systemOrder.HallRequests))
		}
		if _, ok := elevatorStates[node]; !ok { // meaning new node, never seen state before
			elevatorStates[node] = guessedInitElevState
		}
		e := elevatorStates[node]
		hraElevState := HRAElevState{
			Behavior:    string(e.Behavior),
			Floor:       e.Floor,
			Direction:   convertMDtoString(e.Dir),
			CabRequests: systemOrder.CabRequests[node.String()],
		}
		//id := fmt.Sprintf("%s:%d", node.IP, node.Port)
		// hraStates[id] = hraElevState
		hraStates[node.String()] = hraElevState
	}

	hraInput := HRAInput{
		HallRequests: systemOrder.HallRequests,
		States:       hraStates,
	}
	// Example input:
	// input := HRAInput{
	//     HallRequests: [][2]bool{{false, false}, {true, false}, {false, false}, {false, true}},
	//     States: map[string]HRAElevState{
	//         "IP1:Port1": HRAElevState{
	//             Behavior:       "moving",
	//             Floor:          2,
	//             Direction:      "up",
	//             CabRequests:    []bool{false, false, false, true},
	//         },
	//         "IP2:Port2": HRAElevState{
	//             Behavior:       "idle",
	//             Floor:          0,
	//             Direction:      "stop",
	//             CabRequests:    []bool{false, false, false, false},
	//         },
	//     },
	// }

	jsonBytes, err := json.Marshal(hraInput)
	if err != nil {
		panic(fmt.Sprintf("json.Marshal error: %s", err.Error()))
	}

	// TODO: take flag input for path instead
	// ret, err := exec.Command("../Project-resources/cost_fns/hall_request_assigner/"+"hall_request_assigner", "-i", string(jsonBytes)).CombinedOutput()
	ret, err := exec.Command(hraExecutablePath, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("exec.Command error: %s", err.Error()))
	}

	hraOutput := new(map[string][][2]bool)
	err = json.Unmarshal(ret, &hraOutput)
	if err != nil {
		panic(fmt.Sprintf("json.Unmarshal error: %s", err.Error()))
	}

	orders = convertHRAOutput(*hraOutput, systemOrder, elevatorStates)

	// Debugging
	//printHRAOutput(*hraOutput)
	//fmt.Println(orders)

	return orders, newSystemOrder, elevatorStates
}
