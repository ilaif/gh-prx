package cmd

import (
	"context"
	"strings"

	"github.com/caarlos0/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/ilaif/gh-prx/pkg"
	"github.com/ilaif/gh-prx/pkg/utils"
)

type CheckoutNewOpts struct {
}

func newCheckoutNewCmd() *cobra.Command {
	opts := &CreateOptions{}

	cmd := &cobra.Command{
		Use:     "checkout-new <issue-id>",
		Short:   "Create a new branch based on an issue and checkout to it.",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"switch-create", "sc", "cob"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			return checkoutNew(ctx, args[0], opts)
		},
	}

	return cmd
}

func checkoutNew(ctx context.Context, id string, opts *CreateOptions) error {
	log.Debug("Loading config")
	cfg, err := pkg.LoadConfig()
	if err != nil {
		return err
	}

	provider, err := pkg.NewIssueProvider(cfg.Issue.Provider)
	if err != nil {
		return err
	}

	issue, err := provider.Get(ctx, id)
	if err != nil {
		return errors.Wrap(err, "failed to get issue")
	}

	branchName, err := pkg.TemplateBranchName(cfg.Branch, issue)
	if err != nil {
		return err
	}

	out, err := utils.Exec(ctx, "git", "checkout", "-b", branchName)
	if err != nil {
		return errors.Wrap(err, "Failed to create branch")
	}

	log.Info(strings.Trim(out, "\n"))

	return nil
}
