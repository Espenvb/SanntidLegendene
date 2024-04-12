package process_pair

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

func Primary(counter int) {
	var addr string = "localhost:8070"
	fmt.Printf("This is now a primary.")
	exec.Command("gnome-terminal", "--", "go", "run", "main.go").Run()
	for {
		conn, err := net.Dial("udp", addr)
		if err != nil {
			fmt.Println("The following error occured", err)
		}
		conn.Write([]byte(""))
		time.Sleep(20 * time.Millisecond)
	}
}

func Backup() int {
	var addr string = "localhost:8070"
	println("This is a process pair backup.")
	udpConn, err := net.ListenPacket("udp", addr)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer udpConn.Close()

	buffer := make([]byte, 1024)
	for {
		err = udpConn.SetReadDeadline(time.Now().Add(4 * time.Second))
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		_, _, err := udpConn.ReadFrom(buffer)
		if err != nil {
			fmt.Println("Error:", err)
			return 0
		}
	}
}

func WriteToLocalBackup(cabCalls []bool) error {
	file, err := os.Create("localBackup.txt")
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	var boolStr strings.Builder
	for _, value := range cabCalls {
		if value {
			boolStr.WriteString("true")
		} else {
			boolStr.WriteString("false")
		}
	}

	_, err = file.WriteString(boolStr.String())
	if err != nil {
		return err
	}
	return nil
}
