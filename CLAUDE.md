# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A fully functional FTP server (`goftpserver`) written in Go 1.24.2, implementing RFC 959 from first principles with concurrent session management, security jail system, data connections, and complete directory navigation.

## Development Commands

### Build and Run
```bash
go build .                    # Build the project
./goftpserver                 # Run server on 0.0.0.0:2121
```

The server listens on all interfaces (0.0.0.0:2121) with jail directory `/tmp/ftp-jail` - consider adding command line flags for production use.

### Code Quality
```bash
go fmt ./...                  # Format all Go files
go vet ./...                  # Static analysis for potential issues
go mod tidy                   # Clean up module dependencies
go test ./...                 # Run unit tests (validatePath has test coverage)
```

## Architecture Overview

### Multi-File Structure
- **`main.go`** - TCP server, signal handling, and session management
- **`server_registry.go`** - Command dispatch system with extensible registry pattern
- **`server_commands.go`** - Implementation of FTP command handlers
- **`session.go`** - ClientSession struct and core session management including security
- **`protocol.go`** - FTP command parsing and response utilities
- **`server.go`** - FTPServer struct definition
- **`ftp_responses.go`** - RFC 959 compliant response codes and messages
- **`session_test.go`** - Unit tests for path validation security

### Core Components

#### Concurrent Session Architecture
- **FTPServer**: Manages TCP listener and jail configuration (`rootJail` field)
- **ClientSession**: Per-connection state tracking (authentication, current directory, client ID, data connections)
- **Goroutine-per-connection**: Each client gets isolated session with proper cleanup
- **Graceful shutdown**: Signal handling (SIGINT/SIGTERM) with context cancellation

#### Data Connection Management
- **PASV Support**: Server creates data listeners for passive mode transfers
- **IP Address Handling**: Correctly formats external IP addresses for client connections
- **Connection Lifecycle**: Data connections created per-operation, properly cleaned up
- **LIST Implementation**: Directory listings with proper ls-la formatting over data connection

#### Security Architecture (Jail System)
- **rootJail**: Server-level directory restriction (`/tmp/ftp-jail`)
- **currentDir**: Session-relative path tracking (always starts with "/")
- **Path Validation**: `validatePath()` in `session.go:62` prevents directory traversal attacks
- **Depth Tracking**: Algorithm simulates directory traversal to catch escape attempts
- **Authentication**: Anonymous login flow (extensible for real authentication)

### FTP Protocol Implementation

#### Complete Command Set
- **USER/PASS**: Authentication flow with anonymous support
- **PWD**: Print working directory relative to jail
- **CWD**: Change directory with security validation
- **CDUP**: Change to parent directory (uses CWD internally)
- **PASV**: Passive mode data connection setup
- **LIST**: Directory listing over data connection with Unix-style formatting
- **STAT**: Server status (no args) or file status (with args - returns 502 not implemented)
- **HELP**: Dynamic command listing from registry descriptions
- **QUIT**: Graceful session termination

#### Advanced Features
- **Multi-line Responses**: Proper RFC 959 dash/space formatting
- **IP Address Parsing**: Correct PASV response formatting (comma-separated, no spaces)
- **File Permissions**: Unix-style permission display in LIST output
- **Error Handling**: Comprehensive error codes (530 auth required, 550 access denied, etc.)

#### Protocol Compliance
- **Response Codes**: Full RFC 959 compliance with proper 3-digit codes
- **Command Parsing**: Case-insensitive with proper argument handling
- **Connection Flow**: Control + data connection model correctly implemented
- **Transfer States**: Proper 150/226 sequencing for data transfers

## Adding New Commands

1. Add command entry to `serverRegistry` in `server_registry.go`:
   ```go
   "NEWCMD": {
       name:        "NEWCMD <args>",
       description: "Description for HELP output",
       callback:    handleNEWCMD,
   },
   ```

2. Implement handler function in `server_commands.go`:
   ```go
   func handleNEWCMD(cs *ClientSession, args []string) error {
       if !cs.isAuth {
           return cs.sendFTPResponse(530) // Auth required
       }
       // Command implementation
       return cs.sendFTPResponse(250) // Success
   }
   ```

3. For data connection commands (like LIST), follow the PASV → data transfer → cleanup pattern
4. Use `cs.validatePath()` for any file/directory operations

## Security Implementation

### Directory Jail (Fully Implemented)
- **Path Validation**: `validatePath()` in `session.go:62-91` prevents escape attempts
- **Depth Tracking**: Simulates directory traversal to detect `../` escape attempts
- **Absolute Path Join**: All paths resolved against `rootJail` directory
- **Clean Path Processing**: Uses `filepath.Clean()` for normalization

### Connection Security
- **Authentication Gates**: Commands check `cs.isAuth` before execution
- **Path Sanitization**: All user paths validated before filesystem operations
- **Resource Cleanup**: Data connections properly closed after use
- **Error Responses**: Security violations return appropriate FTP error codes

## Testing

### Unit Tests
- **Path Validation**: `session_test.go` covers security-critical `validatePath()` function
- **Test Coverage**: Includes escape attempt detection and legitimate navigation

### Manual Testing
Use the companion FTP client or standard FTP clients for integration testing:
```bash
# From go-ftp-client directory
./goftp -host localhost:2121 -user anonymous -pass test@example.com
```

## Known Limitations

### Missing RFC Features
- **File Transfer**: RETR/STOR commands not yet implemented
- **PORT Mode**: Only PASV supported, no active mode (PORT command)
- **Extended Commands**: No RFC 3659 extensions (MDTM, SIZE, MLSD)
- **Transfer Types**: No ASCII/Binary mode handling (TYPE command)

### Production Considerations
- **User Authentication**: Only anonymous login supported
- **Connection Limits**: No concurrent connection or rate limiting
- **TLS/Security**: No encryption support (FTPS/SFTP)
- **Logging**: Minimal logging beyond debug output
- **Configuration**: Hardcoded port and jail directory

## Architecture Patterns

### Session State Management
Each `ClientSession` maintains:
- Connection state (auth status, current directory)
- Network connections (control + data listener)
- Security context (jail root reference)

### Registry Pattern
Commands follow a consistent pattern:
- Registry-based dispatch for extensibility
- Consistent error handling and response formatting
- Authentication and authorization checks per command

### Resource Management
- Goroutine-per-connection with proper cleanup
- Data connections created/destroyed per operation
- Signal handling for graceful shutdown