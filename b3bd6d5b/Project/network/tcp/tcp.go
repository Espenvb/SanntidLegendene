package tcp

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)


type EditConnMap struct {
	Insert 		bool
	ClientIP 	string
	Conn       	net.Conn
}

func EstablishMainListener(
	isMainListenerCh chan bool, 
	listenerPort string,
	editMastersConnMapCh chan EditConnMap,
	incommingNetworkMsgCh chan []byte,
) {
	var ctxRecieveMsg context.Context
	var cancelRecievMsg context.CancelFunc
	iPToConnMap := make(map[string]net.Conn)
	var wasMainListener bool
	for {
		isMainListener := <- isMainListenerCh
		if isMainListener {
			wasMainListener = true
			go func() {
				tcpAddr, err := net.ResolveTCPAddr("tcp4", ":"+listenerPort)
				if err != nil {
					fmt.Printf("Could not resolve address: %s\n", err)
					os.Exit(1)
				}
				listener, err := net.ListenTCP("tcp", tcpAddr)
				if err != nil {
					fmt.Println("Could not open listener: ", err)
					return
				}
				defer listener.Close()
				for {
					conn, err := listener.Accept()
					if err != nil {
						fmt.Println("Error accepting: ", err)
						continue
					}
					connectionsIP := ((conn.RemoteAddr().(*net.TCPAddr)).IP).String()
					(iPToConnMap)[connectionsIP] = conn
					editMastersConnMapCh <- EditConnMap{true, connectionsIP, conn}
					ctxRecieveMsg, cancelRecievMsg = context.WithCancel(context.Background())
					go recieveMessage(conn, incommingNetworkMsgCh, ctxRecieveMsg)
				}
			}()
			go func() {
				for ip, conn := range iPToConnMap {
					_, err := conn.Read(make([]byte, 1024))
					if err != nil {
						delete((iPToConnMap), ip)
						editMastersConnMapCh <- EditConnMap{true, ip, conn}
					}
				}
			}()
		} else {
			if wasMainListener {
				cancelRecievMsg()
				wasMainListener = false
			}
		}
	}
}

func EstablishConnectionAndListen(
	ipCh chan net.IP, port string,
	connCh chan net.Conn,
	incommingNetworkMsgCh chan []byte,
) {
	var ctxRecieveMsg context.Context
	var cancelRecievMsg context.CancelFunc
	var conn net.Conn
	for {
		IP := <-ipCh
		tcpAddr, err := net.ResolveTCPAddr("tcp4", IP.String()+":"+port)
		if err != nil {
			fmt.Printf("Could not resolve address: %s\n", err)
		}
		if conn != nil {
			cancelRecievMsg()
		}
		for {
			conn, err = net.Dial("tcp", tcpAddr.String())
			if err != nil {
				fmt.Println("Could not connect to server: ", err)
				time.Sleep(50*time.Millisecond)
			}else {
				break
			}
		}
		connCh <- conn
		ctxRecieveMsg, cancelRecievMsg = context.WithCancel(context.Background())
		go recieveMessage(conn, incommingNetworkMsgCh, ctxRecieveMsg)
	}
}

func recieveMessage(
	conn net.Conn,
	incommingMsgCh chan<- []byte,
	ctxRecieveMsg context.Context,
) {
	for {
		select {
		case <- ctxRecieveMsg.Done():
			return
		default:
			buffer := make([]byte, 65536)
			data, err := conn.Read(buffer)
			if err != nil {
				conn.Close()
				if err == io.EOF {
					fmt.Println("Client closed the connection")
				} else {
					fmt.Println("Error:", err)
				}
				return
			}
		msg := make([]byte, data)
		copy(msg, buffer[:data])
		incommingMsgCh <- msg
		}
	}
}
type SendNetworkMsg struct {
	RecieverConn net.Conn
	Message 	[]byte
}

func SendMessage(sendNetworkMsgCh chan SendNetworkMsg) {
	for {
		sendNetworkMsg := <-sendNetworkMsgCh
		conn := sendNetworkMsg.RecieverConn
		message := sendNetworkMsg.Message
		_, err := conn.Write(message)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Connection closed by the server.")
			} else {
				fmt.Println("Error sending data to server:", err)
			}
			return
		}
	}
}

