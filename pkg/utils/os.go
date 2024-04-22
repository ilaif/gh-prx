package utils

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/caarlos0/log"
	"github.com/pkg/errors"
)

func Exec(name string, args ...string) (string, error) {
	log.Debug(fmt.Sprintf("Running '%s %s'", name, strings.Join(args, " ")))
	cmd := exec.Command(name, args...)
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return string(bytes), errors.Wrapf(err, "Failed to run 'git %s':\n%s", strings.Join(args, " "), string(bytes))
	}

	return string(bytes), err
}
