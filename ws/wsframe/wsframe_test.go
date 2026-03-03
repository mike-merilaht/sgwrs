package wsframe

import (
	"bytes"
	"testing"
	"net"
	"time"
	"fmt"
	"smwdd.io/sgwrs/ws/opcode"
)

type connTester struct {
	reader *bytes.Reader
}

func newConnTester(data []byte) *connTester {
	return &connTester{
		reader: bytes.NewReader(data),
	}
}

func (c *connTester) Read(b []byte) (n int, err error) {
	return c.reader.Read(b)
}

func (c *connTester) Write(b []byte) (n int, err error) {
	return 0, nil
}

func (c *connTester) Close() error {
	
	return nil
}

func (c *connTester) LocalAddr() net.Addr {
	return nil
}

func (c *connTester) RemoteAddr() net.Addr {
	return nil
}

func (c *connTester) SetDeadline(t time.Time) error {
	return nil
}

func (c *connTester) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *connTester) SetWriteDeadline(t time.Time) error {
	return nil
}

func TestFrameExtraction_EmptyFrame(t *testing.T) {
	data := []byte{0x81, 0x00, 0x00, 0x00, 0x00, 0x00}
	fmt.Printf("%b\n", data)

	conn := newConnTester(data)

	frame, err := ExtractFrame(conn)
	if err != nil {
		t.Fatal(err)
	}

	if frame.Fin != true {
		t.Errorf("expected FIN=true")
	}

	if frame.Opcode == opcode.OpCodeContinuation {
		t.Errorf("expected Opcode == OpCodeContinuation")
	}

	if frame.Size != 0 {
		t.Errorf("expected frameSize == 0, got %d", frame.Size)
	}
}