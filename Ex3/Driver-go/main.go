package main

import "Driver-go/elevio"
import "fmt"


var d elevio.MotorDirection = elevio.MD_Up
var floor int
//var floor int

func main() {

	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	
	//elevio.SetMotorDirection(d)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	

	for {
		select {
		
		case a := <-drv_buttons:
			fmt.Printf("Detecten button pressed")
			elevio.SetButtonLamp(a.Button, a.Floor, true)
			if elevio.GetFloor() < a.Floor {
				//Go up
				fmt.Printf("Going up")
				d = elevio.MD_Up
				floor = a.Floor
			} else if elevio.GetFloor() > a.Floor {
				//Go down
				fmt.Printf(("Going down"))
				d = elevio.MD_Down
				floor = a.Floor
			}
			elevio.SetMotorDirection(d)
		
		case a := <-drv_floors:
			fmt.Printf("Detected change in flooor")
			fmt.Printf("%+v\n", a)
			if elevio.GetFloor() == floor{
				fmt.Printf("You have arrived stop")
				d = elevio.MD_Stop
				elevio.SetMotorDirection(d)
			}
			fmt.Println(elevio.GetFloor())
		}
	}
}
