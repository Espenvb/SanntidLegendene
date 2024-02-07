package main

import(
	"net"
	"log"
	"time"
	"fmt"
)



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

func slave(conn *net.UDPConn, msgReceived chan bool){
	//readStuff
	lastReceived := time.Now()

	for {
		select {
		case <-time.After(1 * time.Second):
			elapsed := time.Since(lastReceived)
			if elapsed >= 1*time.Second {
				fmt.Println("Timeout expired, no message received in 1 second.")
			}
		case <-msgReceived:
			lastReceived = time.Now()
			fmt.Println("Message received, resetting timer.")
		}
	}
}

func main() {

	Crash := make(chan int)
	received := make(chan int)

	//Setup UDP
	// Create a UDP listener
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:12345")
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		return
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer conn.Close()


	select{

	case a <- Crash:
		//gjør til master og lag ny
	
	case a <- received:
		//oppdater verdi

	case a <- counter:
		// send funksjon

	case a <- 
	}




}
