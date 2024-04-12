package localElevatorHandler

import (
	"Project/localElevator/elevator"
	"Project/localElevator/elevio"
	"Project/localElevator/requests"
	"Project/network/messages"
	"Project/network/tcp"
	"Project/network/udpBroadcast/udpNetwork/localip"
	"net"
	"time"
)


var motorStopTimer *time.Timer
var doorOpenTimer *time.Timer
var elev elevator.Elevator = elevator.InitElev()

var hraElevState = messages.HRAElevState{
	Behaviour:   elevator.EbToString(elev.State),
	Floor:       elev.Floor,
	Direction:   elevator.DirnToString(elev.Dirn),
	CabRequests: elevator.GetCabRequests(elev),
}

var msgElevState = messages.ElevStateMsg{
	IpAddr:    "",
	ElevState: hraElevState,
}

func LocalElevatorHandler(
	buttonsCh chan elevio.ButtonEvent, 
	floorsCh chan int, obstrCh chan bool, 
	masterConnCh chan net.Conn, 
	toNetworkCh chan tcp.SendNetworkMsg, 
	toFSMCh chan []byte, 
	visibleOnNetwork chan bool,
) {
	masterConn := <-masterConnCh
	localip, _ := localip.LocalIP()
	msgElevState.IpAddr = localip
	motorStopTimer = time.NewTimer(24 * time.Hour)
	doorOpenTimer = time.NewTimer(24 * time.Hour)
	for {
		select {
		case button := <-buttonsCh:
			if button.Button == elevio.Cab {
				msgElevState.ElevState.CabRequests[button.Floor] = true
				sendingBytes := messages.PackMessage(messages.MsgElevState, msgElevState)
				toNetworkCh <- tcp.SendNetworkMsg{masterConn, sendingBytes}
				OnRequestButtonPress(button.Floor, button.Button, visibleOnNetwork, toNetworkCh, masterConn)
				elevio.SetButtonLamp(button.Floor, button.Button, true)
			} else {
				hallReq := messages.HallReqMsg{true, button.Floor, button.Button}
				sendingBytes := messages.PackMessage(messages.MsgHallReq, hallReq)
				toNetworkCh <- tcp.SendNetworkMsg{masterConn, sendingBytes}
			}
		case floor := <-floorsCh:
			removingHallButtons := OnFloorArrival(floor, visibleOnNetwork)
			msgElevState.ElevState.Floor = floor
			msgElevState.ElevState.CabRequests[floor] = false
			sendingBytes := messages.PackMessage(messages.MsgElevState, msgElevState)
			toNetworkCh <- tcp.SendNetworkMsg{masterConn, sendingBytes}
			sendHallBtnRemovalToConn(removingHallButtons, floor, toNetworkCh, masterConn)
		case obstr := <-obstrCh:
			elev.ObstructionActive = obstr
		case msgFromNetwork := <-toFSMCh:
			msgType, data := messages.UnpackMessage(msgFromNetwork)
			switch msgType {
			case messages.MsgAssignedHallReq:
				newHallRequests := data.([][2]bool)
				for floor := 0; floor < len(newHallRequests); floor++ {
					for hallIndex := 0; hallIndex < len(newHallRequests[hallIndex]); hallIndex++ {
						value := newHallRequests[floor][hallIndex]
						if value {
							hallButton := elevio.ButtonType(hallIndex)
							OnRequestButtonPress(floor, hallButton, visibleOnNetwork, toNetworkCh, masterConn)
						} else {
							elev.Requests[floor][hallIndex] = false
						}
					}
				}
			case messages.MsgHallLigths:
				floor := data.(messages.HallReqMsg).Floor
				button := data.(messages.HallReqMsg).Button
				addOrRemove :=data.(messages.HallReqMsg).TAddFRemove
				elevio.SetButtonLamp(floor, button, addOrRemove)
			case messages.MsgRestoreCabReq:
				for floor := 0; floor < len(data.([]bool)); floor++ {
					elev.Requests[floor][elevio.Cab] = data.([]bool)[floor]
				}
			}
		case newMasterConn := <-masterConnCh:
			masterConn = newMasterConn
		case <-doorOpenTimer.C:
			OnDoorTimeout(masterConn, toNetworkCh)
		case <-motorStopTimer.C:
			visibleOnNetwork <- false
			motorStopTimer.Reset(24 * time.Hour)
		}
	}
}

func InitLights() {
	elevio.SetDoorOpenLamp(false)
	SetAllLights(elev)
}

func SetAllLights(es elevator.Elevator) {
	for floor := 0; floor < elevio.N_FLOORS; floor++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			elevio.SetButtonLamp(floor, elevio.ButtonType(btn), es.Requests[floor][btn])
		}
	}
}

func SetAllCabLights(e elevator.Elevator) {
	for floor := 0; floor < elevio.N_FLOORS; floor++ {
		elevio.SetButtonLamp(floor, elevio.Cab, e.Requests[floor][elevio.Cab])
	}
}

func SetLight(btn elevio.ButtonType, floor int, onOff bool) {
	elevio.SetButtonLamp(floor, elevio.ButtonType(btn), onOff)
}

func OnInitBetweenFloors() {
	elevio.SetMotorDirection(elevio.Down)
	elev.Dirn = elevio.Down
	elev.State = elevator.Moving
}

func OnRequestButtonPress(
	btnFloor int, 
	btnType elevio.ButtonType, 
	visibleOnNetwork chan bool, 
	toNetworkCh chan tcp.SendNetworkMsg, 
	masterConn net.Conn,
) {
	switch elev.State {
	case elevator.DoorOpen:
		if requests.ShouldClearImmediately(elev, btnFloor, btnType) {
			doorOpenTimer.Reset(time.Duration(elev.Config.DoorOpenDuration) * time.Second)
		} else {
			elev.Requests[btnFloor][btnType] = true
		}
	case elevator.Moving:
		elev.Requests[btnFloor][btnType] = true
	case elevator.Idle:
		elev.Requests[btnFloor][btnType] = true
		var pair requests.DirnBehaviourPair = requests.ChooseDirection(elev)
		elev.Dirn = pair.Dirn
		elev.State = pair.State
		switch pair.State {
		case elevator.DoorOpen:
			elevio.SetDoorOpenLamp(true)
			doorOpenTimer.Reset(time.Duration(elev.Config.DoorOpenDuration) * time.Second)
			var removingHallButtons [2]bool
			elev, removingHallButtons = requests.ClearAtCurrentFloor(elev)
			sendHallBtnRemovalToConn(removingHallButtons, btnFloor, toNetworkCh, masterConn)
		case elevator.Moving:
			elevio.SetMotorDirection(elev.Dirn)
			motorStopTimer.Reset(time.Duration(elev.Config.MotorStopDuration) * time.Second)
			visibleOnNetwork <- true
		case elevator.Idle:
		}
	}
}

func OnFloorArrival(newFloor int, visibleOnNetwork chan bool) [2]bool {
	var removingHallButtons [2]bool
	elev.Floor = newFloor
	elevio.SetFloorIndicator(elev.Floor)
	motorStopTimer.Reset(time.Duration(elev.Config.MotorStopDuration) * time.Second)
	visibleOnNetwork <- true
	switch elev.State {
	case elevator.Moving:
		if requests.ShouldStop(elev) {
			elevio.SetMotorDirection(elevio.Stop)
			elevio.SetDoorOpenLamp(true)
			elev, removingHallButtons = requests.ClearAtCurrentFloor(elev)
			doorOpenTimer.Reset(time.Duration(elev.Config.DoorOpenDuration) * time.Second)
			SetAllCabLights(elev)
			elev.State = elevator.DoorOpen
		} else {
		}
	default:
		break
	}
	return removingHallButtons
}

func OnDoorTimeout(masterConn net.Conn, toNetworkCh chan tcp.SendNetworkMsg) {
	switch elev.State {
	case elevator.DoorOpen:
		if elev.ObstructionActive {
			doorOpenTimer.Reset(time.Duration(elev.Config.DoorOpenDuration)* time.Second)
			break
		}
		var pair requests.DirnBehaviourPair = requests.ChooseDirection(elev)
		elev.Dirn = pair.Dirn
		elev.State = pair.State
		switch elev.State {
		case elevator.DoorOpen:
			doorOpenTimer.Reset(time.Duration(elev.Config.DoorOpenDuration)* time.Second)
			var removingHallButtons [2]bool
			elev, removingHallButtons = requests.ClearAtCurrentFloor(elev)
			sendHallBtnRemovalToConn(removingHallButtons, elev.Floor, toNetworkCh, masterConn)
			SetAllCabLights(elev)
		case elevator.Moving:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elev.Dirn)
			motorStopTimer.Reset(time.Duration(elev.Config.MotorStopDuration)* time.Second)
		case elevator.Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elev.Dirn)
			doorOpenTimer.Reset(24 * time.Hour)
			motorStopTimer.Reset(24 * time.Hour)
		}
	default:
	}
}

func sendHallBtnRemovalToConn(
	removingHallButtons [2]bool, 
	floor int, 
	toNetworkCh chan tcp.SendNetworkMsg, 
	masterConn net.Conn,
) {
	for btnIndex, btnValue := range removingHallButtons {
		buttonType := elevio.ButtonType(btnIndex)
		if btnValue {
			removeHallReq := messages.HallReqMsg{false, floor, buttonType}
			sendingBytes := messages.PackMessage(messages.MsgHallReq, removeHallReq)
			toNetworkCh <- tcp.SendNetworkMsg{masterConn, sendingBytes}
		}
	}
}