package wsframe

import (
	"fmt"
	"net"
	"bufio"
	"smwdd.io/sgwrs/ws/opcode"
	"smwdd.io/sgwrs/utils"
)

const EXTENDED_16 = 126
const EXTENDED_64 = 127

type WSFrame struct  {
	Fin 	bool
	Opcode 	opcode.OpCode
	Size 	int
	Mask 	[]byte
	Payload []byte
}

func ExtractFrame(conn net.Conn) (*WSFrame, error) {
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
	opcode := opcode.OpCode(first_byte & 0b00001111)

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
		payload_size = numbers.BytesToInt(extension)
	} else if (payload_size == EXTENDED_64) {
		// Read next 8 bytes
		extension := make([]byte, 8)
		_, err := reader.Read(extension)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		payload_size = numbers.BytesToInt(extension)
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

	return NewWSFrame(fin, opcode, payload_size, mask, payload), nil
}

func NewWSFrame(fin bool, opcode opcode.OpCode, size int, mask []byte, payload []byte) *WSFrame {
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
	fmt.Printf("fin:     %t\n", frame.Fin)
	fmt.Printf("opcode:  %s\n", frame.Opcode)
	fmt.Printf("size:    %d \n", frame.Size)
	fmt.Printf("mask:    %b\n", frame.Mask)

	fmt.Printf("payload: %s\n", frame.UnmaskPayload())

	fmt.Println("---  MESSAGE END  ---")
}

func (frame WSFrame) PrintNetwork() {
	fmt.Printf("%s - %d bytes\n", frame.Opcode, frame.Size)
}

func (frame WSFrame) UnmaskPayload() []byte {
	data := make([]byte, frame.Size)
	copy(data, frame.Payload)
	for i := 0; i < frame.Size; i++ {
	 	data[i] ^= frame.Mask[i % 4]
	}
	return data;
}

func (frame WSFrame) ToSendableBytes() []byte {
	b := []byte{}

	// FIN + opcode
	if frame.Fin {
		b = append(b, (0b1000 << 4) | byte(frame.Opcode))
	} else {
		b = append(b, (0b0000 << 4) | byte(frame.Opcode))
	}

	// TODO: Handle over 125 size
	if frame.Size >= EXTENDED_16 {
		panic("TOO BIG")
	}

	// payload size
	b = append(b, byte(frame.Size))

	// Unmasked payload
	b = append(b, frame.UnmaskPayload()...)

	return b
}