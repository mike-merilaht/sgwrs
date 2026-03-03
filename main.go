package main

import (
	"fmt"
	"encoding/json"
	"smwdd.io/sgwrs/ws/wsserver"
	"smwdd.io/sgwrs/ws/wsconnectedclient"
)

type Payload struct {
	Cmd string `json:"cmd"`
	Body string `json:"body"`
}

func mapToPayload(raw map[string]any) Payload {
	var payload Payload
	jsonbody, err := json.Marshal(raw)
    if err != nil {
        // do error check
        fmt.Println(err)
        return Payload{}
    }
	
	fmt.Println(jsonbody)
    if err := json.Unmarshal(jsonbody, &payload); err != nil {
        // do error check
        fmt.Println(err)
        return Payload{}
    }

	return payload
} 

func main() {
	server := wsserver.NewWSServer()
	server.RegisterJsonHandler(func(client wsconnectedclient.WSConnectedClient, jsonRaw map[string]any) {
		fmt.Println(jsonRaw)
		payload := mapToPayload(jsonRaw)
		fmt.Println(payload)
		server.Broadcast(client, payload.Body)
	})
	server.Listen();
}