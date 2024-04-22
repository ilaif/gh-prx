package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

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

	WebMode          bool
	NoMaintainerEdit bool
	RecoverFile      string

	IsDraft    bool
	BaseBranch string
	HeadBranch string

	Reviewers []string
	Assignees []string
	Labels    []string
	Projects  []string
	Milestone string

	NoAISummary bool

	DryRun bool
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
		RunE: func(cmd *cobra.Command, _ []string) error {
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
	fl.BoolVar(&opts.NoMaintainerEdit, "no-maintainer-edit", false, "Disable maintainer's ability to modify pull request")
	fl.StringVar(&opts.RecoverFile, "recover", "", "Recover input from a failed run of create")
	fl.BoolVar(&opts.NoAISummary, "no-ai-summary", false, "Disable AI-powered summary")
	fl.BoolVar(&opts.DryRun, "dry-run", false, "Print the pull request body and title without creating the pull request")

	return cmd
}

func create(ctx context.Context, opts *CreateOpts) error { // nolint:cyclop
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

	baseBranch := opts.BaseBranch
	if baseBranch == "" {
		s := utils.StartSpinner("Fetching repository default branch...", "Fetched repository default branch")
		stdOut, _, err := gh.Exec("repo", "view", "--json", "defaultBranchRef", "--jq", ".defaultBranchRef.name")
		s.Stop()
		if err != nil {
			return errors.Wrap(err, "Failed to fetch default branch")
		}
		baseBranch = strings.Trim(stdOut.String(), "\n")
	}

	if cfg.PullRequestTemplatePath != "" {
		prTemplatePath, err := utils.FindRelativePathInRepo(cfg.PullRequestTemplatePath)
		if err != nil {
			return errors.Wrap(err, "Failed to find pull request template path")
		}
		prTemplateBytes, err := utils.ReadFile(prTemplatePath)
		if err != nil {
			return errors.Wrap(err, "Failed to read pull request template")
		}
		cfg.PR.Body = string(prTemplateBytes)
	}

	out, err := utils.Exec("git", "log", "--pretty=format:%s", "--no-merges", b.Original, "^"+baseBranch)
	if err != nil {
		return errors.Wrap(err, "Failed to fetch branch commits")
	}
	commits := strings.Split(out, "\n")
	log.Debug(fmt.Sprintf("Commits:\n%s", strings.Join(commits, "\n")))

	aiSummarizer := func() (string, error) {
		if opts.NoAISummary || !ai.IsAISummarizerAvailable() {
			log.Debug("AI-powered summary is disabled")

			return "", nil
		}

		aiSummary, err := createAISummary(ctx, baseBranch, cfg.PR.Body, commits)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				log.Warn("AI-powered summary timed out, skipping")

				return "", nil
			}

			return "", err
		}

		return aiSummary, nil
	}

	pr, err := pr.TemplatePR(b, cfg.PR, opts.Confirm, cfg.Branch.TokenSeparators, commits, aiSummarizer)
	if err != nil {
		return err
	}

	log.Debug(fmt.Sprintf("Pull request title: %s", pr.Title))
	log.Debug(fmt.Sprintf("Pull request body:\n\n%s", pr.Body))
	log.Debug(fmt.Sprintf("Pull request labels: %v", pr.Labels))

	if opts.DryRun {
		log.Info("Dry run enabled, skipping pull request creation")

		return nil
	}

	if len(pr.Labels) > 0 {
		if err := createLabels(pr.Labels); err != nil {
			return err
		}
	}

	s := utils.StartSpinner("Creating pull request...", "Created pull request")
	args := []string{"pr", "create", "--title", pr.Title, "--body", pr.Body, "--base", baseBranch}
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
	baseBranch string,
	prBody string,
	commits []string,
) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	s := utils.StartSpinner("Creating an AI-powered summary", "Finished summarizing")
	defer s.Stop()

	gitDiffCmd := heredoc.Docf(`
			git diff %[1]s --stat=10000 |
			grep '|' |
			awk '{ if ($3 > 10) print $1 }' |
			xargs git diff ^%[1]s --ignore-all-space --ignore-blank-lines --ignore-space-change --unified=0 --word-diff --
		`, baseBranch)
	gitDiffOutput, err := utils.Exec("sh", "-c", gitDiffCmd)
	if err != nil {
		return "", errors.Wrap(err, "Failed to fetch branch commits")
	}

	aiSummary, err := ai.SummarizeGitDiffOutput(ctx, gitDiffOutput, prBody)
	if err != nil {
		log.Debug("Failed to summarize git diff output, falling back to file and commit diff")
		aiSummary, err = ai.SummarizeGitDiffOutput(ctx, strings.Join(commits, "\n"), prBody)
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
	if opts.NoMaintainerEdit {
		args = append(args, "--no-maintainer-edit")
	}
	if opts.RecoverFile != "" {
		args = append(args, "--recover", opts.RecoverFile)
	}
	if opts.HeadBranch != "" {
		args = append(args, "--head", opts.HeadBranch)
	}

	return args
}
