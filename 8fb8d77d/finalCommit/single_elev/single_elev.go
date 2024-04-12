package single_elev

import (
	"Heis/cost_fns"
	"Heis/driver-go/elevio"
	"Heis/elevator"
	"Heis/process_pair"
	"Heis/requests"
	"fmt"
	"reflect"
)

func ButtonsAndRequests(elevatorID string, elevUpdateRealtimeCh <-chan elevator.Elevator,
	drv_buttons chan elevio.ButtonEvent, sendMapToSlavesCh chan<- map[string]elevator.Elevator, getElevFromSlave chan elevator.Elevator,
	receiveMapFromMasterCh <-chan map[string]elevator.Elevator, newOrderCh chan<- map[string]elevator.Elevator,
	lightsCh <-chan int, sendMyselfToMaster chan elevator.Elevator, isMasterCh chan bool) {

	isMaster := false
	elev := elevator.InitElev()
	elev.ElevID = elevatorID
	mapOfElevs := make(map[string]elevator.Elevator)
	mapOfElevs[elev.ElevID] = elev

	for {
		select {
		case isMaster = <-isMasterCh:
		case a := <-elevUpdateRealtimeCh:
			elev = a
			process_pair.WriteToLocalBackup(elevator.GetCabCalls(elev))
			if isMaster {
				mapOfElevs[elev.ElevID] = elev
			} else {
				sendMyselfToMaster <- elev
			}

		case a := <-drv_buttons:
			btn_floor := a.Floor
			btn_type := a.Button
			fmt.Printf("Button: %+v\n", a)
			elev.Requests[btn_floor][btn_type] = true

			if isMaster {
				mapOfElevs[elev.ElevID] = elev
				mapOfElevs := cost_fns.RunCostFunc(mapOfElevs)
				sendMapToSlavesCh <- mapOfElevs
				newOrderCh <- mapOfElevs
			} else {
				sendMyselfToMaster <- elev
			}

		case a := <-getElevFromSlave:
			if isMaster {
				tempMap := copyMap(mapOfElevs)
				mapOfElevs[a.ElevID] = a
				mapOfElevs = cost_fns.RunCostFunc(mapOfElevs)
				if !reflect.DeepEqual(tempMap, mapOfElevs) {
					sendMapToSlavesCh <- mapOfElevs
					newOrderCh <- mapOfElevs
				}
			}

		case a := <-receiveMapFromMasterCh:
			if !isMaster {
				mapOfElevs = copyMap(a)
				elev = mapOfElevs[elev.ElevID]
				newOrderCh <- mapOfElevs
			}
		case <-lightsCh:
			elevator.SetAllLights(elev, mapOfElevs)

		}
	}
}

func OrderExecution(elevatorId string, elevUpdateRealtimeCh chan<- elevator.Elevator,
	drv_floors chan int, newOrderCh <-chan map[string]elevator.Elevator, doorTimerCh chan bool,
	timedOut chan int, lightsCh chan<- int) {
	elev := elevator.InitElev()
	elev.ElevID = elevatorId

	if elevio.GetFloor() == -1 {
		fmt.Println("Started at an invalid floor, moving down")
		elev = elevator.OnInitBetweenFloors(elev)
		elevUpdateRealtimeCh <- elev
	}

	for {
		select {
		case a := <-newOrderCh:
			elev = a[elev.ElevID]
			if (elev.Behaviour != elevator.EB_Moving) && requests.ShouldClearImmediately(elev) {
				elev.Behaviour = elevator.EB_DoorOpen
				doorTimerCh <- true
				elevUpdateRealtimeCh <- elev
			} else if elev.Behaviour == elevator.EB_Idle {
				pair := requests.ChooseDirection(elev)
				elev.Dirn = pair.Dirn
				elevio.SetMotorDirection(elev.Dirn)
				elev.Behaviour = pair.Behaviour
				elevUpdateRealtimeCh <- elev
			}
			lightsCh <- 1

		case a := <-drv_floors:
			fmt.Printf("Floor: %+v\n", a)
			elev.Floor = a
			elevUpdateRealtimeCh <- elev
			elevio.SetFloorIndicator(elev.Floor)
			switch elev.Behaviour {
			case elevator.EB_Moving:
				if requests.ShouldStop(elev) {
					elevio.SetMotorDirection(elevio.MD_Stop)
					elev.Behaviour = elevator.EB_DoorOpen
					elevUpdateRealtimeCh <- elev
					lightsCh <- 1
					doorTimerCh <- true
				}
			default:

			}
		case <-timedOut:
			fmt.Println("Timer timed out")
			elev = requests.OnDoorTimeout(elev, doorTimerCh, lightsCh, elevUpdateRealtimeCh)

		}
	}
}

func copyMap(source map[string]elevator.Elevator) map[string]elevator.Elevator {
	newMap := make(map[string]elevator.Elevator, len(source))
	for k, v := range source {
		newMap[k] = v
	}
	return newMap
}
