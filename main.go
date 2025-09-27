package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	done := make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		cancel()
		close(done)
	}()

	listener, err := net.Listen("tcp", "0.0.0.0:2121")
	if err != nil {
		log.Fatal(err)
	}
	server := &FTPServer{
		listener: listener,
		rootJail: "/tmp/ftp-jail", // consider adding a command line flag
	}
	fmt.Printf("Starting FTP server at: %s...\n", listener.Addr().String())

	go func() {
		go func() {
			<-ctx.Done()
			_ = listener.Close()
		}()
		for {
			conn, err := listener.Accept()

			if err != nil {
				return
			}
			fmt.Printf("Accepted connection from: %s\n", conn.RemoteAddr().String())

			session := &ClientSession{
				conn:         conn,
				clientID:     conn.RemoteAddr().String(),
				username:     "",
				currentDir:   "/",
				isAuth:       false,
				dataListener: nil,
				server:       server,
			}

			go session.run()
		}
	}()
	<-done
}
