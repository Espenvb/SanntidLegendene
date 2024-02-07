package main
import{
	"net"
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


func count(){
	for{
		Value++
	}
}

func send(conn *net.UDPConn,  udpaddr *net.UDPAddr){
	svar := []byte(Value)
	_, err := conn.WriteToUDP(svar, )
}

func master (){
	count()
	//send

}





func main(){
	Value int := 0
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

go master
go slave

}