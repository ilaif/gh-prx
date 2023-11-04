package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/caarlos0/log"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/ilaif/gh-prx/pkg/branch"
	"github.com/ilaif/gh-prx/pkg/config"
	"github.com/ilaif/gh-prx/pkg/models"
	"github.com/ilaif/gh-prx/pkg/providers"
	"github.com/ilaif/gh-prx/pkg/utils"
)

func NewCheckoutNewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checkout-new [issue-id]",
		Short: "Create a new branch based on an issue and checkout to it.",
		Args:  cobra.MaximumNArgs(1),
		Long: heredoc.Docf(`
			Create a new branch based on an issue and checkout to it.

			If the issue type ({{.Type}}) can't be resolved from the labels automatically,
			the user will be prompted to choose a type.
		`, "`"),
		Example: heredoc.Doc(`
			// Create a new branch based on a list of available issues and checkout to it:
			$ gh prx checkout-new

			// Create a new branch based on issue 1234 and checkout to it:
			$ gh prx checkout-new 1234
		`),
		Aliases: []string{"switch-create", "sc", "cob"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			issueID := ""
			if len(args) > 0 {
				issueID = args[0]
			}

			return checkoutNew(ctx, issueID)
		},
	}

	return cmd
}

func checkoutNew(ctx context.Context, id string) error {
	setupCfg, err := config.LoadSetupConfig()
	if err != nil {
		return err
	}

	cfg, err := config.LoadRepositoryConfig(setupCfg.RepositoryConfig)
	if err != nil {
		return err
	}

	provider, err := providers.NewIssueProvider(cfg, setupCfg)
	if err != nil {
		return err
	}

	if id == "" {
		id, err = chooseIssue(ctx, provider)
		if err != nil {
			return errors.Wrap(err, "Failed to choose issue")
		}
	}

	s := utils.StartSpinner("Fetching issue from provider...", "Fetched issue from provider")
	issue, err := provider.Get(ctx, id)
	s.Stop()
	if err != nil {
		return errors.Wrap(err, "Failed to get issue")
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

func chooseIssue(ctx context.Context, provider providers.IssueProvider) (string, error) {
	s := utils.StartSpinner(
		fmt.Sprintf("Fetching issues from %s...", provider.Name()),
		fmt.Sprintf("Fetched issues from %s", provider.Name()),
	)
	issues, err := provider.List(ctx)
	s.Stop()
	if err != nil {
		return "", errors.Wrap(err, "Failed to list issues")
	}

	if len(issues) == 0 {
		return "", errors.New("No issues found")
	}

	var i int
	if err := survey.Ask([]*survey.Question{
		{
			Name: "issue",
			Prompt: &survey.Select{
				Message: "Select an issue:",
				Options: lo.Map(issues, func(i *models.Issue, _ int) string {
					issueType := ""
					if i.Type != "" {
						issueType = fmt.Sprintf("(%s) ", i.Type)
					}

					return fmt.Sprintf("%s%s - %s", issueType, i.Key, i.Title)
				}),
			},
		},
	}, &i); err != nil {
		return "", errors.Wrap(err, "Failed to prompt for issue")
	}

	return issues[i].Key, nil
}
