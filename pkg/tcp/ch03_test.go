package tcp

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"syscall"
	"testing"
	"time"
)

// .Accept will usually error on client termination
func TestListener(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:3002")

	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		t.Log("closing listener")
		_ = listener.Close()

	}()

	done := make(chan struct{})

	go func() {
		defer func() { done <- struct{}{} }()
		for {
			conn, err := listener.Accept()
			if err != nil {
				t.Log("ERROR FROM ACCEPT: ", err)
				return
			}

			go func(c net.Conn) {
				defer func() {
					c.Close()
					done <- struct{}{}
				}()

				buf := make([]byte, 1024)
				for {
					n, err := c.Read(buf)
					if err != nil {
						if err != io.EOF {
							t.Error(err)
						}
						return
					}

					t.Logf("received: %q", buf[:n])

				}
			}(conn)
		}
	}()
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	conn.Write([]byte("this is the msg"))
	conn.Close()
	<-done
	listener.Close()
	<-done
}

func DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	d := net.Dialer{
		Control: func(network, address string, _ syscall.RawConn) error {
			return &net.DNSError{
				Err:         "connection timed out",
				Name:        address,
				Server:      "127.0.0.1",
				IsTimeout:   true,
				IsTemporary: true,
			}
		},
		Timeout: timeout,
	}
	return d.Dial(network, address)
}

func TestDialTimeout(t *testing.T) {
	c, err := DialTimeout("tcp", "10.0.0.1:http", 5*time.Second)
	if err == nil {
		c.Close()
		t.Fatal("connection did not time out")
	}
	nErr, ok := err.(net.Error)
	fmt.Println(nErr)
	if !ok {
		t.Fatal(err)
	}

	if !nErr.Timeout() {
		t.Fatal("error is not a timeout")
	}
}

// using context to signal timeout
func TestDialContext(t *testing.T) {
	dl := time.Now().Add(1 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), dl)
	defer cancel()

	var d net.Dialer
	d.Control = func(_, _ string, _ syscall.RawConn) error {
		//time.Sleep(6 * time.Second)
		return nil
	}

	conn, err := d.DialContext(ctx, "tcp", "10.0.0.0:80")
	if err == nil {
		conn.Close()
		t.Fatal("connection did not time out")
	}

	nErr, ok := err.(net.Error)
	if !ok {
		t.Error(err)
	} else {
		if !nErr.Timeout() {
			t.Errorf("error is not a timeout: %v\n", err)
		}

	}

	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("expected deadline exceeded, actual: %v\n", ctx.Err())
	}

}

//using context to cancel a dial call instead of timing out
func TestDialContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sync := make(chan struct{})

	go func() {
		defer func() { sync <- struct{}{} }()

		var d net.Dialer
		d.Control = func(_, _ string, _ syscall.RawConn) error {
			time.Sleep(1)
			return nil
		}
		conn, err := d.DialContext(ctx, "tcp", "10.0.0.1:80")
		if err != nil {
			t.Log(err)
			return
		}
		conn.Close()
		t.Error("connection did not time out")

	}()

	cancel()
	<-sync
	if ctx.Err() != context.Canceled {
		t.Errorf("Expected cancel context but received %q", ctx.Err())
	}
}

// cancelling several dials with one context - first arriving goroutine in the channel signals to cancel
// cancel goes through the dialcontext and cancels all goroutines still dialing
func TestDialContextCancelFanOut(t *testing.T) {
	dl := time.Now().Add(10 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), dl)

	listener, err := net.Listen("tcp", "localhost:")

	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = listener.Close() }()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			_ = conn.Close()
		}

	}()

	dial := func(ctx context.Context, address string, response chan int, id int, wg *sync.WaitGroup) {
		defer wg.Done()

		var d net.Dialer

		c, err := d.DialContext(ctx, "tcp", address)
		if err != nil {
			return
		}
		_ = c.Close()
		fmt.Println("before select ", id)
		select {
		case <-ctx.Done():
			fmt.Println("received done ", id)
		case response <- id:
			fmt.Println("inserting")
		}
	}

	res := make(chan int)

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go dial(ctx, listener.Addr().String(), res, i+1, &wg)
	}

	response := <-res
	cancel()
	wg.Wait()
	close(res)

	if ctx.Err() != context.Canceled {
		t.Errorf("expected conctext cancel but received %s\n", ctx.Err())
	}

	t.Logf("dialer %d retrieved the resource\n", response)
}

// proves how deadline can be further pushed while listening
func TestDeadilne(t *testing.T) {
	s := make(chan struct{})

	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Log(err)
			return
		}

		defer func() {
			fmt.Println("called")
			_ = conn.Close()
			close(s)
		}()

		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Error(err)
			return
		}

		buf := make([]byte, 1)
		_, err = conn.Read(buf) // blocked until remote node sends data
		nErr, ok := err.(net.Error)
		if !ok || !nErr.Timeout() {
			t.Errorf("expected timeout error, actual %v\n", err)
		}

		s <- struct{}{}

		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Error(err)
			return
		}

		_, err = conn.Read(buf)
		if err != nil {
			t.Error(err)
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	<-s
	_, err = conn.Write([]byte("1"))
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 1)
	_, err = conn.Read(buf)

	if err != io.EOF {
		t.Errorf("expected server termination; actual %v\n", err)
	}
}
