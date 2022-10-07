package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/briandowns/spinner"
	"github.com/pkg/errors"
)

func StartSpinner(loadingMsg string, finalMsg string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[57], 100*time.Millisecond)
	_ = s.Color("red")
	s.Suffix = fmt.Sprintf(" %s", loadingMsg)
	s.FinalMSG = fmt.Sprintf("• %s\n", finalMsg)

	s.Start()

	return s
}

// EditString opens the default editor with input string.
// Returns the edited string.
func EditString(input string) (string, error) {
	vi := "vi"
	tmpFile, err := ioutil.TempFile(os.TempDir(), "edit-branch-name")
	if err != nil {
		return "", errors.Wrapf(err, "Failed to create temp file")
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	viPath, err := exec.LookPath(vi)
	if err != nil {
		return "", errors.Wrapf(err, "Failed to find '%s' to edit branch name", vi)
	}

	if _, err := tmpFile.WriteString(input); err != nil {
		return "", errors.Wrapf(err, "Failed to write to temp file")
	}

	cmd := exec.Command(viPath, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return "", errors.Wrapf(err, "Failed to start '%s'", vi)
	}
	if err := cmd.Wait(); err != nil {
		return "", errors.Wrapf(err, "Failed to wait for '%s'", vi)
	}

	bytes, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return "", errors.Wrapf(err, "Failed to read temp file")
	}

	return string(bytes), nil
}
