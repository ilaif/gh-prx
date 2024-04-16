package utils

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func ReadFile(filename string) ([]byte, error) {
	filename, err := filepath.Abs(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get absolute path for file '%s'", filename)
	}

	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read file '%s'", filename)
	}

	return buf, nil
}

func WriteFile(filename string, content []byte) error {
	if err := os.WriteFile(filename, content, 0600); err != nil {
		return errors.Wrapf(err, "Failed to write to file '%s'", filename)
	}

	return nil
}

func ReadYaml(filename string, data interface{}) error {
	buf, err := ReadFile(filename)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(buf, data)
	if err != nil {
		return errors.Wrapf(err, "Failed to parse config from '%s'", filename)
	}

	return nil
}

func WriteYaml(filename string, data interface{}) error {
	out, err := yaml.Marshal(data)
	if err != nil {
		return errors.Wrapf(err, "Failed to marshal data to yaml")
	}

	if err := WriteFile(filename, out); err != nil {
		return errors.Wrapf(err, "Failed to write yaml to '%s'", filename)
	}

	return nil
}

// FindRelativePathInRepo traverses up from the current directory recursively until a relative path is found.
// It returns the first relative path found, or os.ErrNotExists if no relative path is found.
func FindRelativePathInRepo(path string) (string, error) {
	// Get the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "Failed to get current working directory")
	}

	// Start traversing up from the current directory
	for {
		// Check if the current directory contains a relative path
		if _, err := os.Stat(filepath.Join(currentDir, path)); err == nil {
			return filepath.Join(currentDir, path), nil
		}

		// Move up to the parent directory
		parentDir := filepath.Dir(currentDir)

		// Check if we have reached the git root directory
		if _, err := os.Stat(filepath.Join(currentDir, ".git")); err == nil {
			break // Break the loop if we have reached the git root directory
		}

		// Update the current directory to the parent directory
		currentDir = parentDir
	}

	return "", os.ErrNotExist // Return an error if no relative path is found
}
