package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	CredentialsJSONField = field.StringField(
		"credentials-json-file-path",
		field.WithDescription("Local path to a Google service-account key JSON file"),
		field.WithRequired(true),
	)

	DeveloperTokenField = field.StringField(
		"developer-token",
		field.WithDescription("Your google ads developer token"),
		field.WithRequired(true),
		field.WithIsSecret(true),
	)

	CustomerIDField = field.StringField(
		"customer-id",
		field.WithDescription("If you are using a manager account to access a client account, you must provide the correct login customer ID"),
	)

	ConfigurationFields = []field.SchemaField{
		CredentialsJSONField,
		DeveloperTokenField,
		CustomerIDField,
	}

	ConfigurationSchema = field.Configuration{
		Fields: ConfigurationFields,
	}
)

//go:generate go run ./gen
var Config = field.NewConfiguration(ConfigurationFields)
