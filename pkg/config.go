package pkg

import (
	"os"

	"github.com/caarlos0/log"
	"github.com/pkg/errors"

	"github.com/ilaif/gh-prx/pkg/utils"
)

const (
	defaultConfigFilepath = ".github/.gh-prx.yaml"
)

type Config struct {
	Branch BranchConfig      `yaml:"branch"`
	PR     PullRequestConfig `yaml:"pr"`
	Issue  IssueConfig       `yaml:"issue"`
}

func (c *Config) SetDefaults() {
	c.Branch.SetDefaults()
	c.PR.SetDefaults()
	c.Issue.SetDefaults()
}

func (c *Config) Validate() error {
	if err := c.Branch.Validate(); err != nil {
		return errors.Wrap(err, "branch")
	}

	if err := c.Issue.Validate(); err != nil {
		return errors.Wrap(err, "issue")
	}

	return nil
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}
	if err := utils.ReadYaml(defaultConfigFilepath, cfg); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, errors.Wrap(err, "Failed to load config")
		}

		log.Infof("No config file found at '%s', using defaults", defaultConfigFilepath)
	}

	cfg.SetDefaults()

	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid config")
	}

	return cfg, nil
}
