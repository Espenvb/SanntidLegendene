package roleFSM

import (
	"Project/localElevator/elevio"
	"Project/network/messages"
	"Project/network/tcp"
	"Project/roleHandler/master"
	"Project/roleHandler/roleDistributor"
	"net"
)

var iPToConnMap map[string]net.Conn

func RoleFSM(
	roleAndSortedAliveElevsCh chan roleDistributor.RoleAndSortedAliveElevs, 
	toRoleFSMCh chan []byte, 
	sendNetworkMsgCh chan tcp.SendNetworkMsg, 
	isMasterCh chan bool, 
	editMastersConnMapCh chan tcp.EditConnMap,
) {
	var allHallReqAndStates = messages.HRAInput{
		HallRequests: make([][2]bool, elevio.N_FLOORS),
		States:       make(map[string]messages.HRAElevState),
	}
	roleAndSortedAliveElevs := <-roleAndSortedAliveElevsCh
	role := roleAndSortedAliveElevs.Role
	sortedAliveElevs := roleAndSortedAliveElevs.SortedAliveElevs
	if role == "Master" {
		isMasterCh <- true
		iPToConnMap = make(map[string]net.Conn)
	}
	for {
		select {
		case newMsg := <-toRoleFSMCh:
			switch role {
			case "Master":
				master.HandlingMessages(newMsg, &iPToConnMap, &sortedAliveElevs, &allHallReqAndStates, sendNetworkMsgCh)
			case "Backup":
				typeMsg, dataMsg := messages.UnpackMessage(newMsg)
				switch typeMsg {
				case messages.MsgHRAInput:
					allHallReqAndStates = messages.HRAInput{
						HallRequests: dataMsg.(messages.HRAInput).HallRequests,
						States:       dataMsg.(messages.HRAInput).States,
					}
				}
			default:
			}
		case changedRoleAndSortedAliveElevs := <-roleAndSortedAliveElevsCh:
			changedRole := changedRoleAndSortedAliveElevs.Role
			sortedAliveElevs = changedRoleAndSortedAliveElevs.SortedAliveElevs
			if changedRole != "" && changedRole != role {
				role = changedRole
				switch role {
				case "Master":
					iPToConnMap = make(map[string]net.Conn)
					isMasterCh <- true
				default:
					isMasterCh <- false
				}
			}
		case editMastersConnMap := <-editMastersConnMapCh:
			insert := editMastersConnMap.Insert
			elevatorIP := editMastersConnMap.ClientIP
			elevatorConn := editMastersConnMap.Conn
			if insert {
				(iPToConnMap)[elevatorIP] = elevatorConn
				if _, exists := allHallReqAndStates.States[elevatorIP]; exists {
					cabRequests := allHallReqAndStates.States[elevatorIP].CabRequests
					cabReqestMsg := messages.PackMessage(messages.MsgRestoreCabReq, cabRequests)
					sendNetworkMsgCh <- tcp.SendNetworkMsg{iPToConnMap[elevatorIP], cabReqestMsg}
				}
			} else {
				delete((iPToConnMap), elevatorIP)
			}
		}
	}
}