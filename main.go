package main

import (
	"context"
	"fmt"
	"time"
)

func c(ctx context.Context) {
	<-ctx.Done()
	fmt.Println("received done")

}

func main() {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*5))
	defer cancel()

	c(ctx)

	//quit := make(chan struct{})
	//go func() {
	//	time.Sleep(time.Second * 5)
	//	//close(quit)
	//	quit <- struct{}{}
	//}()
	//for {
	//	select {
	//	case <-quit:
	//		fmt.Println("quitting")
	//		return
	//	default:
	//
	//	}
	//}

	//listener, err := net.Listen("tcp", "127.0.0.1:0")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//done := make(chan struct{})
	//wg := sync.WaitGroup{}
	//wg.Add(1)
	//go func() {
	//	defer wg.Done()
	//	<-done
	//	fmt.Println("closing listener")
	//	listener.Close()
	//}()
	//log.Printf("bound to %v \n", listener.Addr())
	////go func() {
	//
	//for {
	//	conn, err := listener.Accept()
	//	if err != nil {
	//		fmt.Println("errored out of ", conn.LocalAddr().String())
	//		log.Fatal(err)
	//
	//	}
	//	go func(c net.Conn) {
	//		defer func() {
	//			fmt.Println("closing connection")
	//			c.Close()
	//		}()
	//		_, err := c.Write([]byte("hello you are here"))
	//		if err != nil {
	//
	//			return
	//		}
	//		bs := make([]byte, 50)
	//		for {
	//
	//			n2, err := c.Read(bs)
	//			if err != nil {
	//				return
	//			}
	//			fmt.Println("received: ", string(bs[:n2]))
	//
	//		}
	//	}(conn)
	//
	//}
	////}()

	//tcpconn, err := net.Dial("tcp", listener.Addr().String())
	//if err != nil {
	//	log.Fatal(err)
	//}
	//tcpconn.Write([]byte("sending a msg"))
	//
	//time.Sleep(time.Second * 10)
	//tcpconn.Close()
	//done <- struct{}{}
	//wg.Wait()

}
