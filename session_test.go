package main

import (
	"testing"
)

func TestValidatePath(t *testing.T) {
	// Setup test session
	server := &FTPServer{
		rootJail: "/tmp/ftp-jail",
	}

	session := &ClientSession{
		server:     server,
		currentDir: "/home/user",
	}

	tests := []struct {
		name        string
		userPath    string
		currentDir  string
		expected    string
		shouldError bool
	}{
		// Valid relative paths
		{
			name:       "simple relative file",
			userPath:   "file.txt",
			currentDir: "/home/user",
			expected:   "/tmp/ftp-jail/home/user/file.txt",
		},
		{
			name:       "relative subdirectory",
			userPath:   "documents/test.pdf",
			currentDir: "/home/user",
			expected:   "/tmp/ftp-jail/home/user/documents/test.pdf",
		},
		{
			name:       "relative path with dot",
			userPath:   "./file.txt",
			currentDir: "/home/user",
			expected:   "/tmp/ftp-jail/home/user/file.txt",
		},

		// Valid absolute paths (relative to jail)
		{
			name:       "absolute path from jail root",
			userPath:   "/uploads/file.txt",
			currentDir: "/home/user",
			expected:   "/tmp/ftp-jail/uploads/file.txt",
		},
		{
			name:       "root directory",
			userPath:   "/",
			currentDir: "/home/user",
			expected:   "/tmp/ftp-jail",
		},

		// Valid navigation within jail
		{
			name:       "relative parent directory",
			userPath:   "../shared/file.txt",
			currentDir: "/home/user",
			expected:   "/tmp/ftp-jail/home/shared/file.txt",
		},

		// Security violations - should fail
		{
			name:        "escape attempt with relative path",
			userPath:    "../../../etc/passwd",
			currentDir:  "/home/user",
			shouldError: true,
		},
		{
			name:        "escape attempt with absolute path",
			userPath:    "/../../../etc/passwd",
			currentDir:  "/home/user",
			shouldError: true,
		},
		{
			name:        "multiple escape attempts",
			userPath:    "../../../../../../../../etc/passwd",
			currentDir:  "/",
			shouldError: true,
		},
		{
			name:        "escape from deep directory",
			userPath:    "../../../../../../etc/passwd",
			currentDir:  "/home/user/documents/projects/go",
			shouldError: true,
		},

		// Edge cases
		{
			name:       "empty path",
			userPath:   "",
			currentDir: "/home/user",
			expected:   "/tmp/ftp-jail/home/user",
		},
		{
			name:       "current directory reference",
			userPath:   ".",
			currentDir: "/home/user",
			expected:   "/tmp/ftp-jail/home/user",
		},
		{
			name:       "current directory from root",
			userPath:   ".",
			currentDir: "/",
			expected:   "/tmp/ftp-jail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the current directory for this test
			session.currentDir = tt.currentDir

			result, err := session.validatePath(tt.userPath)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for path %q, but got none. Result: %s", tt.userPath, result)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for path %q: %v", tt.userPath, err)
				}
				if result != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, result)
				}
			}
		})
	}
}

func TestValidatePathDifferentJails(t *testing.T) {
	tests := []struct {
		name     string
		jail     string
		userPath string
		expected string
	}{
		{
			name:     "different jail directory",
			jail:     "/var/ftp",
			userPath: "/uploads/test.txt",
			expected: "/var/ftp/uploads/test.txt",
		},
		{
			name:     "jail with trailing slash",
			jail:     "/tmp/ftp-jail/",
			userPath: "/test.txt",
			expected: "/tmp/ftp-jail/test.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &FTPServer{rootJail: tt.jail}
			session := &ClientSession{
				server:     server,
				currentDir: "/",
			}

			result, err := session.validatePath(tt.userPath)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
