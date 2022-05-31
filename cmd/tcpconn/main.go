package main

import (
	"fmt"
	"github.com/networking-with-go/pkg/tcp"
)

func main() {
	err := tcp.TcpConn()
	if err != nil {
		fmt.Println(err)
		return
	}

}
