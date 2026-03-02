package wsconnectedclient

import (
	"net"
)

type WSConnectedClient struct {
	Addr string
	Conn net.Conn
}
