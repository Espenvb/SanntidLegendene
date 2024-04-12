package tcp

import (
	"encoding/json"
	"fmt"
	"net"
	"reflect"
)

type TaggedJson struct {
	Type string
	JSON []byte
}

func Transmit(conn net.Conn, data interface{}) {
	if conn == nil {
		fmt.Println("[error] Connection is nil")
		return

	}
	buffer, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("[error] Failed to encode data with error: %v\n", err)
		return
	}

	buffer, err = json.Marshal(TaggedJson{reflect.TypeOf(data).Name(), buffer})
	if err != nil {
		fmt.Printf("[error] Failed to make buffer with error:")
	}

	_, err = conn.Write(buffer)
	if err != nil {
		fmt.Printf("[error] Failed to write: %v\n", err)
		return
	}

}

func Receive(conn net.Conn, data ...interface{}) {
	if conn == nil {
		fmt.Println("[error] conn is nil")
		return
	}
	defer conn.Close()

	channels := make(map[string]interface{})

	for _, channel := range data {
		if channel == nil || reflect.TypeOf(channel).Kind() != reflect.Chan {
			panic("Arguments contains one or more non channel type\n")
		}

		channels[reflect.TypeOf(channel).Elem().Name()] = channel
	}

	var tj TaggedJson
	buffer := make([]byte, 1024)
	for {
		length, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("[error] Failed to read: %v\n", err)
			break
		}

		err = json.Unmarshal(buffer[0:length], &tj)

		if err != nil {
			fmt.Printf("[error] Failed to marshal JSON with error: %v", err)
			continue
		}

		channel, ok := channels[tj.Type]

		if !ok {
			fmt.Printf("[warning] Recieved type we are not listening to: %v\n", tj.Type)
			continue
		}

		value := reflect.New(reflect.TypeOf(channel).Elem())

		err = json.Unmarshal(tj.JSON, value.Interface())

		if err != nil {
			fmt.Printf("[error] Failed to unmarshal data with error code: %v\n", err)
			continue
		}

		reflect.Select([]reflect.SelectCase{{
			Dir:  reflect.SelectSend,
			Chan: reflect.ValueOf(channel),
			Send: reflect.Indirect(value),
		}})
	}
}
