package providers

import (
	"context"

	"github.com/ilaif/gh-prx/pkg/config"
	"github.com/ilaif/gh-prx/pkg/models"
)

type IssueProvider interface {
	Get(ctx context.Context, issue string) (*models.Issue, error)
}

func NewIssueProvider(provider string, cfg *config.SetupConfig) (IssueProvider, error) {
	switch provider {
	case "github":
		return &GitHubIssueProvider{}, nil
	case "jira":
		if err := cfg.JiraConfig.Validate(); err != nil {
			return nil, err
		}

		return &JiraIssueProvider{Config: cfg.JiraConfig}, nil
	default:
		return nil, config.ErrInvalidProvider
	}
}
