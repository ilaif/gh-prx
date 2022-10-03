package utils

import (
	"io/fs"
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

func WriteStringToFile(filename string, text string) error {
	if err := os.WriteFile(filename, []byte(text), fs.FileMode(os.O_WRONLY)); err != nil {
		return errors.Wrapf(err, "Failed to write to file '%s'", filename)
	}

	return nil
}

func ReadYaml(filename string, config interface{}) error {
	buf, err := ReadFile(filename)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(buf, config)
	if err != nil {
		return errors.Wrapf(err, "Failed to parse config from '%s'", filename)
	}

	return nil
}
