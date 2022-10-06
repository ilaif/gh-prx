package utils

import (
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/samber/lo"
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
	}, nil
}
