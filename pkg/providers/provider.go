package providers

import (
	"context"

	"github.com/ilaif/gh-prx/pkg/config"
	"github.com/ilaif/gh-prx/pkg/models"
)

type IssueProvider interface {
	Get(ctx context.Context, issue string) (*models.Issue, error)
	List(ctx context.Context) ([]*models.Issue, error)
}

func NewIssueProvider(cfg *config.RepositoryConfig, setupCfg *config.SetupConfig) (IssueProvider, error) {
	switch cfg.Issue.Provider {
	case "github":
		return &GitHubIssueProvider{
			CheckoutNewConfig: cfg.CheckoutNew.GitHub,
		}, nil
	case "jira":
		if err := setupCfg.JiraConfig.Validate(); err != nil {
			return nil, err
		}

		return &JiraIssueProvider{
			Config:         setupCfg.JiraConfig,
			CheckoutNewCfg: cfg.CheckoutNew.Jira,
		}, nil
	default:
		return nil, config.ErrInvalidProvider
	}
}
