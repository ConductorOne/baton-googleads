package main

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-sdk/pkg/cli"
	"github.com/spf13/cobra"
)

// config defines the external configuration required for the connector to run.
type config struct {
	cli.BaseConfig `mapstructure:",squash"` // Puts the base config options in the same place as the connector options

	CredentialsJSON string `mapstructure:"credentials-json-file-path"`
	DeveloperToken  string `mapstructure:"developer-token"`
	CustomerID      string `mapstructure:"customer-id"`
}

// validateConfig is run after the configuration is loaded, and should return an error if it isn't valid.
func validateConfig(ctx context.Context, cfg *config) error {
	if cfg.CredentialsJSON == "" {
		return fmt.Errorf("credentials json file path is missing, please provide it via --credentials-json-file-path flag or $BATON_CREDENTIALS_JSON_FILE_PATH environment variable")
	}

	if cfg.DeveloperToken == "" {
		return fmt.Errorf("developer token is missing, please provide it via --developer-token flag or $BATON_DEVELOPER_TOKEN environment variable")
	}

	return nil
}

// cmdFlags sets the cmdFlags required for the connector.
func cmdFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String("credentials-json-file-path", "", "JSON credentials file path for the Google ads account credentials. ($BATON_CREDENTIALS_JSON_FILE_PATH)")
	cmd.PersistentFlags().String("developer-token", "", "Your google ads developer token. ($BATON_DEVELOPER_TOKEN)")
	cmd.PersistentFlags().String("customer-id", "", "If you are using a manager account to access a client account, you must provide the correct login customer ID. ($BATON_CUSTOMER_ID)")
}
