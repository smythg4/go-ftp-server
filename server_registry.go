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
			name:        "USER <username>",
			description: "Specify user name for login",
			callback:    handlerUSER,
		},
		"PASS": {
			name:        "PASS <password>",
			description: "Specify password for login",
			callback:    handlerPASS,
		},
		"HELP": {
			name:        "HELP [command]",
			description: "Display help information",
			callback:    handlerHELP,
		},
		"QUIT": {
			name:        "QUIT",
			description: "Logout and close connection",
			callback:    handlerQUIT,
		},
		"PWD": {
			name:        "PWD",
			description: "Print current working directory",
			callback:    handlePWD,
		},
		"CWD": {
			name:        "CWD <directory>",
			description: "Change working directory",
			callback:    handleCWD,
		},
		"CDUP": {
			name:        "CDUP",
			description: "Change to parent directory",
			callback:    handleCDUP,
		},
		"PASV": {
			name:        "PASV",
			description: "Enter passive mode for data connections",
			callback:    handlePASV,
		},
		"LIST": {
			name:        "LIST [directory]",
			description: "List directory contents",
			callback:    handleLIST,
		},
		"STAT": {
			name:        "STAT [pathname]",
			description: "Show server or file status",
			callback:    handleSTAT,
		},
		"RETR": {
			name:        "RETR [pathname]",
			description: "Initiate download of remote file",
			callback:    handleRETR,
		},
		"SIZE": {
			name:        "SIZE [pathname]",
			description: "Retrieve size of remote file",
			callback:    handleSIZE,
		},
		"STOR": {
			name:        "STOR [pathname]",
			description: "Upload local file to remote current working directory",
			callback:    handleSTOR,
		},
		"DELE": {
			name:        "DELE [pathname]",
			description: "Delete remote file at given file path",
			callback:    handleDELE,
		},
		"NOOP": {
			name:        "NOOP",
			description: "No operation (keepalive)",
			callback:    handleNOOP,
		},
	}
}
