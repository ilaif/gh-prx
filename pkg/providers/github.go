package providers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cli/go-gh"
	"github.com/pkg/errors"

	"github.com/ilaif/gh-prx/pkg/models"
)

var (
	LabelToType = map[string]string{
		"bug":           "fix",
		"enhancement":   "feat",
		"documentation": "docs",

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
