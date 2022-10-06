package branch

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ilaif/gh-prx/pkg/config"
	"github.com/ilaif/gh-prx/pkg/models"
)

func Test_ParseBranch(t *testing.T) {
	tests := []struct {
		name     string
		config   config.BranchConfig
		expected models.Branch
		err      bool
	}{
		{
			name:   "fix/1234-fix-thing",
			config: config.BranchConfig{},
			expected: models.Branch{
				Fields:   map[string]any{"Type": "fix", "Issue": "1234", "Description": "fix-thing"},
				Original: "fix/1234-fix-thing",
			},
		},
		{
			name:   "feat/1234-add-foo",
			config: config.BranchConfig{},
			expected: models.Branch{
				Fields:   map[string]any{"Type": "feat", "Issue": "1234", "Description": "add-foo"},
				Original: "feat/1234-add-foo",
			},
		},
		{
			name:   "chore-1234-update-deps",
			config: config.BranchConfig{Pattern: "{{.Type}}-{{.Issue}}-{{.Description}}"},
			expected: models.Branch{
				Fields:   map[string]any{"Type": "chore", "Issue": "1234", "Description": "update-deps"},
				Original: "chore-1234-update-deps",
			},
		},
		{
			name:     "bug-name-1234-fix-thing",
			config:   config.BranchConfig{Pattern: "{{.Type}}-{{.Author}}-{{.Issue}}-{{.Description}}"},
			expected: models.Branch{},
			err:      true,
		},
		{
			name:   "1234-some-new-features",
			config: config.BranchConfig{Pattern: "{{.Issue}}-{{.Description}}"},
			expected: models.Branch{
				Fields:   map[string]any{"Issue": "1234", "Description": "some-new-features"},
				Original: "1234-some-new-features",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := assert.New(t)

			test.config.SetDefaults()

			b, err := ParseBranch(test.name, test.config)
			if test.err {
				a.Error(err)
			} else {
				a.NoError(err)
			}
			a.Equal(test.expected, b)
		})
	}
}
