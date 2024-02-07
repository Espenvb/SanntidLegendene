package main

import{
	"net"
	"fmt"
	"os"
	"os/exec"
	"time"
	"math/rand"
}

func start_node() *exec.Cmd {
	//Starts a new node
	fmt.Println("Starting new node")

	cmd := exec.Command("cmd", "/C", "start", "powershell", "go", "run", "ov4.go")
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
	minTime = 4
	rand.Seed(time.Now().UnixNano())
	randTime := rand.Intn(5)
	time.Sleep((minTime + randTime)*time.Second)
	os.Exit(0)
}

//type countermaster struct{
	//counter int
	//func send
	//func read
	//func count
	//state for backup eller master
//}


func send(addr string,conn *net.UDPConn){
	buffer := make([]byte, 1024)

	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		log.Fatal(err)
	}

	message := string(buffer[:n])
	fmt.Printf("melding:", message)

}




func main(){

	broadcastaddr := "255.255.255.255"
	port := 12345

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d",broadcastaddr ,port))
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
    if err != nil {
        fmt.Println("Error dialing UDP connection:", err)
        return

	defer conn.Close()
	}



for{

}

}