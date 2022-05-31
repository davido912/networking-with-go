package echoserver

import (
	"context"
	"fmt"
	"net"
)

func EchoServerUDP(ctx context.Context, addr string) (net.Addr, error) {
	s, err := net.ListenPacket("udp", addr)
	if err != nil {
		return nil, err
	}

	go func() {
		go func() {

			<-ctx.Done()
			fmt.Println("closing")
			_ = s.Close()
		}()
		buf := make([]byte, 1024)

		for {
			n, clientAddr, err := s.ReadFrom(buf)
			if err != nil {
				return
			}

			_, err = s.WriteTo(buf[:n], clientAddr)
			if err != nil {
				return
			}
		}
	}()

	return s.LocalAddr(), nil
}
