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
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()

	// check if the file exists and is readable
	fileInfo, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return 0, fmt.Errorf("file not found")
	} else if err != nil {
		return 0, fmt.Errorf("file access error")
	}

	if existingClient, exists := fm.activeUploads[fullPath]; exists {
		return 0, fmt.Errorf("file busy (upload in progress by client %s)", existingClient)
	}

	return fileInfo.Size(), nil
}

func (fm *FileManager) ReserveDownload(fullPath, clientID string) (*os.File, error) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	// check if the file exists and is readable
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

	if existingClient, exists := fm.activeUploads[fullPath]; exists {
		return nil, fmt.Errorf("file busy (upload in progress by client %s)", existingClient)
	}

	// reserve for download
	fm.activeDownloads[fullPath] = clientID
	return os.Open(fullPath)
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
