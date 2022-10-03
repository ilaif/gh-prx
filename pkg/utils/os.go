package utils

import (
	"context"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

func Exec(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return string(bytes), errors.Wrapf(err, "Failed to run 'git %s':\n%s", strings.Join(args, " "), string(bytes))
	}

	return string(bytes), err
}
