package wsserver

import (
	"sync"
	"net"
	"bytes"
	"fmt"
	"bufio"
	"io"
	"strings"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"smwdd.io/sgwrs/logger"
	"smwdd.io/sgwrs/ws/wsframe"
	"smwdd.io/sgwrs/ws/opcode"
	"smwdd.io/sgwrs/ws/wsconnectedclient"
)


type WSServer struct {
	clients []wsconnectedclient.WSConnectedClient
	jsonHandler func(wsconnectedclient.WSConnectedClient, map[string]any)
	mu sync.Mutex
}

func (server *WSServer) handshake(conn net.Conn) {
	reader := bufio.NewReader(conn)
	
	var req bytes.Buffer
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return
		}
		req.Write(line)
		if bytes.HasSuffix(req.Bytes(), []byte("\r\n\r\n")) {
			break
		}
	}

	var key string
	for _, line := range bytes.Split(req.Bytes(), []byte("\r\n")) {
		if bytes.HasPrefix(line, []byte("Sec-WebSocket-Key:")) {
			key = string(bytes.TrimSpace(bytes.Split(line, []byte(":"))[1]))
		}
	}

	if key == "" {
		return
	}

	magic_string := "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

	hash := sha1.New()
	hash.Write([]byte(key + magic_string))
	b64 := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	f := "HTTP/1.1 101 Switching Protocols\r\n" +
	"Upgrade: websocket\r\n" + 
	"Connection: Upgrade\r\n" + 
	"Sec-WebSocket-Accept: " +
	b64 +
	"\r\n" +
	"\r\n"
	conn.Write([]byte(f))
}

func NewWSServer() *WSServer {
	return &WSServer{clients: []wsconnectedclient.WSConnectedClient{}, jsonHandler: func (client wsconnectedclient.WSConnectedClient, m map[string]any) {}}
}

func (server *WSServer) Broadcast(from wsconnectedclient.WSConnectedClient, body string) {
	jsonbody, err := json.Marshal(map[string]string{"cmd": "chat", "body": body})
    if err != nil {
        // do error check
        fmt.Println(err)
        return
    }
	body_bytes := []byte(jsonbody)
	fmt.Printf("%s\n", string(jsonbody))
	frame := wsframe.NewWSFrame(true, opcode.OpCodeText, len(body_bytes), []byte{0, 0, 0, 0}, body_bytes)
	payload := frame.ToSendableBytes()
	for _, client := range server.clients {
		client.Conn.Write(payload)
	} 
}

func (server *WSServer) handlePing(conn net.Conn, frame *wsframe.WSFrame) {
	f := wsframe.NewWSFrame(frame.Fin, opcode.OpCodePong, frame.Size, frame.Mask, frame.Payload)
	conn.Write(f.ToSendableBytes())
}

func (server *WSServer) handleText(conn net.Conn, frame *wsframe.WSFrame) {
	str := string(frame.UnmaskPayload())
	if str == "ping" {
		f := wsframe.NewWSFrame(true, opcode.OpCodeText, 4, []byte{0, 0, 0, 0}, []byte("pong"))
		conn.Write(f.ToSendableBytes())
	}

	if strings.HasPrefix(str, "{") {
		i := server.findClientIndex(conn)
		var result map[string]any
		json.Unmarshal([]byte(str), &result)
		server.jsonHandler(server.clients[i], result)
	}
}

func (server *WSServer) handleConn(conn net.Conn) {
	defer conn.Close()

	server.handshake(conn)

	server.addClient(conn)

	for {
		frame, err := wsframe.ExtractFrame(conn)
		if err == io.EOF {
			continue;
		} else if err != nil {
			panic(err);
		}

		logger.Network(conn, fmt.Sprintf("%s - %d bytes\n", frame.Opcode, frame.Size))

		if frame.Opcode == opcode.OpCodeClosed {
			break
		}

		switch frame.Opcode {
		case opcode.OpCodeClosed:
			// Included for completeness, but never hit
		case opcode.OpCodePong:
			// No expected response
		case opcode.OpCodePing:
			server.handlePing(conn, frame)
		case opcode.OpCodeText:
			server.handleText(conn, frame)
		default:
			frame.Print()
			fmt.Printf("Unsupported opcode: %s\n", frame.Opcode)
		}
	}

	server.dropClient(conn)
}

func (server *WSServer) addClient(conn net.Conn) {
	server.mu.Lock()
    defer server.mu.Unlock()

	client := wsconnectedclient.WSConnectedClient{conn.RemoteAddr().String(), conn}
	server.clients = append(server.clients, client)

	fmt.Printf("connected %s\n", client.Addr)

	server.PrintConnectedClients()
}

func (server *WSServer) PrintConnectedClients() {
	fmt.Printf("clients connected: %d\n", len(server.clients))
}

func (server *WSServer) findClientIndex(conn net.Conn) int {
	fmt.Println(len(server.clients))
	for i := 0; i < len(server.clients); i++ {
		if server.clients[i].Addr == conn.RemoteAddr().String() {
			return i
		}
	}
	return -1;
}

func (server *WSServer) dropClient(conn net.Conn) {
	server.mu.Lock()
    defer server.mu.Unlock()

	index := server.findClientIndex(conn)
	if index == -1 {
		return
	}
	
	client := server.clients[index]
	fmt.Printf("disconnected %s\n", client.Addr)
	server.clients = append(server.clients[:index], server.clients[index+1:]...)
	server.PrintConnectedClients()
}

func (server *WSServer) RegisterJsonHandler(handler func(wsconnectedclient.WSConnectedClient, map[string]any)) {
	server.jsonHandler = handler
}

func (server *WSServer) Listen() {
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		go server.handleConn(conn)
	}
}