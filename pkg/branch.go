package pkg

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/AlecAivazis/survey/v2"
	"github.com/caarlos0/log"
	"github.com/pkg/errors"

	"github.com/ilaif/gh-prx/pkg/utils"
)

const (
	DefaultBranchTemplate = "{{.Type}}/{{with .Issue}}{{.}}-{{end}}{{.Description}}"
	DefaultBranchPattern  = `{{.Type}}\/({{.Issue}}-)?{{.Description}}`
)

var (
	DefaultIssueTypes = []string{
		"fix", "feat", "chore", "docs", "refactor", "test", "style", "build", "ci", "perf", "revert",
	}
	VariablePatterns = map[string]string{
		"Type":        `fix|feat|chore|docs|refactor|test|style|build|ci|perf|revert`,
		"Issue":       `[0-9]+`,
		"Description": `.*`,
	}
	DefaultTokenSeparators = []string{"-", "_"}
	LabelToType            = map[string]string{
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
}

func (c *BranchConfig) SetDefaults() {
	if c.Template == "" {
		c.Template = DefaultBranchTemplate
	}

	if c.Pattern == "" {
		c.Pattern = DefaultBranchPattern
	}

	if c.VariablePatterns == nil {
		c.VariablePatterns = VariablePatterns
	}

	if len(c.TokenSeparators) == 0 {
		c.TokenSeparators = DefaultTokenSeparators
	}
	c.TokenSeparators = append(c.TokenSeparators, "/")
}

func (c *BranchConfig) Validate() error {
	for _, tokenSeparator := range c.TokenSeparators {
		if len(tokenSeparator) != 1 {
			return errors.Errorf("token_separators: Invalid token separator '%s': Should be exactly 1 character", tokenSeparator)
		}
	}

	return nil
}

type Branch struct {
	Fields   map[string]any
	Original string
}

func ParseBranch(name string, cfg BranchConfig) (Branch, error) {
	log.Debugf("Parsing branch name '%s'", name)

	branchPattern := cfg.Pattern
	for placeholder, pattern := range cfg.VariablePatterns {
		namedGroupPattern := fmt.Sprintf("(?P<%s>%s)", placeholder, pattern)
		branchPattern = strings.ReplaceAll(branchPattern, fmt.Sprintf("{{.%s}}", placeholder), namedGroupPattern)
	}

	branchRegexp, err := regexp.Compile(branchPattern)
	if err != nil {
		return Branch{}, errors.Wrap(err, "Failed to compile branch pattern")
	}

	branch := Branch{
		Original: name,
		Fields:   map[string]any{},
	}

	matches := branchRegexp.FindStringSubmatch(name)
	if len(matches) == 0 {
		return Branch{}, errors.Errorf("Failed to parse branch name '%s' with pattern '%s'", name, branchPattern)
	}
	for i, name := range branchRegexp.SubexpNames() {
		if i != 0 && name != "" {
			branch.Fields[name] = matches[i]
		}
	}

	return branch, nil
}

func TemplateBranchName(branchCfg BranchConfig, issue *Issue) (string, error) {
	log.Debug("Templating branch name")

	funcMaps, err := generateTemplateFunctions(branchCfg.TokenSeparators)
	if err != nil {
		return "", errors.Wrap(err, "Failed to generate template functions")
	}

	tpl := bytes.Buffer{}
	t, err := template.New("branch-name-tpl").Funcs(funcMaps).Parse(branchCfg.Template)
	if err != nil {
		return "", errors.Wrap(err, "Failed to parse branch name template")
	}

	issueType, err := resolveIssueType(issue)
	if err != nil {
		return "", errors.Wrap(err, "Failed to resolve issue type")
	}

	if err := t.Execute(&tpl, map[string]interface{}{
		"Type":        issueType,
		"Issue":       issue.Code,
		"Description": normalizeIssueTitle(issue.Title),
	}); err != nil {
		return "", errors.Wrap(err, "Failed to template branch name")
	}

	name := normalizeBranchName(tpl.String(), branchCfg.TokenSeparators)

	return name, nil
}

func resolveIssueType(issue *Issue) (string, error) {
	issueType := ""
	for _, label := range issue.Labels {
		if it, ok := LabelToType[label]; ok {
			issueType = it

			break
		}
	}

	if issueType == "" {
		log.Info("Could not determine issue type from labels, asking user")

		if err := survey.AskOne(&survey.Select{
			Message: "Choose an issue type",
			Options: DefaultIssueTypes,
		}, &issueType, survey.WithValidator(survey.Required)); err != nil {
			return "", errors.Wrap(err, "Failed to ask for issue type")
		}
	}

	return issueType, nil
}

func normalizeBranchName(name string, tokenSeparators []string) string {
	runeTokenSeparators := []rune{}
	for _, tokenSep := range tokenSeparators {
		runeTokenSeparators = append(runeTokenSeparators, rune(tokenSep[0]))
	}

	name = utils.RemoveConsecutiveDuplicates(name, runeTokenSeparators)
	name = strings.Trim(name, "-")

	return name
}
