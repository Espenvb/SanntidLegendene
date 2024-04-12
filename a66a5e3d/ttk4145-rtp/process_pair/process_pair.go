package process_pair

import (
	"fmt"
	"os/exec"
	"time"
)

var _isInitialized bool = false

// corresponding to flags in main
var isSupervisor bool

var groupId int
var broadcastPort int
var communicationPort int

var elevatorServerIP string
var elevatorServerPort int
var m_numFloors int

var hraExecutablePath string

func Init(
	isSupervisorIn bool,
	groupIdIn int,
	broadcastPortIn int,
	communicationPortIn int,
	elevatorServerIPIn string,
	elevatorServerPortIn int,
	m_numFloorsIn int,
	hraExecutablePathIn string,
) {
	_isInitialized = true
	isSupervisor = isSupervisorIn
	groupId = groupIdIn
	broadcastPort = broadcastPortIn
	communicationPort = communicationPortIn
	elevatorServerIP = elevatorServerIPIn
	elevatorServerPort = elevatorServerPortIn
	m_numFloors = m_numFloorsIn
	hraExecutablePath = hraExecutablePathIn
}

func getMainProgramGoFlags() string {
	return fmt.Sprintf("--group-id %d --bport %d --cport %d --elevator-server-ip %s --eport %d -m %d --hra='%s'",
		groupId,
		broadcastPort,
		communicationPort,
		elevatorServerIP,
		elevatorServerPort,
		m_numFloors,
		hraExecutablePath,
	)
}

func startSupervisor() {
	startGoProgram := fmt.Sprintf("go run main.go --is-supervisor=true %s; exec bash", getMainProgramGoFlags())
	(exec.Command("gnome-terminal", "--tab", "--", "bash", "-c", startGoProgram)).Run()
	// (exec.Command("gnome-terminal", "--", startGoProgram)).Run()
}

func startMainProgram() {
	startGoProgram := fmt.Sprintf("go run main.go %s; exec bash", getMainProgramGoFlags())
	(exec.Command("gnome-terminal", "--tab", "--", "bash", "-c", startGoProgram)).Run()
	// (exec.Command("gnome-terminal", "--", startGoProgram)).Run()
}

func SendProcessPairAliveMessages(pairMessageChan chan string) {
	if !_isInitialized {
		panic("Using non initialized process_pair package!")
	}

	message := fmt.Sprintf("hello from the other pair, cport=%d, isSupervisor=%t",
		communicationPort,
		isSupervisor)
	for {
		pairMessageChan <- message
		time.Sleep(50 * time.Millisecond)
	}
}

func ObserverProcessPair(partner string, pairMessageChan chan string) {
	if !_isInitialized {
		panic("Using non initialized process_pair package!")
	}

	timeout := 3 * time.Second
	timer := time.NewTimer(timeout)

	for {
		select {
		case <-pairMessageChan:
			// fmt.Println(msg)
			timer.Stop()
			timer.Reset(timeout)
		case <-timer.C:
			switch partner {
			case "main-program":
				startMainProgram()
			case "supervisor":
				startSupervisor()
			}
			timer.Reset(timeout) // grant some time for restarting to avoid multiple restarts
		}
	}
}
