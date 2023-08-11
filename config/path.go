package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const pathFile = "path"

func LoadPath() (string, error) {
	cd, err := configDir()
	if err != nil {
		return "", err
	}

	d, err := os.ReadFile(filepath.Join(cd, pathFile))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "/", nil
		}
		return "", err
	}

	return strings.TrimSpace(string(d)), err
}

func SavePath(s string) error {
	cd, err := configDir()
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(cd, pathFile), []byte(s), 0644)
}
