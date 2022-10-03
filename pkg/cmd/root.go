package cmd

import (
	"context"
	"os"

	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
)

func Execute(version string) {
	rootCmd := NewRootCmd(version)
	ctx := context.Background()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.WithError(err).Error("Command failed")
		os.Exit(1)
	}
}

func NewRootCmd(version string) *cobra.Command {
	var (
		debug bool
	)

	var rootCmd = &cobra.Command{
		Use:           "prx",
		Short:         "Extended Git & GitHub CLI flows",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.SetLevel(log.InfoLevel)
			log.DecreasePadding() // remove the default padding

			if debug {
				log.Info("Debug logs enabled")
				log.SetLevel(log.DebugLevel)
			}
		},
	}

	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "verbose logging")

	rootCmd.AddCommand(
		newCreateCmd(),
		newCheckoutNewCmd(),
	)

	return rootCmd
}
