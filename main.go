package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type FTPServer struct {
	listener net.Listener
	rootJail string // the jail root directory
}

type FTPCommand struct {
	Command string
	Args    []string
	Raw     string
}

func (f *FTPCommand) String() string {
	if len(f.Args) == 0 {
		return f.Command
	}
	return fmt.Sprintf("%s %s", f.Command, strings.Join(f.Args, " "))
}

type ClientSession struct {
	conn       net.Conn
	clientID   string
	username   string
	isAuth     bool
	currentDir string
	dataAddr   string
	server     *FTPServer
}

func parseCommand(input string) (*FTPCommand, error) {
	trimmed := strings.TrimSpace(input)
	parts := strings.Fields(trimmed)
	if len(parts) < 1 {
		return nil, fmt.Errorf("empty command")
	}
	return &FTPCommand{
		Command: strings.ToUpper(parts[0]),
		Args:    parts[1:],
		Raw:     input,
	}, nil
}
func (cs *ClientSession) sendFTPResponse(code int, message ...string) error {
	if len(message) > 0 {
		return cs.sendResponse(fmt.Sprintf("%d %s\r\n", code, message[0]))
	}
	return cs.sendResponse(fmt.Sprintf("%d %s\r\n", code, ftpResponses[code]))
}
func (cs *ClientSession) sendMultilineResponse(code int, lines []string, finalMessage string) error {
	// Send initial line with dash
	cs.sendResponse(fmt.Sprintf("%d-%s\r\n", code, lines[0]))

	// Send middle lines (no code prefix)
	for i := 1; i < len(lines); i++ {
		cs.sendResponse(fmt.Sprintf(" %s\r\n", lines[i]))
	}

	// Send final line with space
	return cs.sendResponse(fmt.Sprintf("%d %s\r\n", code, finalMessage))
}

func (cs *ClientSession) run() {
	defer func() { _ = cs.conn.Close() }()
	lines := []string{
		"Welcome to the go-ftp server!",
		"\tIt's gonna be so great, you have no idea",
	}
	err := cs.sendMultilineResponse(220, lines, "Service ready for new user.")
	if err != nil {
		fmt.Printf("error sending welcome message: %v", err)
	}

	buf := make([]byte, 1024)
	for {
		n, err := cs.conn.Read(buf)
		if err != nil {
			fmt.Printf("Client %s disconnected: %v\n", cs.clientID, err)
			return
		}
		fmt.Printf(" %s: %s", cs.clientID, string(buf[:n]))

		cmd, err := parseCommand(string(buf[:n]))
		if err != nil {
			fmt.Printf("Error - %v", err)
			return
		}

		if handler, exists := serverRegistry[cmd.Command]; exists {
			err := handler.callback(cs, cmd.Args)
			if err != nil {

				if err.Error() == "client requested quit" {
					fmt.Printf("Client %s session ending: %v\n", cs.clientID, err)
					return
				}
				fmt.Printf("Error - %v\n", err)
			}
		} else {
			cs.sendFTPResponse(502)
		}
	}
}

func (cs *ClientSession) sendResponse(resp string) error {
	_, err := cs.conn.Write([]byte(resp))
	return err
}

func (cs *ClientSession) Close() error {
	return cs.conn.Close()
}

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

	listener, err := net.Listen("tcp", "127.0.0.1:2121")
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
				conn:       conn,
				clientID:   conn.RemoteAddr().String(),
				username:   "",
				currentDir: "/",
				isAuth:     false,
				dataAddr:   "",
				server:     server,
			}

			go session.run()
		}
	}()
	<-done
}
