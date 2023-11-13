package providers

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	graphql "github.com/hasura/go-graphql-client"
	"github.com/pkg/errors"

	"github.com/ilaif/gh-prx/pkg/config"
	"github.com/ilaif/gh-prx/pkg/models"
)

const (
	LinearGraphQLEndpoint = "https://api.linear.app/graphql"
)

type LinearIssueProvider struct {
	Config config.LinearConfig
	// CheckoutNewCfg config.CheckoutNewJiraConfig
}

func (p *LinearIssueProvider) Name() string {
	return "Linear"
}

func (p *LinearIssueProvider) Get(ctx context.Context, id string) (*models.Issue, error) {
	query := &LinearIssueQuery{}
	vars := map[string]interface{}{
		"id": graphql.String(id),
	}
	if err := p.query(ctx, query, vars); err != nil {
		return nil, err
	}

	return query.Issue.ToIssue(), nil
}

func (p *LinearIssueProvider) List(ctx context.Context) ([]*models.Issue, error) {
	query := &LinearIssues{}
	vars := map[string]interface{}{}
	if err := p.query(ctx, query, vars); err != nil {
		return nil, err
	}

	result := make([]*models.Issue, len(query.Viewer.AssignedIssues.Nodes))
	for i, issue := range query.Viewer.AssignedIssues.Nodes {
		result[i] = issue.ToIssue()
	}

	return result, nil
}

type LinearIssues struct {
	Viewer struct {
		AssignedIssues struct {
			Nodes []LinearIssue
		} `graphql:"assignedIssues(orderBy: updatedAt, filter: { state: { type: { neq: \"completed\" } } })"`
	}
}

type LinearIssueQuery struct {
	Issue LinearIssue `graphql:"issue(id: $id)"`
}

type LinearIssue struct {
	Identifier string
	Title      string
	State      struct {
		Name string
		Type string
	}
	Labels struct {
		Nodes []struct {
			Name string
		}
	}
	BranchName string
}

func (i *LinearIssue) ToIssue() *models.Issue {
	issueType := ""
	for _, label := range i.Labels.Nodes {
		if it, ok := LabelToType[strings.ToLower(label.Name)]; ok {
			issueType = it

			break
		}
	}

	return &models.Issue{
		Key:                 i.Identifier,
		Title:               i.Title,
		Type:                issueType,
		SuggestedBranchName: i.BranchName,
	}
}

func (p *LinearIssueProvider) query(ctx context.Context, query any, vars map[string]any) error {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	client := graphql.NewClient("https://api.linear.app/graphql", httpClient)
	client = client.WithRequestModifier(func(req *http.Request) {
		req.Header.Set("Authorization", os.Getenv("LINEAR_API_KEY"))
	})

	err := client.Query(ctx, query, vars)
	if err != nil {
		return errors.Wrap(err, "Failed to query Linear")
	}

	return nil
}
