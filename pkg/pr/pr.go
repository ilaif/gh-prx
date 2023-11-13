package pr

import (
	"bytes"
	"regexp"
	"strings"
	"text/template"

	"github.com/AlecAivazis/survey/v2"
	"github.com/caarlos0/log"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/ilaif/gh-prx/pkg/config"
	"github.com/ilaif/gh-prx/pkg/models"
	"github.com/ilaif/gh-prx/pkg/utils"
)

var (
	TypeToLabel = map[string]string{
		"fix":  "bug",
		"feat": "enhancement",
		"docs": "documentation",
	}

	mdCheckboxMatcher          = regexp.MustCompile(`^\s*[\-\*]\s*\[(x|\s)\]`)
	commitMsgSeparatorMatcher  = regexp.MustCompile(`[\*\-]`)
	mapHasNoEntryForKeyMatcher = regexp.MustCompile(`map has no entry for key "(.*)"`)
)

type CommitsFetcher func() ([]string, error)
type AISummarizer func() (string, error)

func TemplatePR(
	b models.Branch,
	prCfg config.PullRequestConfig,
	confirm bool,
	tokenSeparators []string,
	commitsFetcher CommitsFetcher,
	aiSummarizer func() (string, error),
) (*models.PullRequest, error) {
	log.Debug("Templating PR")

	funcMaps, err := utils.GenerateTemplateFunctions(tokenSeparators)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to generate template functions")
	}

	pr := &models.PullRequest{}

	res := bytes.Buffer{}
	titleTpl, err := template.New("pr-title-tpl").Funcs(funcMaps).Parse(prCfg.Title)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse pr title template")
	}

	for {
		err := titleTpl.Option("missingkey=error").Execute(&res, b.Fields)
		if err == nil {
			break
		}

		matches := mapHasNoEntryForKeyMatcher.FindStringSubmatch(err.Error())

		if len(matches) == 0 {
			return nil, errors.Wrap(err, "Failed to template pr title")
		}

		log.Warn("Missing key in branch fields, prompting user to enter it manually")
		answer := ""
		if err := survey.AskOne(&survey.Input{
			Message: "Enter value for " + matches[1],
		}, &answer, survey.WithValidator(survey.Required)); err != nil {
			return nil, errors.Wrap(err, "Failed to ask for missing field value")
		}

		b.Fields[matches[1]] = answer

		continue
	}

	pr.Title = res.String()

	bodyData := lo.Assign(b.Fields, make(map[string]any))

	if strings.Contains(prCfg.Body, ".Commits") {
		log.Debug("Fetching commits")
		commits, err := fetchCommits(prCfg.IgnoreCommitsPatterns, commitsFetcher)
		if err != nil {
			return nil, err
		}
		bodyData["Commits"] = commits
	}

	if strings.Contains(prCfg.Body, ".AISummary") {
		aiSummary, err := aiSummarizer()
		if err != nil {
			return nil, err
		}
		bodyData["AISummary"] = aiSummary
	}

	res = bytes.Buffer{}
	bodyTpl, err := template.New("pr-body-tpl").Funcs(funcMaps).Parse(prCfg.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse pr body template")
	}
	if err := bodyTpl.Option("missingkey=zero").Execute(&res, bodyData); err != nil {
		return nil, errors.Wrap(err, "Failed to template pr body")
	}
	pr.Body = res.String()

	if *prCfg.AnswerChecklist {
		body, err := answerPRChecklist(pr.Body, confirm)
		if err != nil {
			return nil, err
		}
		pr.Body = body
	}

	if typeAny, ok := b.Fields["Type"]; ok {
		if typeStr, ok := typeAny.(string); ok {
			typeStr = strings.ToLower(typeStr)

			label, ok := TypeToLabel[typeStr]
			if !ok {
				label = typeStr
			}

			pr.Labels = append(pr.Labels, label)
		}
	}

	return pr, nil
}

func fetchCommits(ignoreCommitsPatterns []string, commitsFetcher CommitsFetcher) ([]string, error) {
	commits, err := commitsFetcher()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch commits")
	}

	ignoreCommitsMatcher, err := regexp.Compile(strings.Join(ignoreCommitsPatterns, "|"))
	if err != nil {
		return nil, errors.Wrap(err, "Failed to compile ignore commits matcher")
	}

	splitCommits := []string{}
	for _, commit := range commits {
		splitCommits = append(splitCommits, commitMsgSeparatorMatcher.Split(commit, -1)...)
	}

	splitCommits = lo.Filter(splitCommits, func(commit string, _ int) bool {
		return !ignoreCommitsMatcher.MatchString(commit) && commit != ""
	})

	return lo.Reverse(splitCommits), nil
}

func answerPRChecklist(body string, confirm bool) (string, error) {
	if confirm {
		log.Info("Answering yes to all checklist items (if exists in PR description)")
	}

	bodyLines := strings.Split(body, "\n")
	newBodyLines := []string{}
	for _, line := range bodyLines {
		addLine := true

		if mdCheckboxMatcher.MatchString(line) {
			q := mdCheckboxMatcher.ReplaceAllString(line, "")

			answer := "no"
			if !confirm {
				if err := survey.AskOne(&survey.Select{
					Message: q,
					Options: []string{"yes", "no", "skip"},
				}, &answer, survey.WithValidator(survey.Required)); err != nil {
					return "", errors.Wrap(err, "Failed to ask for action item status")
				}
			}

			switch answer {
			case "yes":
				line = mdCheckboxMatcher.ReplaceAllString(line, "- [x]")
			case "no":
				line = mdCheckboxMatcher.ReplaceAllString(line, "- [ ]")
			case "skip":
				addLine = false
			}
		}

		if addLine {
			newBodyLines = append(newBodyLines, line)
		}
	}

	return strings.Join(newBodyLines, "\n"), nil
}
