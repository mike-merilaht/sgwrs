package main

import (
	"net"
	"fmt"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"bufio"
	"io"
)

const EXTENDED_16 = 126
const EXTENDED_64 = 127

type OpCode int
const (
	OpCodeContinuation OpCode = iota
	OpCodeText 
	OpCodeBinary

	OpCodeReserved3
	OpCodeReserved4
	OpCodeReserved5
	OpCodeReserved6
	OpCodeReserved7

	OpCodeClosed

	OpCodePing
	OpCodePong

	OpCodeReserved11
	OpCodeReserved12
	OpCodeReserved13
	OpCodeReserved14
	OpCodeReserved15
)

func (opcode OpCode) String() string {
    return [...]string{
		"0 (Continuation)",
		"1 (Text)",
		"2 (Binary)",
		"3 (RSV3)",
		"4 (RSV4)",
		"5 (RSV5)",
		"6 (RSV6)",
		"7 (RSV7)",
		"8 (Closed)",
		"9 (Ping)",
		"10 (Pong)",
		"11 (RSV11)",
		"12 (RSV12)",
		"13 (RSV13)",
		"14 (RSV14)",
		"15 (RSV15)",
	}[opcode]
}

func bytesToInt(bytes []byte) int {
    result := 0
    for i := 0; i < len(bytes); i++ {
        result = result << 8
        result += int(bytes[i])

    }

    return result
}

func handshake(conn net.Conn) {
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

type WSFrame struct  {
	fin 	bool
	opcode 	OpCode
	size 	int
	mask 	[]byte
	payload []byte
}

func newFrame(fin bool, opcode OpCode, size int, mask []byte, payload []byte) *WSFrame {
	frame := WSFrame{
		fin,
		opcode,
		size,
		mask,
		payload,
	}
	return &frame
}

func (frame WSFrame) Print() {
	fmt.Println("--- MESSAGE START ---")
	fmt.Printf("fin:     %t\n", frame.fin)
	fmt.Printf("opcode:  %s\n", frame.opcode)
	fmt.Printf("size:    %d \n", frame.size)
	fmt.Printf("mask:    %b\n", frame.mask)

	fmt.Printf("payload: %s\n", frame.UnmaskPayload())

	fmt.Println("---  MESSAGE END  ---")
}

func (frame WSFrame) PrintNetwork() {
	fmt.Printf("%s - %d bytes\n", frame.opcode, frame.size)
}

func (frame WSFrame) UnmaskPayload() []byte {
	data := make([]byte, frame.size)
	copy(data, frame.payload)
	for i := 0; i < frame.size; i++ {
	 	data[i] ^= frame.mask[i % 4]
	}
	return data;
}

func (frame WSFrame) ToSendableBytes() []byte {
	b := []byte{}

	// FIN + opcode
	if frame.fin {
		b = append(b, (0b1000 << 4) | byte(frame.opcode))
	} else {
		b = append(b, (0b0000 << 4) | byte(frame.opcode))
	}

	// TODO: Handle over 125 size
	if frame.size >= EXTENDED_16 {
		panic("TOO BIG")
	}

	// payload size
	b = append(b, byte(frame.size))

	// Unmasked payload
	b = append(b, frame.UnmaskPayload()...)

	return b
}

func extractFrame(conn net.Conn) (*WSFrame, error) {
	reader := bufio.NewReader(conn)

	header := make([]byte, 2)
	_, err := reader.Read(header)
	if err != nil {
		return nil, err
	}

	first_byte := header[0]

	/**
	* Bit 0
	*/
	fin := (first_byte & 0b10000000) != 0

	// Ignoring bits 1 through 3

	/**
	* Bits 4 through 7 are opcode
	*/
	opcode := OpCode(first_byte & 0b00001111)

	second_byte := header[1]

	/**
	* Bit 8 is whether or not we are masked
	*/
	// is_masked := (second_byte & 0b10000000) != 0

	// TODO: Reject unmasked. Must close connection

	/**
	* Bit 9 through 15 are the payload length
	*/
	payload_size := int(second_byte & 0b01111111)

	if payload_size == EXTENDED_16 {
		// Read next 2 bytes
		extension := make([]byte, 2)
		_, err := reader.Read(extension)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		payload_size = bytesToInt(extension)
	} else if (payload_size == EXTENDED_64) {
		// Read next 8 bytes
		extension := make([]byte, 8)
		_, err := reader.Read(extension)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		payload_size = bytesToInt(extension)
		panic("UNSUPPORTED PAYLOAD SIZE (64bits)")
	}

	mask := make([]byte, 4)
	_, err = reader.Read(mask)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	payload := make([]byte, payload_size)
	_, err = reader.Read(payload)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return newFrame(fin, opcode, payload_size, mask, payload), nil
}

func handlePing(conn net.Conn, frame *WSFrame) {
	f := newFrame(frame.fin, OpCodePong, frame.size, frame.mask, frame.payload)
	conn.Write(f.ToSendableBytes())
}

func handleText(conn net.Conn, frame *WSFrame) {
	str := string(frame.UnmaskPayload())
	if str == "ping" {
		f := newFrame(true, OpCodeText, 4, []byte{0, 0, 0, 0}, []byte("pong"))
		conn.Write(f.ToSendableBytes())
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	handshake(conn)

	for {
		frame, err := extractFrame(conn)
		if err == io.EOF {
			continue;
		} else if err != nil {
			panic(err);
		}

		frame.PrintNetwork()

		switch frame.opcode {
		case OpCodePong:
			// No expected response
		case OpCodePing:
			handlePing(conn, frame)
		case OpCodeText:
			handleText(conn, frame)
		case OpCodeClosed:
		default:
			frame.Print()
			fmt.Printf("Unsupported opcode: %s\n", frame.opcode)
		}
	}
}

func main() {
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

		go handleConn(conn)
	}
}