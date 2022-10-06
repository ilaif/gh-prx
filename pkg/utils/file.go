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
