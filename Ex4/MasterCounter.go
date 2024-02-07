package main

import{

	"net"
}

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

func Initilize_counter(FsmState) CounterManager{
	CounterManager1:= CounterManager{
		counter: 0,
		state: FsmState(1),

	}
	return CounterManager1
}


func main(){
	//initialize counter
	

	CounterManager1:= Initilize_counter(Master)
	CounterSlave := Initilize_counter(Slave)

	go doMasterStuff(){
		
		CounterManager1.countNumber()
		CounterManager1.sendToSlave()
	}
	for{
		select
			CounterManager.FsmState == 1
				//Do master shit

		CounterManager1.countNumber()
		CounterManager1.sendToSlave()
	}
}


//Become master
