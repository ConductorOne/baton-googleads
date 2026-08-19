package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	clients "github.com/shenzhencenter/google-ads-pb/clients"
	"google.golang.org/api/option"
)

// Connector caches the gRPC clients for the lifetime of the sync. The SDK does
// not expose a Close/Cleanup hook for the connector itself — clients are
// released when the process exits. Partial-init failure in New is cleaned up
// inline.
type Connector struct {
	developerToken  string
	loginCustomerID string
	adsClient       *clients.GoogleAdsClient
	customerClient  *clients.CustomerClient
}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (c *Connector) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	return []connectorbuilder.ResourceSyncer{
		newUserBuilder(c.developerToken, c.loginCustomerID, c.adsClient),
		newAccountBuilder(c.developerToken, c.loginCustomerID, c.customerClient),
		newRoleBuilder(c.developerToken, c.loginCustomerID, c.adsClient),
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
	adsClient, err := clients.NewGoogleAdsClient(ctx, option.WithCredentialsFile(jsonCredentialsFile))
	if err != nil {
		return nil, fmt.Errorf("error creating google ads client: %w", err)
	}

	customerClient, err := clients.NewCustomerClient(ctx, option.WithCredentialsFile(jsonCredentialsFile))
	if err != nil {
		adsClient.Close()
		return nil, fmt.Errorf("error creating customer client: %w", err)
	}

	return &Connector{
		developerToken:  developerToken,
		loginCustomerID: customerID,
		adsClient:       adsClient,
		customerClient:  customerClient,
	}, nil
}
