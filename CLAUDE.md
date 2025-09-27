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
- **`server_commands.go`** - Implementation of FTP command handlers (USER/PASS/PWD/CWD/LIST/RETR/STOR/SIZE/STAT)
- **`session.go`** - ClientSession struct and core session management including security
- **`protocol.go`** - FTP command parsing and response utilities
- **`server.go`** - FTPServer struct and FileManager with concurrency control
- **`ftp_responses.go`** - RFC 959 compliant response codes and messages
- **`session_test.go`** - Unit tests for path validation security

### Core Components

#### Concurrent Session Architecture
- **FTPServer**: Manages TCP listener and FileManager with concurrency control
- **FileManager**: Handles file operations, concurrent access control, and security jail
- **ClientSession**: Per-connection state tracking (authentication, current directory, client ID, data connections)
- **Goroutine-per-connection**: Each client gets isolated session with proper cleanup
- **Graceful shutdown**: Signal handling (SIGINT/SIGTERM) with context cancellation

#### Data Connection Management
- **PASV Support**: Server creates data listeners for passive mode transfers
- **IP Address Handling**: Correctly formats external IP addresses for client connections
- **Connection Lifecycle**: Data connections created per-operation, properly cleaned up
- **File Transfers**: LIST, RETR operations over data connection with proper cleanup

#### Security Architecture (Jail System)
- **FileManager**: Centralized security and concurrency control in `server.go`
- **rootJail**: Server-level directory restriction (`/tmp/ftp-jail`)
- **currentDir**: Session-relative path tracking (always starts with "/")
- **Path Validation**: `validatePath()` prevents directory traversal attacks
- **Depth Tracking**: Algorithm simulates directory traversal to catch escape attempts
- **Concurrent Access Control**: File reservation system prevents upload/download conflicts
- **Authentication**: Anonymous login flow (extensible for real authentication)

### FTP Protocol Implementation

#### Complete Command Set
- **USER/PASS**: Authentication flow with anonymous support
- **PWD**: Print working directory relative to jail
- **CWD**: Change directory with security validation
- **CDUP**: Change to parent directory (uses CWD internally)
- **PASV**: Passive mode data connection setup
- **LIST**: Directory listing over data connection with Unix-style formatting
- **RETR**: File download with concurrent access control and atomic operations
- **STOR**: File upload to current directory with temporary file + atomic rename
- **SIZE**: File size query over control connection (RFC 3659)
- **STAT**: Server status (no args) or file status (with args - returns 502 not implemented)
- **HELP**: Dynamic command listing from registry descriptions
- **QUIT**: Graceful session termination

#### Advanced Features
- **Multi-line Responses**: Proper RFC 959 dash/space formatting
- **IP Address Parsing**: Correct PASV response formatting (comma-separated, no spaces)
- **File Permissions**: Unix-style permission display in LIST output
- **Atomic Operations**: STOR uses temporary files with atomic rename for safety
- **Concurrent Access Control**: FileManager prevents conflicting file operations
- **Progress Support**: SIZE command enables client progress tracking
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

3. For data connection commands (like LIST/RETR), follow the PASV → data transfer → cleanup pattern
4. For file operations, use `cs.server.fileManager` methods for security and concurrency control
5. Use `filepath.Base()` for STOR to prevent path traversal in filenames

## Security Implementation

### Directory Jail (Fully Implemented)
- **FileManager Integration**: `validatePath()` in `server.go` prevents escape attempts
- **Depth Tracking**: Simulates directory traversal to detect `../` escape attempts
- **Absolute Path Join**: All paths resolved against `rootJail` directory
- **Clean Path Processing**: Uses `filepath.Clean()` for normalization

### Concurrency Security
- **File Reservations**: `ReserveUpload()` and `ReserveDownload()` prevent conflicts
- **Mutex Protection**: All FileManager operations are thread-safe
- **Atomic Uploads**: STOR uses temporary files with atomic rename
- **Resource Cleanup**: Proper cleanup on success and failure paths

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
# From companion client directory
./goftp -host localhost:2121 -user anonymous -pass test@example.com
```

## Known Limitations

### Missing RFC Features
- **PORT Mode**: Only PASV supported, no active mode (PORT command)
- **Extended Commands**: Limited RFC 3659 support (has SIZE, missing MDTM, MLSD)
- **Transfer Types**: No ASCII/Binary mode handling (TYPE command)
- **File Management**: No DELE (delete), MKD/RMD (make/remove directory) commands

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
- Security context (via server.fileManager reference)

### FileManager Pattern
Centralized file operations with:
- Thread-safe concurrent access control
- Security jail enforcement
- File reservation system for uploads/downloads
- Atomic operations (temporary files + rename)

### Registry Pattern
Commands follow a consistent pattern:
- Registry-based dispatch for extensibility
- Consistent error handling and response formatting
- Authentication and authorization checks per command

### Resource Management
- Goroutine-per-connection with proper cleanup
- Data connections created/destroyed per operation
- File reservations released on completion or error
- Signal handling for graceful shutdown

## Companion FTP Client

This FTP server was developed alongside a companion FTP client implementation.

**Integration Testing:**
```bash
# Terminal 1: Start server (from server directory)
go build && ./goftpserver

# Terminal 2: Connect client (from client directory)
go build && ./goftp -host localhost:2121 -user anonymous -pass test@example.com
```

**Client Features:**
- Interactive REPL with progress tracking
- Supports all server commands (RETR/STOR/LIST/SIZE/etc.)
- Real-time progress bars using SIZE command
- Graceful error handling and proper FTP flow