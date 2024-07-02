package connector

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
)

type Connector struct {
	developerToken  string
	loginCustomerID string
	credentialsJSON string
}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (c *Connector) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	return []connectorbuilder.ResourceSyncer{
		newUserBuilder(c.developerToken, c.loginCustomerID, c.credentialsJSON),
		newAccountBuilder(c.developerToken, c.loginCustomerID, c.credentialsJSON),
		newRoleBuilder(c.developerToken, c.loginCustomerID, c.credentialsJSON),
	}
}

// Metadata returns metadata about the connector.
func (c *Connector) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "Google ads connector",
		Description: "Connector syncing accounts, users and roles from Google ads.",
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (c *Connector) Validate(ctx context.Context) (annotations.Annotations, error) {
	return nil, nil
}

// New returns a new instance of the connector.
func New(ctx context.Context, jsonCredentialsFile, developerToken, customerID string) (*Connector, error) {
	return &Connector{
		developerToken:  developerToken,
		loginCustomerID: customerID,
		credentialsJSON: jsonCredentialsFile,
	}, nil
}
