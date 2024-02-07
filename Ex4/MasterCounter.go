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

func Initilize_counter(FsmState) CounterManager{
	CounterManager1:= CounterManager{
		counter: 0,
		state: FsmState(1),

	}
	return CounterManager1
}


func start_node() *exec.Cmd {
	//Starts a new node
	fmt.Println("Starting new node")

	cmd := exec.Command("cmd", "/C", "start", "powershell", "go", "run", "MasterCounter.go")
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error:", err)
		return cmd
	}
	fmt.Println("Created new node")
	return cmd
}

func kill_self(){
	//Kills itself after a random amount of time
	minTime := 4
	//rand.Seed(time.Now().UnixNano())
	randTime := time.Duration((minTime + rand.Intn(5)))*time.Second
	time.Sleep((randTime))
	os.Exit(0)
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
