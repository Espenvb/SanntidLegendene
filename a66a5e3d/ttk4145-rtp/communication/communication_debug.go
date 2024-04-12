package communication

import (
	"fmt"
	pt "project/project_types"
	"time"
)

// Add this to main program to plug in debugging:
// debugStringChan := make(chan string, 1)
// go communication.Receiver(*communicationPort, debugStringChan)
// go communication.DebugSendReceive(debugStringChan)
// go communication.DebugSendSend()
func DebugSendReceive(stringChan <-chan string) {
	for {
		fmt.Println(<-stringChan)
	}
}

func DebugSendSend() {
	for {
		time.Sleep(5 * time.Second)
		Send(
			pt.Node{
				GroupId: 100,
				IP:      "localhost", // "localhost" if not at lab
				Port:    10001,
				// Type:    pt.NT_Undefined,
			},
			"Hello World!",
		)
	}
}
