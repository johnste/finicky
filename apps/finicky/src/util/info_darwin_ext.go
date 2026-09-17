//go:build darwin

package util

import "path/filepath"

func LogDir() (string, error) {
	home, err := UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Logs", "Finicky"), nil
}
