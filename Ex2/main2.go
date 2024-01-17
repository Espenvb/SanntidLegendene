package main

import (
	"log"
	"net"
	"fmt"
)






func main(){
	port := 30000
	portsend := 20010


	

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



	addr2, err := net.ResolveUDPAddr("udp",fmt.Sprintf(":%d",portsend))

	udpaddr := &net.UDPAddr{
		IP:	net.ParseIP("10.100.23.129"),
		Port: 20010,
	}
	
	svar := []byte("heihei")




	høre, err := net.ListenUDP("udp", addr2)


	_, err = conn.WriteToUDP(svar,udpaddr)
	if err != nil {
		log.Fatal(err)
	}
	nybuffer := make([]byte,1024)
	p, _, err := høre.ReadFromUDP(nybuffer)
	if err!=nil {
		log.Fatal(err)
	}


	fmt.Printf("svar:",string(nybuffer[:p]))




	
	
}
	