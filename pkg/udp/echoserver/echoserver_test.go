package echoserver

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"testing"
	"time"
)

func TestEchoServerUDP(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	serverAddr, err := EchoServerUDP(ctx, "localhost:")

	if err != nil {
		t.Fatal(err)
	}

	defer cancel()

	client, err := net.ListenPacket("udp", "localhost:")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = client.Close() }()

	msg := []byte("ping")
	_, err = client.WriteTo(msg, serverAddr)
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 1024)
	n, addr, err := client.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	if addr.String() != serverAddr.String() {
		t.Fatalf("received reply from %q instead of %q", addr, serverAddr)
	}

	if !bytes.Equal(buf[:n], msg) {
		t.Errorf("expected reply %q, actual reply %q", msg, buf[:n])
	}
}

// an example of how a question can 'interlop' or in other words, interfere and send msgs to a client
func TestListenPacketUDP(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	serverAddr, err := EchoServerUDP(ctx, "localhost:")
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()

	client, err := net.ListenPacket("udp", "localhost:")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = client.Close() }()

	interloper, err := net.ListenPacket("udp", "localhost:")
	if err != nil {
		t.Fatal(err)
	}

	interrupt := []byte("pardon me")

	n, err := interloper.WriteTo(interrupt, client.LocalAddr())
	if err != nil {
		t.Fatal(err)
	}

	_ = interloper.Close()

	if l := len(interrupt); l != n {
		t.Fatalf("wrote %d bytes of %d", n, l)
	}

	ping := []byte("ping")
	_, err = client.WriteTo(ping, serverAddr)
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 1024)
	n, addr, err := client.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(interrupt, buf[:n]) {
		t.Errorf("expected reply %q; actual reply %q", interrupt, buf[:n])
	}

	if addr.String() != interloper.LocalAddr().String() {
		t.Errorf("expected msg from %q, actual sender is %q", interloper.LocalAddr(), addr)
	}

	n, addr, err = client.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(ping, buf[:n]) {
		t.Errorf("expected reply %q; actual reply %q", interrupt, buf[:n])
	}

	if addr.String() != serverAddr.String() {
		t.Errorf("expected msg from %q, actual sender is %q", serverAddr, addr)
	}
}

// examples how using net.Conn with UDP will guarantee no interloping from a 3rd client. must be aware though that usage of this will not
// return error if destination fails to receive packets on net.Conn write ops
func TestDialUDP(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	serverAddr, err := EchoServerUDP(ctx, "localhost:")
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()

	client, err := net.Dial("udp", serverAddr.String())
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = client.Close() }()

	interloper, err := net.ListenPacket("udp", "localhost:")
	if err != nil {
		t.Fatal(err)
	}

	interrupt := []byte("pardon me")
	n, err := interloper.WriteTo(interrupt, client.LocalAddr())
	if err != nil {
		t.Fatal(err)
	}

	_ = interloper.Close()

	if l := len(interrupt); l != n {
		t.Fatalf("wrote %d bytes of %d", n, l)
	}

	ping := []byte("ping")

	_, err = client.Write(ping)
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 1024)
	n, err = client.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(ping, buf[:n]) {
		t.Errorf("expected reply %q, actual reply %q", ping, buf[:n])
	}

	err = client.SetDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Read(buf)
	if err == nil {
		t.Fatal("unexpected packet")
	}
	fmt.Println(err)

}
