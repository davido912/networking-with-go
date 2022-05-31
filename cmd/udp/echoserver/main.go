package main

import (
	"context"
	"fmt"
	"github.com/networking-with-go/pkg/udp/echoserver"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	addr, err := echoserver.EchoServerUDP(ctx, "localhost:")
	if err != nil {
		fmt.Println("hit err")
		return
	}

	fmt.Println(addr)
	time.Sleep(time.Second * 10)
	cancel()
	time.Sleep(time.Second * 10)
}
