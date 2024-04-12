package execution

import (
	"fmt"
	"project/elevio"
	pt "project/project_types"
	"time"
)


func DebugLightService(hallRequestsChan chan<- [][2]bool, cabRequestsChan chan<- []bool) {
	for {
		mockHallRequests := [][2]bool{{false, true}, {false, false}, {false, false}, {true, false}}
		mockCabRequests := []bool{true, false, false, false}
		hallRequestsChan <- mockHallRequests
		cabRequestsChan <- mockCabRequests
		time.Sleep(5 * time.Second)

		mockHallRequests = [][2]bool{{false, false}, {true, false}, {false, false}, {false, false}}
		mockCabRequests = []bool{false, false, false, true}
		hallRequestsChan <- mockHallRequests
		cabRequestsChan <- mockCabRequests
		time.Sleep(5 * time.Second)

		// trying to set a non existing button lamp, does not fail
		mockHallRequests = [][2]bool{{true, false}, {false, false}, {false, false}, {false, false}}
		mockCabRequests = []bool{false, false, false, true}
		hallRequestsChan <- mockHallRequests
		cabRequestsChan <- mockCabRequests
		time.Sleep(5 * time.Second)

		mockHallRequests = [][2]bool{{true, false}, {false, false}, {false, false}, {false, false}}
		//mockCabRequests = []bool{false, false, false, true, true} // resulting in a crash of the elevator server (range violation)
		hallRequestsChan <- mockHallRequests
		cabRequestsChan <- mockCabRequests
		time.Sleep(5 * time.Second)
	}
}

func DebugButtonPressForwarding(buttonPressChan <-chan elevio.ButtonEvent) {
	for {
		fmt.Println("ButtonPress:", <-buttonPressChan)
	}
}

func DebugExecutionFloor(orderChan chan<- pt.Order,
	floorUpdate <-chan int,
	obstruction <-chan bool,
	ownNodeChangedChan <-chan pt.Node,
	elevatorChan chan<- pt.Elevator,
	orderAckChan chan<- pt.OrderAck,
	floorServicedChan chan<- pt.FloorServiced) {
	for {
		mockOrder := pt.Order{
			HallRequests: [][2]bool{{false, true}, {false, false}, {false, false}, {true, false}},
			CabRequests: []bool{true, false, false, false},
		}
		orderChan <- mockOrder
		time.Sleep(5 * time.Second)
	}
}
