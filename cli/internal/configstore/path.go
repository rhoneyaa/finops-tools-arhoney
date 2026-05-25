package configstore

import (
	"os"
	"path/filepath"
	"runtime"
)

// DefaultPath returns the OS-appropriate finops config file path.
func DefaultPath() (string, error) {
	if runtime.GOOS == "windows" {
		dir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(dir, "finops", "config.yaml"), nil
	}

	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "finops", "config.yaml"), nil
}

// ResolvePath returns flagPath when set, otherwise DefaultPath.
func ResolvePath(flagPath string) (string, error) {
	if flagPath != "" {
		return flagPath, nil
	}
	return DefaultPath()
}
