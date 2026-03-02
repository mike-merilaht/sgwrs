package main

import (
	"smwdd.io/sgwrs/ws/wsserver"
)

func main() {
	server := wsserver.NewWSServer()
	server.Listen();
}