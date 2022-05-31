package tcp

import (
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
)

func proxyConn(source, destination string) error {
	connSource, err := net.Dial("tcp", source)
	if err != nil {
		return err
	}
	defer connSource.Close()

	connDestination, err := net.Dial("tcp", destination)
	if err != nil {
		return err
	}
	defer connDestination.Close()

	go func() {
		_, _ = io.Copy(connSource, connDestination)
	}()

	_, err = io.Copy(connDestination, connSource)
	return err

}

func proxy(from io.Reader, to io.Writer) error {
	fromWriter, isWriter := from.(io.Writer)
	toReader, isReader := to.(io.Reader)

	if isReader && isWriter {
		go func() {
			_, _ = io.Copy(fromWriter, toReader)
		}()
	}

	_, err := io.Copy(to, from)
	return err
}
func TestProxy(t *testing.T) {
	var wg sync.WaitGroup
	server, err := net.Listen("tcp", "localhost:")
	if err != nil {
		t.Fatal(err)
	}

	wg.Add(1)

	go func() {
		defer func() { fmt.Println("closing goroutine accepting"); wg.Done() }()
		for {
			conn, err := server.Accept()
			if err != nil {
				return
			}

			go func(c net.Conn) {
				defer c.Close()
				for {
					buf := make([]byte, 1024)
					n, err := c.Read(buf)
					if err != nil {
						if err != io.EOF {
							t.Error(err)

						}
						return
					}

					switch msg := string(buf[:n]); msg {
					case "ping":
						_, err = c.Write([]byte("pong"))
					default:
						_, err = c.Write(buf[:n])
					}
					if err != nil {
						if err != io.EOF {
							t.Error(err)
						}
						return
					}

				}
			}(conn)
		}
	}()

	proxyServer, err := net.Listen("tcp", "localhost:")
	if err != nil {
		t.Fatal(err)
	}
	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			conn, err := proxyServer.Accept()
			if err != nil {
				return
			}

			go func(from net.Conn) {
				defer from.Close()
				to, err := net.Dial("tcp", server.Addr().String())
				if err != nil {
					t.Error(err)
					return
				}

				defer to.Close()

				err = proxy(from, to) // copies both reader and writer buffers with the server it proxied too
				if err != nil && err != io.EOF {
					t.Error(err)
				}

			}(conn)
		}
	}()

	conn, err := net.Dial("tcp", proxyServer.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	msgs := []struct{ Message, Reply string }{
		{"ping", "pong"},
		{"pong", "pong"},
		{"echo", "echo"},
		{"ping", "pong"},
	}

	for i, m := range msgs {
		_, err = conn.Write([]byte(m.Message))
		if err != nil {
			t.Fatal(err)
		}

		buf := make([]byte, 1024)

		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}

		actual := string(buf[:n])
		t.Logf("%q -> proxy -> %q", m.Message, actual)
		if actual != m.Reply {
			t.Errorf("%d: expected reply: %q; actual %q", i, m.Reply, actual)
		}

	}

	_ = conn.Close()
	_ = proxyServer.Close()
	_ = server.Close()
	wg.Wait()

}
