package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func handlerUSER(cs *ClientSession, args []string) error {
	if len(args) != 1 {
		return cs.sendFTPResponse(501)
	}
	cs.username = args[0]
	cs.isAuth = false // reset auth state
	return cs.sendFTPResponse(331)
}

func handlerPASS(cs *ClientSession, args []string) error {
	// update this function to actually handle authentication
	if cs.username == "" {
		return cs.sendFTPResponse(530)
	}
	if cs.username == "anonymous" {
		cs.isAuth = true
		return cs.sendFTPResponse(230)
	}

	if len(args) != 1 {
		return cs.sendFTPResponse(501)
	}

	return cs.sendFTPResponse(530)
}

func handlerHELP(cs *ClientSession, args []string) error {
	var lines []string
	lines = append(lines, "The following commands are recognized:\r\n")

	for _, cmd := range serverRegistry {
		lines = append(lines, fmt.Sprintf("%s - %s", cmd.name, cmd.description))
	}
	cs.sendMultilineResponse(214, lines, "Help OK.")
	return nil
}

func handlePWD(cs *ClientSession, args []string) error {
	if !cs.isAuth {
		return cs.sendFTPResponse(530)
	}
	return cs.sendFTPResponse(257, fmt.Sprintf("\"%s\" is current directory", cs.currentDir))
}

func handleCWD(cs *ClientSession, args []string) error {
	if !cs.isAuth {
		return cs.sendFTPResponse(530)
	}

	if len(args) != 1 {
		return cs.sendFTPResponse(501)
	}

	// validate the path first (security check)
	fullPath, err := cs.validatePath(args[0])
	if err != nil {
		return cs.sendFTPResponse(550) // file not available
	}

	// check if the directory exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return cs.sendFTPResponse(550) // doesn't exist
	} else if err != nil {
		return cs.sendFTPResponse(550) // other filesystem error
	}

	// update sessions CWD
	newDir := strings.TrimPrefix(fullPath, cs.server.rootJail)
	if newDir == "" {
		newDir = "/"
	}
	cs.currentDir = newDir

	return cs.sendFTPResponse(250) // requested file action okay, completed
}

func handleCDUP(cs *ClientSession, args []string) error {
	// used the CWD handler to accomplish this
	return handleCWD(cs, []string{".."})
}

func parsePort(portStr string) (int, int, error) {
	portInt, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, 0, err
	}
	// return the high octet followed by the low octet
	return portInt / 256, portInt % 256, nil
}

func handlePASV(cs *ClientSession, args []string) error {
	if !cs.isAuth {
		return cs.sendFTPResponse(530)
	}
	dataListener, err := net.Listen("tcp", "0.0.0.0:")
	if err != nil {
		return cs.sendFTPResponse(425) // can't open data connection
	}
	cs.dataListener = dataListener

	serverAddr := cs.conn.LocalAddr().String()
	serverIP, _, _ := net.SplitHostPort(serverAddr)

	addr := cs.dataListener.Addr().String()
	_, portStr, _ := net.SplitHostPort(addr)

	portHigh, portLow, err := parsePort(portStr)
	if err != nil {
		return cs.sendFTPResponse(425) // error parsing port
	}

	ipParts := strings.Split(serverIP, ".")

	pasvResponse := fmt.Sprintf("Entering Passive Mode (%s,%s,%s,%s,%d,%d)",
		ipParts[0], ipParts[1], ipParts[2], ipParts[3], portHigh, portLow)

	return cs.sendFTPResponse(227, pasvResponse)
}

func handleLIST(cs *ClientSession, args []string) error {
	if !cs.isAuth {
		return cs.sendFTPResponse(530)
	}

	if cs.dataListener == nil {
		return cs.sendFTPResponse(425) // no data connection
	}

	reqPath := "."
	if len(args) == 1 {
		reqPath = args[0]
	}

	fullPath, err := cs.validatePath(reqPath)
	if err != nil {
		return cs.sendFTPResponse(550)
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return cs.sendFTPResponse(550)
	}

	err = cs.sendFTPResponse(150) // opening data connection
	if err != nil {
		return err
	}

	dataConn, err := cs.dataListener.Accept()
	if err != nil {
		return cs.sendFTPResponse(425) // unable to make data connection
	}
	defer dataConn.Close()
	defer cs.dataListener.Close()
	cs.dataListener = nil

	//send directory listing over data connection

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		perms := formatPermissions(info.Mode())
		size := info.Size()
		modTime := info.ModTime().Format("Jan 02 15:04")

		line := fmt.Sprintf("%s %3d %-8s %-8s %8d %s %s\r\n",
			perms,
			1,
			"user",
			"group",
			size,
			modTime,
			entry.Name())
		dataConn.Write([]byte(line))
	}

	return cs.sendFTPResponse(226)
}

func formatPermissions(mode os.FileMode) string {
	perms := make([]rune, 10)

	// File type
	if mode.IsDir() {
		perms[0] = 'd'
	} else {
		perms[0] = '-'
	}

	// Owner permissions
	perms[1] = permChar(mode, 0400, 'r')
	perms[2] = permChar(mode, 0200, 'w')
	perms[3] = permChar(mode, 0100, 'x')

	// Group permissions
	perms[4] = permChar(mode, 0040, 'r')
	perms[5] = permChar(mode, 0020, 'w')
	perms[6] = permChar(mode, 0010, 'x')

	// Other permissions
	perms[7] = permChar(mode, 0004, 'r')
	perms[8] = permChar(mode, 0002, 'w')
	perms[9] = permChar(mode, 0001, 'x')

	return string(perms)
}

func permChar(mode os.FileMode, mask os.FileMode, char rune) rune {
	if mode&mask != 0 {
		return char
	}
	return '-'
}

func handlerQUIT(cs *ClientSession, args []string) error {
	cs.sendFTPResponse(221)
	return fmt.Errorf("client requested quit")
}
