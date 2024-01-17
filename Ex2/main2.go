package main

import (
	"log"
	"net"
	"fmt"
)


func getIP(){
	port := 30000

	addr, err := net.ResolveUDPAddr("udp",fmt.Sprintf(":%d",port))
	if err!=nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp",addr)
	if err!=nil {
		log.Fatal(err)
	}

	defer conn.Close()

	
	buffer := make([]byte,1024)

	n, _, err := conn.ReadFromUDP(buffer)
	if err!=nil {t
	}

	message := string(buffer[:n])
	fmt.Printf("melding:", message)
}

func sendmessage(*net.UDPAddr halla){
	svar := []byte("heihei")
	_, err = halla.WriteToUDP(svar,udpaddr)
	if err != nil {
		log.Fatal(err)
	}

}


func readmessage(*net.UDPAddr halla){
	nybuffer := make([]byte,1024)
	p, _, err := halla.ReadFromUDP(nybuffer)
	if err!=nil {
		log.Fatal(err)
	}


	fmt.Printf("svar:",string(nybuffer[:p]))


}




func main(){
	serverPort := 20010


udpaddr := &net.UDPAddr{
		IP:	net.ParseIP("10.100.23.129"),
		Port: 20010,
	}
	
addr2, err := net.ResolveUDPAddr("udp",fmt.Sprintf(":%d",serverPort))

høre, err := net.ListenUDP("udp", addr2)	


getIP()
sendmessage(høre)
readmessage(høre)


	
	
}
	