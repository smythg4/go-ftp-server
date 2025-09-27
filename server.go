package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type FTPServer struct {
	listener    net.Listener
	fileManager *FileManager
}

type FileManager struct {
	activeUploads   map[string]string // filename -> clientID
	activeDownloads map[string]string // filename -> clientID
	mutex           sync.RWMutex
	rootJail        string // the jail root directory
}

func (fm *FileManager) ReserveUpload(filename, clientID string) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	if existingClient, exists := fm.activeUploads[filename]; exists {
		return fmt.Errorf("file busy (client %s)", existingClient)
	}
	fm.activeUploads[filename] = clientID
	return nil
}

func (fm *FileManager) ReleaseUpload(fullPath, clientID string) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()
	delete(fm.activeUploads, fullPath)
}

func (fm *FileManager) GetSize(fullPath string) (int64, error) {
	// Do I/O operations outside critical section
	fileInfo, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return 0, fmt.Errorf("file not found")
	} else if err != nil {
		return 0, fmt.Errorf("file access error")
	}

	// Only check active operations under mutex
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()

	if existingClient, exists := fm.activeUploads[fullPath]; exists {
		return 0, fmt.Errorf("file busy (upload in progress by client %s)", existingClient)
	}

	return fileInfo.Size(), nil
}

func (fm *FileManager) ReserveDownload(fullPath, clientID string) (*os.File, error) {
	// Do I/O operations outside critical section
	fileInfo, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found")
	} else if err != nil {
		return nil, fmt.Errorf("file access error")
	}

	// don't allow downloading directories
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("cannot download directories")
	}

	// Check and reserve under mutex
	fm.mutex.Lock()
	if existingClient, exists := fm.activeUploads[fullPath]; exists {
		fm.mutex.Unlock()
		return nil, fmt.Errorf("file busy (upload in progress by client %s)", existingClient)
	}

	// reserve for download
	fm.activeDownloads[fullPath] = clientID
	fm.mutex.Unlock()

	// Open file outside mutex
	file, err := os.Open(fullPath)
	if err != nil {
		// If open fails, clean up reservation
		fm.mutex.Lock()
		delete(fm.activeDownloads, fullPath)
		fm.mutex.Unlock()
		return nil, fmt.Errorf("failed to open file")
	}
	return file, nil
}

func (fm *FileManager) ReleaseDownload(fullPath, clientID string) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()
	delete(fm.activeDownloads, fullPath)
}

func (fm *FileManager) validatePath(userPath, currentDir string) (string, error) {
	// Security check FIRST - before any path operations
	if strings.Contains(userPath, "..") {
		// Now we need to validate if this .. usage is legitimate
		// Split into components and simulate the path traversal
		currentDepth := len(strings.Split(strings.Trim(currentDir, "/"), "/"))
		if currentDir == "/" {
			currentDepth = 0
		}
		components := strings.Split(userPath, "/")
		depth := currentDepth
		for _, comp := range components {
			if comp == ".." {
				depth--
				if depth < 0 {
					return "", fmt.Errorf("path traversal attempt detected")
				}
			} else if comp != "." && comp != "" {
				depth++
			}
		}
	}
	// Now do the normal path operations
	path := filepath.Clean(userPath)
	if !strings.HasPrefix(userPath, "/") {
		path = filepath.Join(currentDir, path)
	}
	fullPath := filepath.Join(fm.rootJail, path)
	return fullPath, nil
}

func (fm *FileManager) WriteFile(fullPath, tempPath string) error {
	return os.Rename(tempPath, fullPath) // atomic rename
}

func (fm *FileManager) MakeDirectory(fullPath string) error {
	return os.MkdirAll(fullPath, os.ModeDir|0755)
}

func (fm *FileManager) DeleteFile(fullPath, clientID string) error {
	// Do I/O validation outside critical section
	fileInfo, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("file not found")
	} else if err != nil {
		return fmt.Errorf("file access error")
	}

	// don't allow deleting directories
	if fileInfo.IsDir() {
		return fmt.Errorf("cannot delete directories")
	}

	// Check and reserve under mutex
	fm.mutex.Lock()
	if existingClient, exists := fm.activeUploads[fullPath]; exists {
		fm.mutex.Unlock()
		return fmt.Errorf("file busy (upload in progress by client %s)", existingClient)
	}
	if existingClient, exists := fm.activeDownloads[fullPath]; exists {
		fm.mutex.Unlock()
		return fmt.Errorf("file busy (download in progress by client %s)", existingClient)
	}
	// reserve for deletion
	fm.activeUploads[fullPath] = clientID
	fm.mutex.Unlock()

	// Perform actual deletion outside mutex
	err = os.Remove(fullPath)

	// Clean up reservation
	fm.mutex.Lock()
	delete(fm.activeUploads, fullPath)
	fm.mutex.Unlock()

	if err != nil {
		return fmt.Errorf("system error deleting file")
	}
	return nil
}
