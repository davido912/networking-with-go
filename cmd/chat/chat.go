package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Users map[string]*net.Conn

var (
	mu    *sync.Mutex
	users Users = make(map[string]*net.Conn)
)

func server() (net.Listener, error) {
	return net.Listen("tcp", "localhost:57751")
}

func acceptConnection(srv net.Listener) {
	for {
		conn, err := srv.Accept()
		if err != nil {
			fmt.Println("hit error with connection, closing")
			_ = conn.Close()
		}
		go func() {
			_, err = conn.Write([]byte("Please choose nickname: "))
			if err != nil {
				_ = conn.Close()
			}

			resp := make([]byte, 1024)

			n, err := conn.Read(resp)
			if err != nil {
				_ = conn.Close()
			}
			users[string(resp[:n])] = &conn
			fmt.Printf("current users %+v \n", users)

			user_resp := []byte("online users:")

			for k, _ := range users {
				user_resp = append(user_resp, k...)
			}
			conn.Write(user_resp)

			n, err = conn.Read(resp)
			if err != nil {
				_ = conn.Close()
			}

			target_user := string(resp[:n])
			conn.Write([]byte(fmt.Sprintf("pairing with %q", target_user)))

			if len(users) > 1 {

				go func() {

					uconn := *users[target_user]
					defer func() {
						fmt.Println("closing last goroutine")
						conn.Close()
						uconn.Close()

					}()

					for {
						n, err = conn.Read(resp)
						if err != nil {
							fmt.Println("conn hit err ", err)
							return

						}
						_, err := uconn.Write(resp[:n])
						if err != nil {
							fmt.Println("err hit ", err)
							return
						}
					}
				}()
			}
		}()

	}
}

func main() {
	done := make(chan os.Signal)

	serv, err := server()

	fmt.Println("running on ", serv.Addr().String())
	if err != nil {
		fmt.Println("err ", err)
		return
	}

	defer func() {
		fmt.Println("closing server")
		serv.Close()
		close(done)
	}()

	go func() {
		signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)

	}()
	go acceptConnection(serv)

	<-done
	fmt.Println("received sigterm")

}
