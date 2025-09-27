package main

import (
	"fmt"
	"net"
	"strings"
	"time"
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
		"go-ftp-server v0.0.0.1",
		"   RFC 959 compliant. Built from scratch in Go.",
		"   Simple FTP. No frills. Just files.",
	}
	err := cs.sendMultilineResponse(220, lines, "Service ready for new user.")
	if err != nil {
		cs.logf("Error sending welcome message: %v", err)
	}

	buf := make([]byte, 1024)
	for {
		n, err := cs.conn.Read(buf)
		if err != nil {
			cs.logf("Disconnected: %v", err)
			return
		}
		cs.logf("Command: %s", strings.TrimSpace(string(buf[:n])))

		cmd, err := parseCommand(string(buf[:n]))
		if err != nil {
			cs.logf("Parse error: %v", err)
			return
		}

		if handler, exists := serverRegistry[cmd.Command]; exists {
			err := handler.callback(cs, cmd.Args)
			if err != nil {
				if err.Error() == "client requested quit" {
					cs.logf("Session ending: %v", err)
					return
				}
				cs.logf("Command error: %v", err)
			}
		} else {
			cs.sendFTPResponse(502)
		}
	}
}

func (cs *ClientSession) Close() error {
	return cs.conn.Close()
}

// logf prints a timestamped log message for this session
func (cs *ClientSession) logf(format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] %s: %s\n", timestamp, cs.clientID, fmt.Sprintf(format, args...))
}

func (cs *ClientSession) validatePath(userPath string) (string, error) {
	return cs.server.fileManager.validatePath(userPath, cs.currentDir)
}
