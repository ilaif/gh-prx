package models

import (
	"regexp"
	"strings"
)

var (
	InvalidTitleCharsMatcher = regexp.MustCompile(`[^.a-zA-Z0-9]`)
)

type Issue struct {
	Key                 string
	Title               string
	Type                string
	SuggestedBranchName string // Optional, populated for Linear issues
}

func (i *Issue) NormalizedTitle() string {
	title := InvalidTitleCharsMatcher.ReplaceAllString(i.Title, "-")
	title = strings.ToLower(title)
	title = strings.Trim(title, "-")

	return title
}
