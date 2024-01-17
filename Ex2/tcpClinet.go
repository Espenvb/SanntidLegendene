package main

import (
	"fmt"
	"log"
	"net"
)

//portDel := "33546"


func connecting(conn net.Conn){
	defer conn.Close()

	fmt.Println("accepted")

	melding := "velkommen"
	conn.Write([]byte(melding))

}



func main() {

	conn, err := net.Dial("tcp", "10.100.23.129:33546")
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	

	sendbuffer := append([]byte("Connect to: 10.100.23.20:33546"), 0)
	_, err1 := conn.Write(sendbuffer)
	if err1 != nil {
		log.Fatal(err)
	}

	

		listen,err := net.Listen("tcp", fmt.Sprintf(":%d",33546))
		if err != nil {
			log.Fatal(err)
		}
		defer listen.Close()

		for{
			conn2, err := listen.Accept()
			if err!=nil{
				log.Fatal(err)
			}
		
		
	go connecting(conn2)

		}
	


		
	/*	n, err := conn.Read(buffer)
		if err != nil {
			log.Fatal(err)
		}

		message := string(buffer[:n])
		fmt.Printf("Message : %s\n", message)
*/

	


}