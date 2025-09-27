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
	}
}
