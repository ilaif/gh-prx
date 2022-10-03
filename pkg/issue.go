package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/cli/go-gh"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

var (
	IssueProviders           = []string{"github"}
	ErrInvalidProvider       = errors.New("Invalid provider")
	InvalidTitleCharsMatcher = regexp.MustCompile(`[^.a-zA-Z0-9]`)
)

type IssueConfig struct {
	Provider string `yaml:"provider"`
}

func (c *IssueConfig) SetDefaults() {
	if c.Provider == "" {
		c.Provider = "github"
	}
}

func (c *IssueConfig) Validate() error {
	if !lo.Contains(IssueProviders, c.Provider) {
		return errors.Wrapf(ErrInvalidProvider, "Provider must be one of %s", strings.Join(IssueProviders, ", "))
	}

	return nil
}

type Issue struct {
	Code   string
	Title  string
	Labels []string
}

type IssueProvider interface {
	Get(ctx context.Context, issue string) (*Issue, error)
}

type GitHubIssue struct {
	Number int      `json:"number"`
	Title  string   `json:"title"`
	Labels []string `json:"labels"`
}

func (i *GitHubIssue) ToIssue() *Issue {
	return &Issue{
		Code:   fmt.Sprintf("%d", i.Number),
		Title:  i.Title,
		Labels: i.Labels,
	}
}

type GitHubIssueProvider struct {
}

func (p *GitHubIssueProvider) Get(ctx context.Context, id string) (*Issue, error) {
	stdOut, _, err := gh.Exec("issue", "view", id, "--json", "number,title")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get GitHub issue")
	}

	issue := &GitHubIssue{}
	if err := json.Unmarshal(stdOut.Bytes(), &issue); err != nil {
		return nil, errors.Wrap(err, "Failed to parse GitHub issue")
	}

	return issue.ToIssue(), nil
}

func NewIssueProvider(provider string) (IssueProvider, error) {
	switch provider {
	case "github":
		return &GitHubIssueProvider{}, nil
	default:
		return nil, ErrInvalidProvider
	}
}

func normalizeIssueTitle(title string) string {
	title = InvalidTitleCharsMatcher.ReplaceAllString(title, "-")
	title = strings.ToLower(title)

	return title
}
