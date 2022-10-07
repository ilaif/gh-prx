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
	Config config.JiraConfig
}

func (p *JiraIssueProvider) Get(ctx context.Context, id string) (*models.Issue, error) {
	url := fmt.Sprintf("%s/rest/api/3/issue/%s", p.Config.Endpoint, id)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create request for '%s'", url)
	}

	req.SetBasicAuth(p.Config.User, p.Config.Token)

	client := &http.Client{Timeout: time.Second * 5}
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get Jira issue '%s'", id)
	}
	defer res.Body.Close()

	issue := &JiraIssue{}
	if err := json.NewDecoder(res.Body).Decode(issue); err != nil {
		return nil, errors.Wrap(err, "Failed to parse Jira issue")
	}

	return issue.ToIssue(), nil
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
