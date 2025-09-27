package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// logf prints a timestamped log message
func logf(format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] %s\n", timestamp, fmt.Sprintf(format, args...))
}

func main() {
	done := make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logf("Shutdown signal received, stopping server...")
		cancel()
		close(done)
	}()

	listener, err := net.Listen("tcp", "0.0.0.0:2121")
	if err != nil {
		log.Fatal(err)
	}
	fileManager := &FileManager{
		rootJail:        "/tmp/ftp-jail", // consider adding a command line flag
		activeUploads:   make(map[string]string),
		activeDownloads: make(map[string]string),
	}
	server := &FTPServer{
		listener:    listener,
		fileManager: fileManager,
	}
	logf("Starting FTP server on %s", listener.Addr().String())

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
			logf("Accepted connection from %s", conn.RemoteAddr().String())

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
