


type CounterManager struct{
	//send message to slave
	counter int
	func sendToSlave(halla *net.UDPConn, udpaddr *net.UDPAddr){
		//send message to slave
		svar := []byte("heihei")
		_, err := halla.WriteToUDP(svar, udpaddr)
		if err != nil {
			log.Fatal(err)
		}
	}
	//count
	func countNumber(){
		counter++
	}
	//state machine
	type FsmState int 
	const (
		Slave FsmState = 0
		Master = 1
	)
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
	for{
		//Count and sende things
		CounterManager1.countNumber()
		CounterManager1.sendToSlave()
	}
}


//Become master
