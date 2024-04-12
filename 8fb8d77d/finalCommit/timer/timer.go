package timer

import (
	"Heis/driver-go/elevio"
	"time"
)

func Timer(doorTimerCh chan bool, timedOut chan int) {
	resetDoorTimer := false
	timer := time.NewTimer(3 * time.Second)
	for {
		if elevio.GetObstruction() {
			timer.Reset(3 * time.Second)
		}
		if resetDoorTimer {
			timer.Reset(3 * time.Second)
			resetDoorTimer = false
		}
		select {
		case a := <-doorTimerCh:
			resetDoorTimer = a
		case <-timer.C:
			timedOut <- 1
		default:
		}
		time.Sleep(10 * time.Millisecond)
	}
}
