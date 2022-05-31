package tcp

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"reflect"
	"testing"
)

// a simple payload shipping over packets and simple read.
func TestReadIntoBuffer(t *testing.T) {
	payload := make([]byte, 1<<24) // generate a 16MB size buffer
	_, err := rand.Read(payload)   // generate random payload
	if err != nil {
		t.Fatal(err)
	}

	listener, err := net.Listen("tcp", "localhost:")
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Log(err)
			return
		}

		defer conn.Close()

		_, err = conn.Write(payload)
		if err != nil {
			t.Error(err)
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 1<<19) // 512 KB
	var total int
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				t.Error(err)
			}
			if err == io.EOF {
				fmt.Println("it is EOF")
			}
			break
		}
		t.Logf("read %d bytes", n) // data read from the conn
		total += n
	}
	fmt.Println("read a total of ", total)
	conn.Close()
}

const payload = "This is the first message, this is the second message."

// using bufio to scan delimited packets
func TestScanner(t *testing.T) {
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
		_, err = conn.Write([]byte(payload))
		if err != nil {
			t.Error(err)
		}

	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	scanner.Split(bufio.ScanWords)

	var words []string

	for scanner.Scan() {
		words = append(words, scanner.Text())
	}

	err = scanner.Err()
	if err != nil {
		t.Error(err)
	}

	expected := []string{"This", "is", "the", "first", "message,", "this", "is", "the", "second", "message."}

	if !reflect.DeepEqual(words, expected) {
		t.Fatalf("innacurate scanned word list \n got %v \n expected %v", words, expected)

	}
	t.Logf("scanned words: %v", words)
}
