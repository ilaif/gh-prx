package setup

import (
	"context"

	"github.com/MakeNowJust/heredoc"
	"github.com/caarlos0/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/ilaif/gh-prx/pkg/config"
)

type ProviderOpts struct {
	Endpoint string
	User     string
	Token    string
}

func NewProviderCmd() *cobra.Command {
	opts := &ProviderOpts{}

	cmd := &cobra.Command{
		Use:       "provider <provider>",
		Short:     "Setup a provider.",
		ValidArgs: config.Providers,
		Args:      cobra.ExactArgs(1),
		Long: heredoc.Docf(`
			Setup a provider. A provider is a service that hosts issues.

			Currently supported providers are:
			- github:
				- Configured by running %[1]sgh auth login%[1]s
			- jira:
				- %[1]suser%[1]s is your email address
				- %[1]stoken%[1]s can be created at https://id.atlassian.com/manage-profile/security/api-tokens
				- %[1]sendpoint%[1]s is your jira server: https://<your-jira-server>.atlassian.net
		`, "`"),
		Example: heredoc.Doc(`
			$ gh prx setup provider jira --user <email> --token <token>
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			return setupProvider(ctx, args[0], opts)
		},
	}

	fl := cmd.Flags()
	fl.StringVarP(&opts.Endpoint, "endpoint", "e", "", "Endpoint of the provider.")
	fl.StringVarP(&opts.User, "user", "u", "", "The user to use for the provider.")
	fl.StringVarP(&opts.Token, "token", "t", "", "The token to use for the provider.")

	return cmd
}

func setupProvider(ctx context.Context, provider string, opts *ProviderOpts) error {
	log.Infof("Setting up provider '%s'", provider)

	cfg, err := config.LoadSetupConfig()
	if err != nil {
		return errors.Wrap(err, "Failed to load setup config")
	}

	switch provider {
	case "jira":
		if opts.Endpoint == "" || opts.User == "" || opts.Token == "" {
			return errors.New("endpoint, user and token are required for the jira provider setup")
		}

		cfg.JiraConfig.Endpoint = opts.Endpoint
		cfg.JiraConfig.User = opts.User
		cfg.JiraConfig.Token = opts.Token
	default:
		return config.ErrInvalidProvider
	}

	if err := config.SaveSetupConfig(cfg); err != nil {
		return errors.Wrap(err, "Failed to save setup config")
	}

	log.Infof("Successfully setup provider '%s'", provider)

	return nil
}
