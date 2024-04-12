package execution

import (
	"project/elevio"
	pt "project/project_types"
)

func updateLights(lastLightOrder pt.LightOrder, lightOrder pt.LightOrder) {
	for i := 0; i < _m_numFloors; i++ {
		elevio.SetButtonLamp(elevio.BT_Cab, i, lightOrder.CabRequests[i])
		elevio.SetButtonLamp(elevio.BT_HallUp, i, lightOrder.HallRequests[i][elevio.BT_HallUp])
		elevio.SetButtonLamp(elevio.BT_HallDown, i, lightOrder.HallRequests[i][elevio.BT_HallDown])
	}
}

func updateLightsDelta(lastLightOrder pt.LightOrder, lightOrder pt.LightOrder) {
	for i := 0; i < _m_numFloors; i++ {
		if lastLightOrder.CabRequests[i] && !lightOrder.CabRequests[i] {
			elevio.SetButtonLamp(elevio.BT_Cab, i, false)
		} else if !lastLightOrder.CabRequests[i] && lightOrder.CabRequests[i] {
			elevio.SetButtonLamp(elevio.BT_Cab, i, true)
		}

		if lastLightOrder.HallRequests[i][elevio.BT_HallDown] && !lightOrder.HallRequests[i][elevio.BT_HallDown] {
			elevio.SetButtonLamp(elevio.BT_HallDown, i, false)
		} else if !lastLightOrder.HallRequests[i][elevio.BT_HallDown] && lightOrder.HallRequests[i][elevio.BT_HallDown] {
			elevio.SetButtonLamp(elevio.BT_HallDown, i, true)
		}

		if lastLightOrder.HallRequests[i][elevio.BT_HallUp] && !lightOrder.HallRequests[i][elevio.BT_HallUp] {
			elevio.SetButtonLamp(elevio.BT_HallUp, i, false)
		} else if !lastLightOrder.HallRequests[i][elevio.BT_HallUp] && lightOrder.HallRequests[i][elevio.BT_HallUp] {
			elevio.SetButtonLamp(elevio.BT_HallUp, i, true)
		}
	}
}

func ButtonLightService(
	lightOrderChan <-chan pt.LightOrder,
	lightOrderAckChan chan<- pt.LightOrderAck,
) {
	if !_initialized {
		panic("trying to start un-initialized execution package!")
	}

	lastLightOrder := pt.LightOrder{
		HallRequests: pt.EmptyHallRequests(_m_numFloors),
		CabRequests:  pt.EmptyCabRequests(_m_numFloors),
	}
	for {
		select {
		case lightOrder := <-lightOrderChan:
			// fmt.Println("ButtonLightService: received lightOrder", lightOrder)
			lightOrderAckChan <- pt.LightOrderAck{
				LightOrder: lightOrder,
				Node:       _node,
			}
			updateLights(lastLightOrder, lightOrder)
			lastLightOrder = lightOrder
		}
	}
}

// TODO: forward buttonpress (elevio.ButtonEvent) to a channel ButtonPress

func ButtonPressForwarder(
	elevioButtonPressChan <-chan elevio.ButtonEvent,
	buttonPressChan chan<- pt.ButtonPress,
) {
	if !_initialized {
		panic("trying to start un-initialized execution package!")
	}
	for {
		select {
		case elevioButtonPress := <-elevioButtonPressChan:
			buttonPress := pt.ButtonPress{
				Button: elevioButtonPress,
				Node:   _node,
			}
			// fmt.Println("buttonPress:", buttonPress)
			buttonPressChan <- buttonPress
		}
	}
}
