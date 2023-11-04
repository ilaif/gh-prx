package config

import (
	"os"
	"path"

	"github.com/caarlos0/log"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"github.com/ilaif/gh-prx/pkg/utils"
)

type SetupConfig struct {
	JiraConfig   JiraConfig   `yaml:"jira"`
	LinearConfig LinearConfig `yaml:"linear"`

	// RepositoryConfig a global config for all repositories.
	// Per-repository config properties will override this one.
	RepositoryConfig RepositoryConfig `yaml:"global"`
}

func (c *SetupConfig) SetDefaults() {
	c.JiraConfig.SetDefaults()
	c.LinearConfig.SetDefaults()
}

type JiraConfig struct {
	Endpoint string `yaml:"endpoint"`
	User     string `yaml:"user"`
	Token    string `yaml:"token"`
}

func (c *JiraConfig) SetDefaults() {
}

func (c *JiraConfig) Validate() error {
	var merr *multierror.Error

	if c.Endpoint == "" {
		merr = multierror.Append(merr, errors.New("Jira endpoint is missing"))
	}

	if c.User == "" {
		merr = multierror.Append(merr, errors.New("Jira user is missing"))
	}

	if c.Token == "" {
		merr = multierror.Append(merr, errors.New("Jira token is missing"))
	}

	if err := merr.ErrorOrNil(); err != nil {
		return errors.Wrap(err, "Invalid Jira config, please run 'gh prx setup provider jira'")
	}

	return nil
}

type LinearConfig struct {
	APIKey string `yaml:"api_key"`
}

func (c *LinearConfig) SetDefaults() {
	if c.APIKey == "" {
		if apiKey := os.Getenv("LINEAR_API_KEY"); apiKey != "" {
			c.APIKey = apiKey
		}
	}
}

func (c *LinearConfig) Validate() error {
	var merr *multierror.Error

	if c.APIKey == "" {
		merr = multierror.Append(merr, errors.New("Linear API key is missing"))
	}

	if err := merr.ErrorOrNil(); err != nil {
		return errors.Wrap(err, "Invalid Linear config, please run 'gh prx setup provider linear'")
	}

	return nil
}

func LoadSetupConfig() (*SetupConfig, error) {
	log.Debug("Loading setup config")
	cfgDir, err := getSetupConfigDir()
	if err != nil {
		return nil, err
	}

	filename := path.Join(cfgDir, "config.yaml")
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return &SetupConfig{}, nil
		}

		return nil, errors.Wrap(err, "Failed to check if setup config file exists")
	}

	cfg := &SetupConfig{}
	if err := utils.ReadYaml(filename, cfg); err != nil {
		return nil, errors.Wrap(err, "Failed to load setup config")
	}

	cfg.SetDefaults()

	return cfg, nil
}

func SaveSetupConfig(cfg *SetupConfig) error {
	cfgDir, err := getSetupConfigDir()
	if err != nil {
		return err
	}

	if _, err := os.Stat(cfgDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(cfgDir, os.ModePerm); err != nil {
				return errors.Wrap(err, "Failed to create config dir")
			}
		} else {
			return errors.Wrapf(err, "Failed to check if config dir '%s' exists", cfgDir)
		}
	}

	filename := path.Join(cfgDir, "config.yaml")

	log.Infof("Saving config to %s", filename)

	if err := utils.WriteYaml(filename, cfg); err != nil {
		return errors.Wrap(err, "Failed to save config")
	}

	return nil
}

func getSetupConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "Failed to get user home dir")
	}

	cfgDir := path.Join(homeDir, "./.config/gh-prx")

	return cfgDir, nil
}
