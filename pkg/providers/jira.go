package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/ilaif/gh-prx/pkg/config"
	"github.com/ilaif/gh-prx/pkg/models"
)

var (
	JiraIssueTypeToType = map[string]string{
		"bug":   "fix",
		"story": "feat",
		"task":  "chore",
	}
)

type JiraIssueProvider struct {
	Config         config.JiraConfig
	CheckoutNewCfg config.CheckoutNewJiraConfig
}

func (p *JiraIssueProvider) Get(ctx context.Context, id string) (*models.Issue, error) {
	path := fmt.Sprintf("rest/api/3/issue/%s", id)
	issue := &JiraIssue{}
	if err := p.jiraGetRequest(ctx, path, issue); err != nil {
		return nil, err
	}

	return issue.ToIssue(), nil
}

func (p *JiraIssueProvider) List(ctx context.Context) ([]*models.Issue, error) {
	path := fmt.Sprintf("rest/api/3/search?jql=%s", p.CheckoutNewCfg.IssueJQL)
	issues := &JiraIssues{}
	if err := p.jiraGetRequest(ctx, path, issues); err != nil {
		return nil, err
	}

	result := make([]*models.Issue, len(issues.Issues))
	for i, issue := range issues.Issues {
		result[i] = issue.ToIssue()
	}

	return result, nil
}

type JiraIssues struct {
	Issues []JiraIssue `json:"issues"`
}

type JiraIssue struct {
	Key    string     `json:"key"`
	Fields JiraFields `json:"fields"`
}

type JiraFields struct {
	Summary   string        `json:"summary"`
	IssueType JiraIssueType `json:"issuetype"`
}

type JiraIssueType struct {
	Name string `json:"name"`
}

func (i *JiraIssue) ToIssue() *models.Issue {
	issueType := ""
	if it, ok := JiraIssueTypeToType[strings.ToLower(i.Fields.IssueType.Name)]; ok {
		issueType = it
	}

	return &models.Issue{
		Key:   i.Key,
		Title: i.Fields.Summary,
		Type:  issueType,
	}
}

func (p *JiraIssueProvider) jiraGetRequest(ctx context.Context, path string, response any) error {
	url := fmt.Sprintf("%s/%s", p.Config.Endpoint, path)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return errors.Wrapf(err, "Failed to create request for '%s'", url)
	}
	req.SetBasicAuth(p.Config.User, p.Config.Token)

	client := &http.Client{Timeout: time.Second * 10}
	res, err := client.Do(req)
	if err != nil {
		return errors.Wrapf(err, "Failed to request for '%s'", url)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusNotFound {
			return errors.Errorf("Request '%s' not found", path)
		}

		return errors.Errorf("Request '%s' failed: %s", path, res.Status)
	}

	if err := json.NewDecoder(res.Body).Decode(response); err != nil {
		return errors.Wrap(err, "Failed to parse response")
	}

	return nil
}
