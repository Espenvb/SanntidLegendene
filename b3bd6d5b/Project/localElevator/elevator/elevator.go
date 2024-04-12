package elevator

import (
	"Project/localElevator/elevio"
	"fmt"
)

type Behaviour int

const (
	Idle     Behaviour = 0
	DoorOpen Behaviour = 1
	Moving   Behaviour = 2
)

type ClearRequestVariant int

const (
	CV_All    ClearRequestVariant = 0
	CV_InDirn ClearRequestVariant = 1
)

type Config struct {
	ClearRequestVariant ClearRequestVariant
	DoorOpenDuration    float64
	MotorStopDuration   float64
}

type Elevator struct {
	Floor             int
	Dirn              elevio.MotorDirection
	Requests          [][]bool
	State             Behaviour
	ObstructionActive bool
	Config            Config
}

func EbToString(elevBehave Behaviour) string {
	switch elevBehave {
	case Idle:
		return "idle"
	case DoorOpen:
		return "doorOpen"
	case Moving:
		return "moving"
	default:
		return "UNDEFINED"
	}
}

func DirnToString(dir elevio.MotorDirection) string {
	switch dir {
	case elevio.Up:
		return "up"
	case elevio.Down:
		return "down"
	case elevio.Stop:
		return "stop"
	default:
		return "UNDEFINED"
	}
}

func ButtonToString(btn elevio.ButtonType) string {
	switch btn {
	case elevio.HallUp:
		return "HallUp"
	case elevio.HallDown:
		return "HallDown"
	case elevio.Cab:
		return "Cab"
	default:
		return "UNDEFINED"
	}
}

func InitElev() Elevator {
	requests := make([][]bool, 0)
	for floor := 0; floor < elevio.N_FLOORS; floor++ {
		requests = append(requests, make([]bool, elevio.N_BUTTONS))
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			requests[floor][btn] = false
		}
	}
	return Elevator{
		Floor:             -1,
		Dirn:              elevio.Stop,
		Requests:          requests,
		State:             Idle,
		ObstructionActive: false,
		Config: Config{
			ClearRequestVariant: CV_InDirn,
			DoorOpenDuration:    3.0,
			MotorStopDuration:   10.0,
		},
	}
}

func GetCabRequests(elev Elevator) []bool {
	var cabRequests []bool
	for floor := 0; floor < elevio.N_FLOORS; floor++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			if elevio.ButtonType(btn) == elevio.Cab {
				cabRequests = append(cabRequests, elev.Requests[floor][btn])
			}
		}
	}
	return cabRequests
}


func ElevatorPrint(es Elevator) {
	fmt.Printf(" +--------------------+\n")
	fmt.Printf(
		" |floor = %-2d        |\n"+
			"  |dirn  = %-12.12s|\n"+
			"  |behav = %-12.12s|\n",
		es.Floor,
		DirnToString(es.Dirn),
		EbToString(es.State),
	)
	fmt.Printf(" +--------------------+\n")
	fmt.Printf("  |  | up  | dn  | cab |\n")
	for f := elevio.N_FLOORS - 1; f >= 0; f-- {
		fmt.Printf("  | %d", f)
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			if (f == elevio.N_FLOORS-1 && btn == int(elevio.HallUp)) || (f == 0 && btn == int(elevio.HallDown)) {
				fmt.Printf("|     ")
			} else {
				if es.Requests[f][btn] {
					fmt.Printf("|  #  ")
				} else {
					fmt.Printf("|  -  ")
				}
			}
		}
		fmt.Printf("|\n")
	}
	fmt.Printf(" +--------------------+\n")
}
