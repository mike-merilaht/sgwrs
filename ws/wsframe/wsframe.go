package wsframe

import (
	"fmt"
	"smwdd.io/sgwrs/ws/opcode"
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