package main

var ftpResponses = map[int]string{
	// 1xx - Positive Preliminary replies
	110: "Restart marker reply.",
	120: "Service ready in nnn minutes.",
	125: "Data connection already open; transfer starting.",
	150: "File status okay; about to open data connection.",

	// 2xx - Positive Completion replies
	200: "Command okay.",
	202: "Command not implemented, superfluous at this site.",
	211: "System status, or system help reply.",
	212: "Directory status.",
	213: "File status.",
	214: "Help message.",
	215: "NAME system type.",
	220: "Service ready for new user.",
	221: "Service closing control connection.",
	225: "Data connection open; no transfer in progress.",
	226: "Closing data connection.",
	227: "Entering Passive Mode (h1,h2,h3,h4,p1,p2).",
	230: "User logged in, proceed.",
	250: "Requested file action okay, completed.",
	257: "PATHNAME created.",

	// 3xx - Positive Intermediate replies
	331: "User name okay, need password.",
	332: "Need account for login.",
	350: "Requested file action pending further information.",

	// 4xx - Transient Negative Completion replies
	421: "Service not available, closing control connection.",
	425: "Can't open data connection.",
	426: "Connection closed; transfer aborted.",
	450: "Requested file action not taken.",
	451: "Requested action aborted: local error in processing.",
	452: "Requested action not taken. Insufficient storage space in system.",

	// 5xx - Permanent Negative Completion replies
	500: "Syntax error, command unrecognized.",
	501: "Syntax error in parameters or arguments.",
	502: "Command not implemented.",
	503: "Bad sequence of commands.",
	504: "Command not implemented for that parameter.",
	530: "Not logged in.",
	532: "Need account for storing files.",
	550: "Requested action not taken. File unavailable.",
	551: "Requested action aborted: page type unknown.",
	552: "Requested file action aborted. Exceeded storage allocation.",
	553: "Requested action not taken. File name not allowed.",
}
