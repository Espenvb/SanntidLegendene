package alive

import (
	"fmt"
	pt "project/project_types"
	"time"
)

func DebugAlive(nodeUpdateChan <-chan pt.NodeUpdate, startAliveMessageSending chan<- bool) {
	// debug alive
	startAliveMessageSending <- true
	go testAliveListener(nodeUpdateChan)
	// TODO: document behaviour (UNDEFINED node is different from BACKUP node with same ip,port though)
	time.Sleep(4 * time.Second)
	startAliveMessageSending <- false
}

func testAliveListener(nodeUpdateChan <-chan pt.NodeUpdate) {
	for nodeUpdate := range nodeUpdateChan {
		fmt.Println("nodeUpdate:", nodeUpdate)
	}
}
