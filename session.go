package main

import (
	"fmt"
	"net"
)

type ClientSession struct {
	conn         net.Conn
	clientID     string
	username     string
	isAuth       bool
	currentDir   string
	dataListener net.Listener
	server       *FTPServer
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

func (cs *ClientSession) Close() error {
	return cs.conn.Close()
}

func (cs *ClientSession) validatePath(userPath string) (string, error) {
	return cs.server.fileManager.validatePath(userPath, cs.currentDir)
}
