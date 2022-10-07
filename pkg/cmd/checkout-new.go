package cmd

import (
	"context"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/caarlos0/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/ilaif/gh-prx/pkg/branch"
	"github.com/ilaif/gh-prx/pkg/config"
	"github.com/ilaif/gh-prx/pkg/providers"
	"github.com/ilaif/gh-prx/pkg/utils"
)

func NewCheckoutNewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checkout-new <issue-id>",
		Short: "Create a new branch based on an issue and checkout to it.",
		Args:  cobra.ExactArgs(1),
		Long: heredoc.Docf(`
			Create a new branch based on an issue and checkout to it.

			If the issue type ({{.Type}}) can't be resolved from the labels automatically,
			the user will be prompted to choose a type.
		`, "`"),
		Example: heredoc.Doc(`
			$ gh prx checkout-new 1234 # Where 1234 is the issue number
		`),
		Aliases: []string{"switch-create", "sc", "cob"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			return checkoutNew(ctx, args[0])
		},
	}

	return cmd
}

func checkoutNew(ctx context.Context, id string) error {
	log.Debug("Loading config")
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	setupCfg, err := config.LoadSetupConfig()
	if err != nil {
		return err
	}

	provider, err := providers.NewIssueProvider(cfg.Issue.Provider, setupCfg)
	if err != nil {
		return err
	}

	s := utils.StartSpinner("Fetching issue from provider...", "Fetched issue from provider")
	issue, err := provider.Get(ctx, id)
	s.Stop()
	if err != nil {
		return errors.Wrap(err, "failed to get issue")
	}

	branchName, err := branch.TemplateBranchName(cfg, issue)
	if err != nil {
		return err
	}

	log.Debugf("Creating branch '%s' and checking out to it", branchName)
	out, err := utils.Exec("git", "checkout", "-b", branchName)
	if err != nil {
		return errors.Wrap(err, "Failed to create branch")
	}

	log.Info(strings.Trim(out, "\n"))

	return nil
}
