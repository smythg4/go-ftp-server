# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A concurrent FTP server (`goftpserver`) written in Go 1.24.2, implementing RFC 959 from first principles with proper session management, RFC-compliant responses, and security jail architecture.

## Development Commands

### Build and Run
```bash
go build .                    # Build the project
./goftpserver                 # Run server on 127.0.0.1:2121
```

The server hardcodes port 2121 and jail directory `/tmp/ftp-jail` - consider adding command line flags for production use.

### Code Quality
```bash
go fmt ./...                  # Format all Go files
go vet ./...                  # Static analysis for potential issues
go mod tidy                   # Clean up module dependencies
```

## Architecture Overview

### Multi-File Structure
- **`main.go`** - TCP server, signal handling, and session management
- **`server_registry.go`** - Command dispatch system with extensible registry pattern
- **`server_commands.go`** - Implementation of FTP command handlers
- **`ftp_responses.go`** - RFC 959 compliant response codes and messages

### Core Components

#### Concurrent Session Architecture
- **FTPServer**: Manages TCP listener and jail configuration (`rootJail` field)
- **ClientSession**: Per-connection state tracking (authentication, current directory, client ID)
- **Goroutine-per-connection**: Each client gets isolated session with proper cleanup
- **Graceful shutdown**: Signal handling (SIGINT/SIGTERM) with context cancellation

#### Command Processing Pipeline
1. **Command Parsing**: `parseCommand()` converts raw input to structured `FTPCommand` with uppercase normalization
2. **Registry Dispatch**: `serverRegistry` map routes commands to handler functions
3. **Response Generation**: `sendFTPResponse()` and `sendMultilineResponse()` ensure RFC compliance
4. **Error Handling**: Commands return errors; QUIT specifically returns "client requested quit" to terminate session

#### Security Architecture (Jail System)
- **rootJail**: Server-level directory restriction (currently `/tmp/ftp-jail`)
- **currentDir**: Session-relative path tracking (always starts with "/")
- **Path validation**: Future `validatePath()` method will prevent directory traversal attacks
- **Authentication**: Simple anonymous-only auth in PASS handler (extensible for real auth)

### FTP Protocol Implementation

#### RFC-Compliant Responses
- **Single-line responses**: `sendFTPResponse(code)` uses `ftpResponses` map for standard messages
- **Multi-line responses**: `sendMultilineResponse()` handles dash/space formatting per RFC
- **Standard codes**: 220 (welcome), 331 (need password), 230 (logged in), 221 (goodbye), 5xx (errors)

#### Command Registry Pattern
```go
type serverCommand struct {
    callback    func(*ClientSession, []string) error
    description string  // Used by HELP command
    name        string
}
```

#### Session Lifecycle
1. **Connection**: TCP accept creates `ClientSession` with default state
2. **Welcome**: Multi-line 220 response sent automatically
3. **Authentication**: USER → PASS flow sets `isAuth` flag
4. **Command processing**: Loop reads/parses/dispatches until error or QUIT
5. **Cleanup**: `defer conn.Close()` ensures resource cleanup

### Current Command Support
- **USER/PASS**: Anonymous authentication flow
- **HELP**: Dynamic command listing from registry descriptions
- **QUIT**: Graceful session termination
- **Unimplemented**: Returns proper 502 response code

## Adding New Commands

1. Add command entry to `serverRegistry` in `server_registry.go`
2. Implement handler function in `server_commands.go` following signature:
   ```go
   func handlerCOMMAND(cs *ClientSession, args []string) error
   ```
3. Use `cs.sendFTPResponse(code)` for standard responses
4. Use `cs.sendMultilineResponse(code, lines, finalMsg)` for complex output
5. Return non-nil error only for session-terminating conditions

## Security Considerations

### Directory Jail (Partially Implemented)
- Server has `rootJail` field but path validation not yet implemented
- All sessions start with `currentDir: "/"` (relative to jail)
- Future path operations must validate against directory traversal

### Connection Security
- No authentication beyond anonymous login
- No connection limits or rate limiting
- No TLS/encryption support
- Consider adding user account system and connection quotas

## RFC 959 Compliance

The server follows RFC 959 standards for:
- Response code formatting (3-digit codes with proper spacing/dashes)
- Multi-line response syntax
- Standard response messages with periods
- Command case-insensitivity (uppercase normalization)
- Basic authentication flow (USER/PASS sequence)

Missing RFC features:
- Data connections (PASV/PORT/LIST/RETR/STOR)
- Directory operations (CWD/CDUP/PWD/MKD/RMD)
- File operations beyond basic protocol
- Extended commands (SIZE/MDTM per RFC 3659)