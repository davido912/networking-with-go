package tcp

import (
	"fmt"
	"net"
)

func TcpConn() error {
	done := make(chan struct{})
	addr, err := net.ResolveTCPAddr("tcp", "localhost:")
	if err != nil {
		return err
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}

	go func() {
		defer func() { done <- struct{}{} }()
		tcpConn, err := listener.AcceptTCP()
		if err != nil {
			fmt.Println("tcp conn error")
			return
		}

		if err != nil {
			fmt.Println("erred")
			return
		}

		bs := make([]byte, 1024)

		_, err = tcpConn.Read(bs)
		if err != nil {
			return
		}
		fmt.Println("read ", string(bs))

		//time.Sleep(time.Second * 10)
		fmt.Println(tcpConn.RemoteAddr(), tcpConn.LocalAddr())

		//defer tcpConn.Close()
	}()

	tcpaddr, err := net.ResolveTCPAddr("tcp", listener.Addr().String())
	if err != nil {
		return err
	}
	tcpconn, err := net.DialTCP("tcp", nil, tcpaddr)
	if err != nil {
		return err
	}

	n, err := tcpconn.Write([]byte("hello this is a msg that will hopefully block it"))
	if err != nil {
		return err
	}
	fmt.Println("wrote ", n)

	<-done
	_ = listener.Close()

	return nil

}
