package utils

import (
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

func Exec(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return string(bytes), errors.Wrapf(err, "Failed to run 'git %s':\n%s", strings.Join(args, " "), string(bytes))
	}

	return string(bytes), err
}
