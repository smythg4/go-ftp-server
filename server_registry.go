package main

type serverCommand struct {
	callback    func(*ClientSession, []string) error
	description string
	name        string
}

var serverRegistry map[string]serverCommand

func init() {
	serverRegistry = map[string]serverCommand{
		"USER": {
			name:        "USER",
			description: "Specify user name for authentication",
			callback:    handlerUSER,
		},
		"PASS": {
			name:        "PASS",
			description: "Submit password for authentication",
			callback:    handlerPASS,
		},
		"HELP": {
			name:        "HELP",
			description: "Display available commands",
			callback:    handlerHELP,
		},
		"QUIT": {
			name:        "QUIT",
			description: "Disconnect from server",
			callback:    handlerQUIT,
		},
		"PWD": {
			name:        "PWD",
			description: "Print working directory.",
			callback:    handlePWD,
		},
		"CWD": {
			name:        "CWD",
			description: "Change working directory.",
			callback:    handleCWD,
		},
		"CDUP": {
			name:        "CDUP",
			description: "Change to parent directory.",
			callback:    handleCDUP,
		},
		"PASV": {
			name:        "PASV",
			description: "Requests the server-DTP to 'listen' on a data port and to wait for a connection rather",
			callback:    handlePASV,
		},
		"LIST": {
			name:        "LIST",
			description: "Sends list from the server to the passive DTP.  If the pathname specifies a directory or other group of files, the server should transfer a list of files in the specified directory.",
			callback:    handleLIST,
		},
	}
}
