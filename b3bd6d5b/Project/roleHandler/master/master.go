package master

import (
	"Project/localElevator/elevio"
	"Project/network/messages"
	"Project/network/tcp"
	"Project/roleHandler/roleDistributor"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"runtime"
)


func HandlingMessages(
	jsonMsg []byte,
	iPToConnMap *map[string]net.Conn,
	sortedAliveElevs *[]net.IP,
	allHallReqAndStates *messages.HRAInput,
	sendNetworkMsgCh chan tcp.SendNetworkMsg,
) {
	typeMsg, dataMsg := messages.UnpackMessage(jsonMsg)
	switch typeMsg {
	case messages.MsgElevState:
		updatingIPAddr := dataMsg.(messages.ElevStateMsg).IpAddr
		updatingElevState := dataMsg.(messages.ElevStateMsg).ElevState
		(*allHallReqAndStates).States[updatingIPAddr] = updatingElevState
		if len(*iPToConnMap) > 1 && len(*sortedAliveElevs) > 1 {
			backupMsg := messages.PackMessage(messages.MsgHRAInput, (*allHallReqAndStates))
			backupConn := (*iPToConnMap)[(*sortedAliveElevs)[1].String()]
			sendNetworkMsgCh <- tcp.SendNetworkMsg{backupConn, backupMsg}
		}
	case messages.MsgHallReq:
		floor := dataMsg.(messages.HallReqMsg).Floor
		hallButton := dataMsg.(messages.HallReqMsg).Button
		addOrRemoveReq := dataMsg.(messages.HallReqMsg).TAddFRemove
		(*allHallReqAndStates).HallRequests[floor][hallButton] = addOrRemoveReq
		if len(*iPToConnMap) > 1 && len(*sortedAliveElevs) > 1 {
			backupMsg := messages.PackMessage(messages.MsgHRAInput, (*allHallReqAndStates))
			backupConn := (*iPToConnMap)[(*sortedAliveElevs)[roleDistributor.Backup].String()]
			sendNetworkMsgCh <- tcp.SendNetworkMsg{backupConn, backupMsg}
		}
		var inputToHRA = messages.HRAInput{
			HallRequests: make([][2]bool, elevio.N_FLOORS),
			States:       make(map[string]messages.HRAElevState),
		}
		inputToHRA.HallRequests = (*allHallReqAndStates).HallRequests
		for _, ip := range *sortedAliveElevs {
			state, exists := (*allHallReqAndStates).States[ip.String()]
			if exists {
				inputToHRA.States[ip.String()] = state
			} 
		}
		output := runHallRequestAssigner(inputToHRA)
		jsonLightMsg := messages.PackMessage(messages.MsgHallLigths, dataMsg)
		for ipAddr, hallRequest := range output {
			jsonHallReqMsg := messages.PackMessage(messages.MsgAssignedHallReq, hallRequest)
			sendNetworkMsgCh <- tcp.SendNetworkMsg{(*iPToConnMap)[ipAddr], jsonHallReqMsg}
			sendNetworkMsgCh <- tcp.SendNetworkMsg{(*iPToConnMap)[ipAddr], jsonLightMsg}
		}
	}
}

func runHallRequestAssigner(input messages.HRAInput) map[string][][2]bool {
	hraExecutable := ""
	switch runtime.GOOS {
	case "linux":
		hraExecutable = "hall_request_assigner"
	case "windows":
		hraExecutable = "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}
	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
		return nil
	}
	ret, err := exec.Command(hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
		return nil
	}
	output := new(map[string][][2]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		return nil
	}
	return *output
}

