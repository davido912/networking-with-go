// shows usage of Multiwrite and TeeReader to log all network traffic

package tcp

import (
	"io"
	"log"
	"net"
	"os"
)

type Monitor struct {
	*log.Logger
}

func (m *Monitor) Write(p []byte) (int, error) {
	return len(p), m.Output(2, string(p))
}

func ExampleMonitor() {

	monitor := &Monitor{Logger: log.New(os.Stdout, "monitor: ", 0)}

	listener, err := net.Listen("tcp", "localhost:")
	if err != nil {
		monitor.Fatal(err)
	}

	done := make(chan struct{})

	go func() {
		defer close(done)

		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		b := make([]byte, 1024)
		r := io.TeeReader(conn, monitor) // creates a reader that writes everything to the defined writer (monitor in this case)

		n, err := r.Read(b)
		if err != nil && err != io.EOF {
			monitor.Println(err)
			return
		}

		w := io.MultiWriter(conn, monitor)

		_, err = w.Write(b[:n]) // echo the message
		if err != nil && err != io.EOF {
			monitor.Println(err)
			return
		}

	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		monitor.Fatal(err)
	}

	_, err = conn.Write([]byte("Test\n"))
	if err != nil {
		monitor.Fatal(err)
	}

	_ = conn.Close()
	<-done
}
