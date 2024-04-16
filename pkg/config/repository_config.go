package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"dario.cat/mergo"
	"github.com/caarlos0/log"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/ilaif/gh-prx/pkg/utils"
)

const (
	DefaultConfigFilepath = ".github/.gh-prx.yaml"
	DefaultTitle          = "{{.Type}}{{with .Issue}}({{.}}){{end}}: {{humanize .Description}}"
	DefaultBody           = `{{with .Issue}}Closes #{{.}}.

{{end}}## Description

{{if .AISummary}}{{.AISummary}}{{ else }}{{humanize .Description}}

Change(s) in this PR:
{{range $commit := .Commits}}
* {{$commit}}
{{- end}}
{{- end}}

## PR Checklist

- [ ] Tests are included
- [ ] Documentation is changed or added
`
	DefaultBranchTemplate = "{{.Type}}/{{with .Issue}}{{.}}-{{end}}{{.Description}}"
	DefaultBranchPattern  = `{{.Type}}\/({{.Issue}}-)?{{.Description}}`
)

var (
	DefaultIgnoreCommitsPattern = []string{`^wip`}
	DefaultIssueTypes           = []string{
		"fix", "feat", "chore", "docs", "refactor", "test", "style", "build", "ci", "perf", "revert",
	}
	DefaultVariablePatterns = map[string]string{
		"Type":        `fix|feat|chore|docs|refactor|test|style|build|ci|perf|revert`,
		"Issue":       `([a-zA-Z]+\-)*[0-9]+`,
		"Author":      `[a-zA-Z0-9]+`,
		"Description": `.*`,
	}
	DefaultTokenSeparators = []string{"-", "_"}
	Providers              = []string{"github", "jira", "linear"}
	DefaultProvider        = "github"
	ErrInvalidProvider     = errors.New("Invalid provider")
)

type RepositoryConfig struct {
	Branch                  BranchConfig      `yaml:"branch"`
	PR                      PullRequestConfig `yaml:"pr"`
	Issue                   IssueConfig       `yaml:"issue"`
	CheckoutNew             CheckoutNewConfig `yaml:"checkout_new"`
	PullRequestTemplatePath string            `yaml:"pull_request_template_path"`
}

func (c *RepositoryConfig) SetDefaults() {
	c.Branch.SetDefaults()
	c.PR.SetDefaults()
	c.Issue.SetDefaults()
	c.CheckoutNew.SetDefaults()

	if c.PullRequestTemplatePath == "" {
		c.PullRequestTemplatePath = ".github/pull_request_template.md"
	}
}

func (c *RepositoryConfig) Validate() error {
	var merr *multierror.Error

	if err := c.Branch.Validate(); err != nil {
		merr = multierror.Append(merr, errors.Wrap(err, "branch"))
	}

	if err := c.Issue.Validate(); err != nil {
		merr = multierror.Append(merr, errors.Wrap(err, "issue"))
	}

	if err := merr.ErrorOrNil(); err != nil {
		return errors.Wrap(err, "Invalid repository config")
	}

	return nil
}

type IssueConfig struct {
	Provider string   `yaml:"provider"`
	Types    []string `yaml:"types"`
}

func (c *IssueConfig) SetDefaults() {
	if c.Provider == "" {
		c.Provider = DefaultProvider
	}

	if len(c.Types) == 0 {
		c.Types = DefaultIssueTypes
	}
}

func (c *IssueConfig) Validate() error {
	if !lo.Contains(Providers, c.Provider) {
		return errors.Wrapf(ErrInvalidProvider,
			"Provider must be one of %s",
			strings.Join(Providers, ", "),
		)
	}

	return nil
}

type BranchConfig struct {
	// The template structure of your branch names.
	// Example pattern:
	// {{.Type}}/{{.Author}}-{{.Issue}}-{{.Description}}
	//
	// Example branch names:
	// - feature/PROJ-1234-add-foo
	// - chore-PROJ-1234-update-deps
	// - name-fix-PROJ-1234-wrong-thing
	// - bug-name-PROJ-1234-fix-thing
	// - PROJ-1234-some-new-features
	Template string `yaml:"template"`

	// The pattern that should be used to validate the branch name and extract variables.
	//
	// Example: {{.Type}}\/({{.Issue}}-)?{{.Description}}
	// Where the variables are defined in the `variable_patterns` field.
	Pattern string `yaml:"pattern"`

	// The patterns that should be fetched from the branch name.
	//
	// Example:
	//   "Issue": "([A-Z]+\\-[0-9]+)"
	//   "Author": "([a-z0-9]+)"
	//   "Description": "(.*)"
	//   "Type": "(fix|feat|chore|docs|refactor|test|style|build|ci|perf|revert)"
	VariablePatterns map[string]string `yaml:"variable_patterns"`

	TokenSeparators []string `yaml:"token_separators"`

	MaxLength int `yaml:"max_length"`
}

func (c *BranchConfig) SetDefaults() {
	if c.Template == "" {
		c.Template = DefaultBranchTemplate
	}

	if c.Pattern == "" {
		c.Pattern = DefaultBranchPattern
	}

	if c.VariablePatterns == nil {
		c.VariablePatterns = DefaultVariablePatterns
	}

	if len(c.TokenSeparators) == 0 {
		c.TokenSeparators = DefaultTokenSeparators
	}
	c.TokenSeparators = append(c.TokenSeparators, "/")

	if c.MaxLength == 0 {
		c.MaxLength = 60
	}
}

func (c *BranchConfig) Validate() error {
	for _, tokenSeparator := range c.TokenSeparators {
		if len(tokenSeparator) != 1 {
			return errors.Errorf("token_separators: Invalid token separator '%s': Should be exactly 1 character", tokenSeparator)
		}
	}

	return nil
}

type PullRequestConfig struct {
	Title                 string   `yaml:"title"`
	IgnoreCommitsPatterns []string `yaml:"ignore_commits_patterns"`
	AnswerChecklist       *bool    `yaml:"answer_checklist"`
	PushToRemote          *bool    `yaml:"push_to_remote"`

	Body string `yaml:"-"`
}

func (c *PullRequestConfig) SetDefaults() {
	if c.Title == "" {
		c.Title = DefaultTitle
	}

	if c.Body == "" {
		c.Body = DefaultBody
	}

	if len(c.IgnoreCommitsPatterns) == 0 {
		c.IgnoreCommitsPatterns = DefaultIgnoreCommitsPattern
	}

	if c.AnswerChecklist == nil {
		trueVal := true
		c.AnswerChecklist = &trueVal
	}

	if c.PushToRemote == nil {
		trueVal := true
		c.PushToRemote = &trueVal
	}
}

type CheckoutNewConfig struct {
	Jira   CheckoutNewJiraConfig   `yaml:"jira"`
	GitHub CheckoutNewGitHubConfig `yaml:"github"`
}

func (c *CheckoutNewConfig) SetDefaults() {
	c.Jira.SetDefaults()
	c.GitHub.SetDefaults()
}

type CheckoutNewJiraConfig struct {
	IssueJQL string `yaml:"issue_jql"`
	Project  string `yaml:"project"`
}

func (c *CheckoutNewJiraConfig) SetDefaults() {
	if c.IssueJQL == "" {
		if c.Project != "" {
			c.IssueJQL = fmt.Sprintf("project=%s+AND+", c.Project)
		}
		c.IssueJQL += "assignee=currentUser()+AND+statusCategory!=Done+ORDER+BY+updated+DESC"
	}
}

type CheckoutNewGitHubConfig struct {
	IssueListFlags []string `yaml:"issue_list_flags"`
}

func (c *CheckoutNewGitHubConfig) SetDefaults() {
	if len(c.IssueListFlags) == 0 {
		c.IssueListFlags = []string{"--state", "open", "--assignee", "@me"}
	}
}

func LoadRepositoryConfig(globalRepoConfig *RepositoryConfig) (*RepositoryConfig, error) {
	cfg := &RepositoryConfig{}

	if actualConfigFilepath, err := utils.FindRelativePathInRepo(DefaultConfigFilepath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, errors.Wrap(err, "Failed to load config")
		}
		log.Infof("No config file found at '%s', using defaults", DefaultConfigFilepath)
	} else {
		log.Debug(fmt.Sprintf("Loading repository config from '%s'", actualConfigFilepath))
		if err := utils.ReadYaml(actualConfigFilepath, cfg); err != nil {
			return nil, errors.Wrap(err, "Failed to load config")
		}
	}

	if globalRepoConfig != nil {
		if err := mergo.Merge(cfg, globalRepoConfig); err != nil {
			return nil, errors.Wrap(err, "Failed to merge global config to repository config")
		}
	}

	cfg.SetDefaults()

	cfgBytes, _ := json.MarshalIndent(cfg, "", "  ")
	log.Debug(fmt.Sprintf("Loaded repository config: %s", string(cfgBytes)))

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}
