package main

import (
	"log"
	"net"
	"bufio"
	"fmt"
)

func main(){
	port := ":30000"

	l, err := net.Listen("tcp",port)
	if err!=nil {
		log.Fatal(err)
	}
	defer l.Close()
	log.Println("Server listening on %s",port)
	for {
		//wait for connection
		conn, err := net.Dial("tcp",port)
		if err != nil {
			log.Fatal(err)
		}
		go ReadStuff(conn)
	}
}
func ReadStuff(c net.Conn){
	fmt.Println("Start ReadStuff")
	defer c.Close()
	reader := bufio.NewReader(c)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Recived",line)
	}
}
	