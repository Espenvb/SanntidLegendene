package messages

import (
	"Project/localElevator/elevio"
	"encoding/json"
	"fmt"
	"strings"
)

const MsgHRAInput = "HRAInput"
const MsgElevState = "ElevState"
const MsgHallReq = "HallReq"
const MsgHallLigths = "HallLights"
const MsgAssignedHallReq = "AssignedHallReq"
const MsgRestoreCabReq = "CabReq"

type HRAElevState struct {
	Behaviour   string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

type ElevStateMsg struct {
	IpAddr    string       `json:"ipAdress"`
	ElevState HRAElevState `json:"elevState"`
}

type dataWithTypeMsg struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type HallReqMsg struct {
	TAddFRemove bool              `json:"tAdd_fRemove"`
	Floor       int               `json:"floor"`
	Button      elevio.ButtonType `json:"button"`
}

func PackMessage(structType string, msg interface{}) []byte {
	msgJsonBytes, _ := json.Marshal(msg)
	dataToSend := dataWithTypeMsg{
		Type: structType,
		Data: msgJsonBytes,
	}
	finalJSONBytes, _ := json.Marshal(dataToSend)
	finalJSONBytes = append(finalJSONBytes, '&')
	return finalJSONBytes
}

func UnpackMessage(jsonBytes []byte) (string, interface{}) {
	var dataWithTypeMsg dataWithTypeMsg
	err := json.Unmarshal(jsonBytes, &dataWithTypeMsg)
	if err != nil {
		return dataWithTypeMsg.Type, nil
	}
	switch dataWithTypeMsg.Type {
	case MsgHRAInput:
		var HRAInputData HRAInput
		err = json.Unmarshal(dataWithTypeMsg.Data, &HRAInputData)
		return dataWithTypeMsg.Type, HRAInputData
	case MsgElevState:
		var MsgElevStateData ElevStateMsg
		err = json.Unmarshal(dataWithTypeMsg.Data, &MsgElevStateData)
		return dataWithTypeMsg.Type, MsgElevStateData
	case MsgHallReq:
		var MsgHallReqData HallReqMsg
		err = json.Unmarshal(dataWithTypeMsg.Data, &MsgHallReqData)
		return dataWithTypeMsg.Type, MsgHallReqData
	case MsgHallLigths:
		var MsgHallLightsData HallReqMsg
		err = json.Unmarshal(dataWithTypeMsg.Data, &MsgHallLightsData)
		return dataWithTypeMsg.Type, MsgHallLightsData
	case MsgAssignedHallReq:
		var MsgAssignedHallReqData [][2]bool
		err = json.Unmarshal(dataWithTypeMsg.Data, &MsgAssignedHallReqData)
		return dataWithTypeMsg.Type, MsgAssignedHallReqData
	case MsgRestoreCabReq:
		var MsgRestoreCabReqData []bool
		err = json.Unmarshal(dataWithTypeMsg.Data, &MsgRestoreCabReqData)
		return dataWithTypeMsg.Type, MsgRestoreCabReqData
	default:
		return dataWithTypeMsg.Type, nil
	}
}

func DistributeMessages(
	jsonMessageCh chan []byte,
	toFSMCh chan []byte,
	toRoleCh chan []byte,
) {
	var dataWithTypeMsg dataWithTypeMsg
	for {
		jsonMsgReceived := <-jsonMessageCh
		jsonObjects := strings.Split(string(jsonMsgReceived), "&")
		for _, jsonObject := range jsonObjects {
			if jsonObject != "" {
				err := json.Unmarshal([]byte(jsonObject), &dataWithTypeMsg)
				if err != nil {
					fmt.Println("Error decoding json:", err)
					break
				}
				switch dataWithTypeMsg.Type {
				case MsgHRAInput, MsgElevState, MsgHallReq:
					toRoleCh <- []byte(jsonObject)
				case MsgHallLigths, MsgAssignedHallReq, MsgRestoreCabReq:
					toFSMCh <- []byte(jsonObject)
				}
			}
		}
	}
}
