package main

import (
	//"fmt"
	"log"
	"net"
)

//portDel := "33546"

/*
func connecting(conn net.Conn){
	defer conn.Close()


	fmt.Println("accepted")

	melding := "velkommen"
	conn.Write([]byte(melding))

}
*/


func main() {

	conn, err := net.Dial("tcp", "localhost:15657")
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	message := make([]byte,4)
	message[0] = 0x01
	message[1] = 0x01
	message[2] = 0x01
	message[3] = 0x81


	
	_, err = conn.Write(message)


	/*buffer := make([]byte,1024)
	n, err := conn.Read(buffer)
	if err != nil{
		log.Fatal(err)
	}

	fmt.Println(buffer[:n])
*/
	}






	
/*
	sendbuffer := append([]byte(""), 0)
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
		    
    ret, err := exec.Command("../hall_request_assigner/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
    if err != nil {
        fmt.Println("exec.Command error: ", err)
        fmt.Println(string(ret))
        return
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

	


