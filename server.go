package main

import (
	"net"
)

type FTPServer struct {
	listener net.Listener
	rootJail string // the jail root directory
}
