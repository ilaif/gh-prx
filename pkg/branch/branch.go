package branch

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/AlecAivazis/survey/v2"
	"github.com/caarlos0/log"
	"github.com/pkg/errors"

	"github.com/ilaif/gh-prx/pkg/config"
	"github.com/ilaif/gh-prx/pkg/models"
	"github.com/ilaif/gh-prx/pkg/utils"
)

func ParseBranch(name string, cfg config.BranchConfig) (models.Branch, error) {
	log.Debugf("Parsing branch name '%s'", name)

	branchPattern := cfg.Pattern
	for placeholder, pattern := range cfg.VariablePatterns {
		namedGroupPattern := fmt.Sprintf("(?P<%s>%s)", placeholder, pattern)
		branchPattern = strings.ReplaceAll(branchPattern, fmt.Sprintf("{{.%s}}", placeholder), namedGroupPattern)
	}

	branchRegexp, err := regexp.Compile(branchPattern)
	if err != nil {
		return models.Branch{}, errors.Wrap(err, "Failed to compile branch pattern")
	}

	branch := models.Branch{
		Original: name,
		Fields:   map[string]any{},
	}

	matches := branchRegexp.FindStringSubmatch(name)
	if len(matches) == 0 {
		return models.Branch{}, errors.Errorf("Failed to parse branch name '%s' with pattern '%s'", name, branchPattern)
	}
	for i, name := range branchRegexp.SubexpNames() {
		if i != 0 && name != "" {
			branch.Fields[name] = matches[i]
		}
	}

	return branch, nil
}

func TemplateBranchName(cfg *config.Config, issue *models.Issue) (string, error) {
	log.Debug("Templating branch name")

	funcMaps, err := utils.GenerateTemplateFunctions(cfg.Branch.TokenSeparators)
	if err != nil {
		return "", errors.Wrap(err, "Failed to generate template functions")
	}

	tpl := bytes.Buffer{}
	t, err := template.New("branch-name-tpl").Funcs(funcMaps).Parse(cfg.Branch.Template)
	if err != nil {
		return "", errors.Wrap(err, "Failed to parse branch name template")
	}

	issueType, err := resolveIssueType(issue, cfg.Issue.Types)
	if err != nil {
		return "", errors.Wrap(err, "Failed to resolve issue type")
	}

	if err := t.Execute(&tpl, map[string]interface{}{
		"Type":        issueType,
		"Issue":       issue.Key,
		"Description": issue.NormalizedTitle(),
	}); err != nil {
		return "", errors.Wrap(err, "Failed to template branch name")
	}

	name := normalizeBranchName(tpl.String(), cfg.Branch.TokenSeparators)

	if len(name) > cfg.Branch.MaxLength {
		var userReply bool
		if err := survey.AskOne(&survey.Confirm{
			Message: fmt.Sprintf("Branch name is too long, do you want to change it?\n>> %s", name),
		}, &userReply); err != nil {
			return "", errors.Wrap(err, "Failed to prompt for branch name")
		}

		if userReply {
			newName, err := utils.EditString(name)
			if err != nil {
				return "", errors.Wrap(err, "Failed to edit branch name")
			}
			name = normalizeBranchName(newName, cfg.Branch.TokenSeparators)
		}
	}

	return name, nil
}

func resolveIssueType(issue *models.Issue, issueTypes []string) (string, error) {
	issueType := issue.Type

	if issueType == "" {
		log.Info("Could not determine issue type from labels, asking user")

		if err := survey.AskOne(&survey.Select{
			Message: "Choose an issue type",
			Options: issueTypes,
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
	name = strings.Trim(name, "-\n")

	return name
}
