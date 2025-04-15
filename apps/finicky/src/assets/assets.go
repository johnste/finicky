package assets

import (
	"embed"
	"io/fs"
)

//go:embed templates/*
var content embed.FS

// GetHTML returns the HTML content
func GetHTML() (string, error) {
	data, err := content.ReadFile("templates/index.html")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetFile returns the contents of a file from the embedded filesystem
func GetFile(path string) ([]byte, error) {
	return content.ReadFile("templates/" + path)
}

// GetFileSystem returns the embedded file system for direct access
func GetFileSystem() fs.FS {
	return content
}
