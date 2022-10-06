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
	JiraConfig JiraConfig `yaml:"jira"`
}

type JiraConfig struct {
	Endpoint string `yaml:"endpoint"`
	User     string `yaml:"user"`
	Token    string `yaml:"token"`
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

func LoadSetupConfig() (*SetupConfig, error) {
	cfgDir, err := getSetupConfigDir()
	if err != nil {
		return nil, err
	}

	filename := path.Join(cfgDir, "config.yaml")
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return &SetupConfig{}, nil
		}

		return nil, errors.Wrap(err, "failed to check if config file exists")
	}

	cfg := &SetupConfig{}
	if err := utils.ReadYaml(filename, cfg); err != nil {
		return nil, errors.Wrap(err, "failed to load config")
	}

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
				return errors.Wrap(err, "failed to create config dir")
			}
		} else {
			return errors.Wrapf(err, "failed to check if config dir '%s' exists", cfgDir)
		}
	}

	filename := path.Join(cfgDir, "config.yaml")

	log.Infof("Saving config to %s", filename)

	if err := utils.WriteYaml(filename, cfg); err != nil {
		return errors.Wrap(err, "failed to save config")
	}

	return nil
}

func getSetupConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "failed to get user home dir")
	}

	cfgDir := path.Join(homeDir, "./.config/gh-prx")

	return cfgDir, nil
}
