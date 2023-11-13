package utils

import (
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func GenerateTemplateFunctions(tokenSeparators []string) (template.FuncMap, error) {
	tokenSeparators = lo.Map(tokenSeparators, func(t string, _ int) string { return fmt.Sprintf("\\%s", t) })
	tokenMatcher, err := regexp.Compile(fmt.Sprintf("[%s]", strings.Join(tokenSeparators, "")))
	if err != nil {
		return nil, errors.Wrap(err, "Failed to compile humanize token matcher")
	}

	return template.FuncMap{
		"humanize": func(text string) (string, error) {
			return tokenMatcher.ReplaceAllString(text, " "), nil
		},
		"title": func(text string) (string, error) {
			return cases.Title(language.English).String(text), nil
		},
		"lower": func(text string) (string, error) {
			return strings.ToLower(text), nil
		},
		"upper": func(text string) (string, error) {
			return strings.ToUpper(text), nil
		},
	}, nil
}
