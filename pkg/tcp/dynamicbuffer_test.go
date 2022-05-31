package tcp

import (
	"bytes"
	"encoding/binary"
	"net"
	"reflect"
	"testing"
)

func TestPayloads(t *testing.T) {
	b1 := Binary("This is a sentence.")
	b2 := Binary("Dont stop")
	s1 := String("Errors are okay.")

	payloads := []Payload{&b1, &s1, &b2}

	listener, err := net.Listen("tcp", "localhost:")
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()

		for _, p := range payloads {
			_, err = p.WriteTo(conn)
			if err != nil {
				t.Error(err)
				break
			}
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	defer conn.Close()

	for i := 0; i < len(payloads); i++ {
		actual, err := decode(conn)
		if err != nil {
			t.Fatal(err)
		}

		if expected := payloads[i]; !reflect.DeepEqual(expected, actual) {
			t.Errorf("value mismatch, got %v but expected %v", actual, expected)
			continue
		}

		t.Logf("[%T] %[1]q", actual)
	}

}

func TestMaxPayloadsSize(t *testing.T) {
	buf := new(bytes.Buffer)
	err := buf.WriteByte(BinaryType)
	if err != nil {
		t.Fatal(err)
	}

	err = binary.Write(buf, binary.BigEndian, uint32(1<<30)) // 1 gb
	if err != nil {
		t.Fatal(err)
	}

	var b Binary
	_, err = b.ReadFrom(buf)

	if err != ErrMaxPayloadSize {
		t.Fatalf("expected ErrMaxPayloadSize; actual: %v", err)
	}
}
