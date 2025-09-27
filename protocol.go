package main

import (
	"fmt"
	"strings"
)

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
	var responseText string
	if len(message) > 0 {
		responseText = message[0]
	} else {
		responseText = ftpResponses[code]
	}

	// Log the response we're sending
	cs.logf("Response: %d %s", code, responseText)

	return cs.sendResponse(fmt.Sprintf("%d %s\r\n", code, responseText))
}
func (cs *ClientSession) sendMultilineResponse(code int, lines []string, finalMessage string) error {
	// Log the multi-line response
	cs.logf("Response: %d-%s (+ %d more lines) %s", code, lines[0], len(lines)-1, finalMessage)

	// Send initial line with dash
	cs.sendResponse(fmt.Sprintf("%d-%s\r\n", code, lines[0]))

	// Send middle lines (no code prefix)
	for i := 1; i < len(lines); i++ {
		cs.sendResponse(fmt.Sprintf(" %s\r\n", lines[i]))
	}

	// Send final line with space
	return cs.sendResponse(fmt.Sprintf("%d %s\r\n", code, finalMessage))
}

func (cs *ClientSession) sendResponse(resp string) error {
	_, err := cs.conn.Write([]byte(resp))
	return err
}
