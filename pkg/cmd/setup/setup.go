package setup

import (
	"github.com/spf13/cobra"
)

func NewSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup commands.",
	}

	cmd.AddCommand(NewProviderCmd())

	return cmd
}
