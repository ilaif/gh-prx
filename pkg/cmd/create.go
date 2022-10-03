package cmd

import (
	"context"
	"strings"

	"github.com/caarlos0/log"
	"github.com/cli/go-gh"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/ilaif/gh-prx/pkg"
	"github.com/ilaif/gh-prx/pkg/utils"
)

type CreateOptions struct {
	Confirm bool

	WebMode     bool
	RecoverFile string

	IsDraft    bool
	BaseBranch string
	HeadBranch string

	Reviewers []string
	Assignees []string
	Labels    []string
	Projects  []string
	Milestone string
}

func newCreateCmd() *cobra.Command {
	opts := &CreateOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a pull request on GitHub, extended.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			return create(ctx, opts)
		},
	}

	fl := cmd.Flags()
	fl.BoolVarP(&opts.Confirm, "confirm", "y", false, "Don't ask for user input")
	fl.BoolVarP(&opts.IsDraft, "draft", "d", false, "Mark pull request as a draft")
	fl.StringVarP(&opts.BaseBranch, "base", "B", "", "The `branch` into which you want your code merged")
	fl.StringVarP(
		&opts.HeadBranch,
		"head",
		"H",
		"",
		"The `branch` that contains commits for your pull request (default: current branch)",
	)
	fl.BoolVarP(&opts.WebMode, "web", "w", false, "Open the web browser to create a pull request")
	fl.StringSliceVarP(&opts.Reviewers, "reviewer", "r", nil, "Request reviews from people or teams by their `handle`")
	fl.StringSliceVarP(
		&opts.Assignees,
		"assignee",
		"a",
		nil,
		"Assign people by their `login`. Use \"@me\" to self-assign.",
	)
	fl.StringSliceVarP(&opts.Labels, "label", "l", nil, "Add labels by `name`")
	fl.StringSliceVarP(&opts.Projects, "project", "p", nil, "Add the pull request to projects by `name`")
	fl.StringVarP(&opts.Milestone, "milestone", "m", "", "Add the pull request to a milestone by `name`")
	fl.Bool("no-maintainer-edit", false, "Disable maintainer's ability to modify pull request")
	fl.StringVar(&opts.RecoverFile, "recover", "", "Recover input from a failed run of create")

	return cmd
}

func create(ctx context.Context, opts *CreateOptions) error {
	log.Debug("Loading config")
	cfg, err := pkg.LoadConfig()
	if err != nil {
		return err
	}

	log.Debug("Fetching current branch name")
	out, err := utils.Exec(ctx, "git", "branch", "--show-current")
	if err != nil {
		return err
	}
	branchName := strings.Trim(out, "\n")

	b, err := pkg.ParseBranch(branchName, cfg.Branch)
	if err != nil {
		return err
	}

	if *cfg.PushToRemote {
		log.Info("Pushing branch to remote")
		out, err = utils.Exec(ctx, "git", "push", "--set-upstream", "origin", b.Original)
		if err != nil {
			return err
		}
		log.IncreasePadding()
		log.Info(strings.Trim(out, "\n"))
		log.DecreasePadding()
	}

	base := opts.BaseBranch
	if base == "" {
		stdOut, _, err := gh.Exec("repo", "view", "--json", "defaultBranchRef", "--jq", ".defaultBranchRef.name")
		if err != nil {
			return errors.Wrap(err, "Failed to fetch default branch")
		}
		base = strings.Trim(stdOut.String(), "\n")
	}

	prTemplateBytes, err := utils.ReadFile(".github/.pull_request_template.md")
	if err == nil {
		cfg.PR.Body = string(prTemplateBytes)
	}

	commitsFetcher := func() ([]string, error) {
		out, err := utils.Exec(ctx, "git", "log", "--pretty=format:%s", "--no-merges", b.Original, "^"+base)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to fetch branch commits")
		}

		return strings.Split(out, "\n"), nil
	}

	pr, err := pkg.TemplatePR(b, cfg.PR, opts.Confirm, cfg.Branch.TokenSeparators, commitsFetcher)
	if err != nil {
		return err
	}

	args := []string{"pr", "create", "--title", pr.Title, "--body", pr.Body, "--base", base}
	args = append(args, generatePrCreateArgsFromOpts(opts, pr.Labels)...)
	stdOut, _, err := gh.Exec(args...)
	if err != nil {
		return errors.Wrap(err, "Failed to create pull request")
	}
	log.Info(strings.Trim(stdOut.String(), "\n"))

	return nil
}

func generatePrCreateArgsFromOpts(opts *CreateOptions, labels []string) []string {
	args := []string{}

	if len(opts.Assignees) > 0 {
		args = append(args, "--assignee", strings.Join(opts.Assignees, ","))
	}
	if len(opts.Labels) > 0 || len(labels) > 0 {
		args = append(args, "--label", strings.Join(append(opts.Labels, labels...), ","))
	}
	if len(opts.Projects) > 0 {
		args = append(args, "--project", strings.Join(opts.Projects, ","))
	}
	if opts.Milestone != "" {
		args = append(args, "--milestone", opts.Milestone)
	}
	if len(opts.Reviewers) > 0 {
		args = append(args, "--reviewer", strings.Join(opts.Reviewers, ","))
	}
	if opts.IsDraft {
		args = append(args, "--draft")
	}
	if opts.WebMode {
		args = append(args, "--web")
	}
	if opts.RecoverFile != "" {
		args = append(args, "--recover", opts.RecoverFile)
	}
	if opts.HeadBranch != "" {
		args = append(args, "--head", opts.HeadBranch)
	}

	return args
}
