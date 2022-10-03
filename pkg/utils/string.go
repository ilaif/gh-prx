package utils

import (
	"strings"

	"github.com/samber/lo"
)

func RemoveConsecutiveDuplicates(s string, chars []rune) string {
	var sb strings.Builder
	var last rune

	for i, r := range s {
		if !lo.Contains(chars, r) || r != last || i == 0 {
			sb.WriteRune(r)
			last = r
		}
	}

	return sb.String()
}
