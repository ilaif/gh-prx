package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/caarlos0/log"
	"github.com/cli/go-gh"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/ilaif/gh-prx/pkg/ai"
	"github.com/ilaif/gh-prx/pkg/branch"
	"github.com/ilaif/gh-prx/pkg/config"
	"github.com/ilaif/gh-prx/pkg/pr"
	"github.com/ilaif/gh-prx/pkg/utils"
)

type CreateOpts struct {
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

	NoAISummary bool
}

func NewCreateCmd() *cobra.Command {
	opts := &CreateOpts{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a pull request on GitHub, extended.",
		Long: heredoc.Docf(`
			Create a pull request on GitHub, extended.

			When the current branch isn't fully pushed to a git remote, the command will push it to origin.
			This behavior can be disabled by setting %[1]spr.push_to_remote: false%[1]s in the config file.

			A pull request title will be generated based on the current branch name and the config file (if present).

			A pull request description (body) template can be defined in %[1]s.github/pull_request_template.md%[1]s.

			All of %[1]sgh pr create%[1]s flags are supported.
		`, "`"),
		Example: heredoc.Doc(`
			$ gh prx create # Good defaults
			$ gh prx create --web # Open the pull request in the browser before creating it
			$ gh prx create --confirm # skip confirmation prompt for PR checklist questions
		`),
		Aliases: []string{"new"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return create(cmd.Context(), opts)
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
	fl.Bool("no-ai-summary", false, "Disable AI-powered summary")

	return cmd
}

func create(ctx context.Context, opts *CreateOpts) error {
	setupCfg, err := config.LoadSetupConfig()
	if err != nil {
		return err
	}

	cfg, err := config.LoadRepositoryConfig(setupCfg.RepositoryConfig)
	if err != nil {
		return err
	}

	log.Debug("Fetching current branch name")
	gitDiffOutput, err := utils.Exec("git", "branch", "--show-current")
	if err != nil {
		return err
	}
	branchName := strings.Trim(gitDiffOutput, "\n")

	b, err := branch.ParseBranch(branchName, cfg.Branch)
	if err != nil {
		return err
	}

	if *cfg.PR.PushToRemote {
		s := utils.StartSpinner("Pushing current branch to remote...", "Pushed branch to remote")
		gitDiffOutput, err = utils.Exec("git", "push", "--set-upstream", "origin", b.Original)
		s.Stop()
		if err != nil {
			return err
		}
		log.Info(strings.Trim(gitDiffOutput, "\n"))
	}

	base := opts.BaseBranch
	if base == "" {
		s := utils.StartSpinner("Fetching repository default branch...", "Fetched repository default branch")
		stdOut, _, err := gh.Exec("repo", "view", "--json", "defaultBranchRef", "--jq", ".defaultBranchRef.name")
		s.Stop()
		if err != nil {
			return errors.Wrap(err, "Failed to fetch default branch")
		}
		base = strings.Trim(stdOut.String(), "\n")
	}

	if cfg.IgnorePullRequestTemplate != nil && *cfg.IgnorePullRequestTemplate {
		prTemplateBytes, err := utils.ReadFile(".github/pull_request_template.md")
		if err == nil {
			cfg.PR.Body = string(prTemplateBytes)
		}
	}

	commitsFetcher := func() ([]string, error) {
		out, err := utils.Exec("git", "log", "--pretty=format:%s", "--no-merges", b.Original, "^"+base)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to fetch branch commits")
		}

		return strings.Split(out, "\n"), nil
	}

	aiSummarizer := func() (string, error) {
		if opts.NoAISummary || !ai.IsAISummarizerAvailable() {
			return "", nil
		}

		return createAISummary(ctx, base, commitsFetcher)
	}

	pr, err := pr.TemplatePR(b, cfg.PR, opts.Confirm, cfg.Branch.TokenSeparators, commitsFetcher, aiSummarizer)
	if err != nil {
		return err
	}

	log.Debug(fmt.Sprintf("Pull request body:\n\n%s", pr.Body))

	if len(pr.Labels) > 0 {
		if err := createLabels(pr.Labels); err != nil {
			return err
		}
	}

	s := utils.StartSpinner("Creating pull request...", "Created pull request")
	args := []string{"pr", "create", "--title", pr.Title, "--body", pr.Body, "--base", base}
	args = append(args, generatePrCreateArgsFromOpts(opts, pr.Labels)...)
	stdOut, _, err := gh.Exec(args...)
	s.Stop()
	if err != nil {
		return errors.Wrap(err, "Failed to create pull request")
	}
	log.Info(strings.Trim(stdOut.String(), "\n"))

	return nil
}

func createAISummary(ctx context.Context,
	base string,
	commitsFetcher func() ([]string, error),
) (string, error) {
	aiSummary := ""

	s := utils.StartSpinner("Creating an AI-powered summary", "Finished summarizing")
	defer s.Stop()

	gitDiffCmd := heredoc.Docf(`
			git diff main --stat |
			grep '|' |
			awk '{ if ($3 > 10) print $1 }' |
			xargs git diff ^%s --ignore-all-space --ignore-blank-lines --ignore-space-change --unified=0 --word-diff --
		`, base)
	gitDiffOutput, err := utils.Exec("sh", "-c", gitDiffCmd)
	if err != nil {
		return "", errors.Wrap(err, "Failed to fetch branch commits")
	}

	aiSummary, err = ai.SummarizeGitDiffOutput(ctx, gitDiffOutput)
	if err != nil {
		// Fallback to file and commit diff

		commits, err := commitsFetcher()
		if err != nil {
			return "", err
		}

		aiSummary, err = ai.SummarizeGitDiffOutput(ctx, strings.Join(commits, "\n"))
		if err != nil {
			return "", err
		}
	}

	return aiSummary, nil
}

func createLabels(labels []string) error {
	s := utils.StartSpinner("Creating labels (if not exist)...", "Created labels")
	defer s.Stop()

	g := errgroup.Group{}
	for _, label := range labels {
		label := label

		g.Go(func() error {
			_, stdErr, err := gh.Exec("label", "create", label)
			if err != nil {
				if !strings.Contains(stdErr.String(), "already exists") {
					return errors.Wrapf(err, "Failed to create label '%s'", label)
				}

				return nil
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "Failed to create labels")
	}

	return nil
}

func generatePrCreateArgsFromOpts(opts *CreateOpts, labels []string) []string {
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
