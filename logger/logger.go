package logger

import (
	"net"
	"fmt"
	"time"
)

func Network(conn net.Conn, message string) {
	fmt.Printf("[%s] - %s - %s", time.Now().Format(time.RFC3339), conn.RemoteAddr().String(), message)
}