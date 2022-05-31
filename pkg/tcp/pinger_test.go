package tcp

import (
	"context"
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

// shows how the pinger runs, uses channels to reset
func TestPinger(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	r, w := io.Pipe()

	done := make(chan struct{})

	resetTimer := make(chan time.Duration, 1)
	resetTimer <- time.Second

	go func() {
		Pinger(ctx, w, resetTimer)
		time.Sleep(time.Second * 5)
		close(done)
	}()

	receivePing := func(d time.Duration, r io.Reader) {
		if d >= 0 {
			fmt.Printf("resetting timer (%s)\n", d)
			resetTimer <- d
		}

		now := time.Now()
		buf := make([]byte, 1024)
		n, err := r.Read(buf)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Printf("received %q (%s)\n", buf[:n], time.Since(now).Round(100*time.Millisecond))
	}

	for i, v := range []int64{0, 200, 300, 0, -1, -1, -1} {
		fmt.Printf("Run %d: \n", i+1)
		receivePing(time.Duration(v)*time.Millisecond, r)
	}

	cancel()
	<-done

}

func TestPingerAdvancedDeadline(t *testing.T) {
	done := make(chan struct{})

	listener, err := net.Listen("tcp", "localhost:")
	if err != nil {
		t.Fatal(err)
	}

	begin := time.Now()

	go func() {
		defer func() { close(done) }()

		conn, err := listener.Accept()
		if err != nil {
			t.Log(err)
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer func() {
			cancel()
			conn.Close()
		}()

		resetTimer := make(chan time.Duration, 1)
		resetTimer <- time.Second

		go Pinger(ctx, conn, resetTimer)

		err = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Error(err)
			return
		}

		buf := make([]byte, 1024)

		for {
			n, err := conn.Read(buf)
			if err != nil {
				fmt.Println(err)
			}
			//nErr, ok := err.(net.Error)
			//fmt.Println(nErr, ok)
			if err != nil {

				return
			}
			t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])
			//resetTimer <- 0
			fmt.Println("extending it")
			err = conn.SetDeadline(time.Now().Add(5 * time.Second))
			if err != nil {
				t.Error(err)
				return
			}
		}

	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = conn.Close() }()

	buf := make([]byte, 1024)

	for i := 0; i < 4; i++ {
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])
	}
	_, err = conn.Write([]byte("PONG!!!"))
	if err != nil {
		t.Fatal(err)
	}

	var i int

	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				t.Fatal(err)
			}
			t.Log(err)
			break
		}
		if i == 3 {
			i = 0
			_, err = conn.Write([]byte("PONG!!!"))
			if err != nil {
				t.Fatal(err)
			}
		}

		i++

		t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])
	}

}
