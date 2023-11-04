package providers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cli/go-gh"
	"github.com/pkg/errors"

	"github.com/ilaif/gh-prx/pkg/config"
	"github.com/ilaif/gh-prx/pkg/models"
)

var (
	LabelToType = map[string]string{
		"bug":           "fix",
		"enhancement":   "feat",
		"documentation": "docs",
		"feature":       "feat",
		"feat":          "feat",
		"fix":           "fix",

		"chore":    "chore",
		"refactor": "refactor",
		"test":     "test",
		"ci":       "ci",
		"perf":     "perf",
		"build":    "build",
		"revert":   "revert",
		"style":    "style",
	}
)

type GitHubIssueProvider struct {
	CheckoutNewConfig config.CheckoutNewGitHubConfig
}

func (p *GitHubIssueProvider) Name() string {
	return "GitHub"
}

func (p *GitHubIssueProvider) Get(_ context.Context, id string) (*models.Issue, error) {
	stdOut, _, err := gh.Exec("issue", "view", id, "--json", "number,title,labels")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get GitHub issue")
	}

	issue := &GitHubIssue{}
	if err := json.Unmarshal(stdOut.Bytes(), &issue); err != nil {
		return nil, errors.Wrap(err, "Failed to parse GitHub issue")
	}

	return issue.ToIssue(), nil
}

func (p *GitHubIssueProvider) List(_ context.Context) ([]*models.Issue, error) {
	args := []string{"issue", "list", "--json", "number,title,labels"}
	args = append(args, p.CheckoutNewConfig.IssueListFlags...)
	stdOut, _, err := gh.Exec(args...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to list GitHub issues")
	}

	issues := []*GitHubIssue{}
	if err := json.Unmarshal(stdOut.Bytes(), &issues); err != nil {
		return nil, errors.Wrap(err, "Failed to parse GitHub issues")
	}

	result := []*models.Issue{}
	for _, issue := range issues {
		result = append(result, issue.ToIssue())
	}

	return result, nil
}

type GitHubIssue struct {
	Number int           `json:"number"`
	Title  string        `json:"title"`
	Labels []GitHubLabel `json:"labels"`
}

type GitHubLabel struct {
	Name string `json:"name"`
}

func (i *GitHubIssue) ToIssue() *models.Issue {
	issueType := ""
	for _, label := range i.Labels {
		if it, ok := LabelToType[label.Name]; ok {
			issueType = it

			break
		}
	}

	return &models.Issue{
		Key:   fmt.Sprintf("%d", i.Number),
		Title: i.Title,
		Type:  issueType,
	}
}
