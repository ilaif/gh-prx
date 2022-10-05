package utils

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

func StartSpinner(loadingMsg string, finalMsg string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[57], 100*time.Millisecond)
	_ = s.Color("red")
	s.Suffix = fmt.Sprintf(" %s", loadingMsg)
	s.FinalMSG = fmt.Sprintf("• %s\n", finalMsg)

	s.Start()

	return s
}
