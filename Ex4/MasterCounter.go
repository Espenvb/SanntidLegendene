package main

import(

	"net"
	"fmt"
	"os"
	"os/exec"
	"time"
	"math/rand"
)

type CounterManager struct{
	//send message to slave
	counter int
	
	
	//count
	//state machine
	State FsmState
}

type FsmState int 
	const (
		Slave FsmState = 0
		Master = 1
	)


func sendToSlave(s *CounterManager,halla *net.UDPConn, udpaddr *net.UDPAddr){
	//send message to slave
	svar := []byte("heihei")
	_, err := halla.WriteToUDP(svar, udpaddr)
	if err != nil {
		log.Fatal(err)
	}
}


func countNumber(s *CounterManager){
	s.counter++
}

func InitializeCounter(state FsmState) CounterManager {
	CounterManager1 := CounterManager{
		counter: 0,
		State:   state,
	}
	return CounterManager1
}

func main() {
	//Create Slave
	Counter := InitializeCounter(Slave)
	for {
		switch Counter.State {
		case Slave:
			// Do slave stuff
			// Listen and stuff
		case Master:
		}
	}
}
