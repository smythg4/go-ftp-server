package main

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"
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

func (cs *ClientSession) validatePath(userPath string) (string, error) {
	// Security check FIRST - before any path operations
	if strings.Contains(userPath, "..") {
		// Now we need to validate if this .. usage is legitimate
		// Split into components and simulate the path traversal
		currentDepth := len(strings.Split(strings.Trim(cs.currentDir, "/"), "/"))
		if cs.currentDir == "/" {
			currentDepth = 0
		}
		components := strings.Split(userPath, "/")
		depth := currentDepth
		for _, comp := range components {
			if comp == ".." {
				depth--
				if depth < 0 {
					return "", fmt.Errorf("path traversal attempt detected")
				}
			} else if comp != "." && comp != "" {
				depth++
			}
		}
	}
	// Now do the normal path operations
	path := filepath.Clean(userPath)
	if !strings.HasPrefix(userPath, "/") {
		path = filepath.Join(cs.currentDir, path)
	}
	fullPath := filepath.Join(cs.server.rootJail, path)
	return fullPath, nil
}

func (cs *ClientSession) Close() error {
	return cs.conn.Close()
}
