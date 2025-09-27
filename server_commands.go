package main

import "fmt"

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

func handlerQUIT(cs *ClientSession, args []string) error {
	cs.sendFTPResponse(221)
	return fmt.Errorf("client requested quit")
}
