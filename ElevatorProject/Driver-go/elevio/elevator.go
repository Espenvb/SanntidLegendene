package elevio



//"fmt"



type FsmState int

const (
	Idle	FsmState = 0
	moving  	  = 1
	DoorOpen 	  = 2
)

type Elevator struct{
	floor int
	direction MotorDirection
	destination int
	state 	FsmState

}

func Initilize_elevator(d MotorDirection) Elevator{
	Elevator1:= Elevator{
		floor: GetFloor(),
		state: FsmState(0),
		direction: d,
		
	}

	return Elevator1

}




