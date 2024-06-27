package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	clients "github.com/shenzhencenter/google-ads-pb/clients"
	servicespb "github.com/shenzhencenter/google-ads-pb/services"
	"google.golang.org/api/option"
	"google.golang.org/grpc/metadata"
)

type accountBuilder struct {
	resourceType    *v2.ResourceType
	developerToken  string
	customerID      string
	credentialsJSON string
}

func (a *accountBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return a.resourceType
}

// Create a new connector resource for Google Ads account.
func accountResource(resource string) (*v2.Resource, error) {
	ret, err := rs.NewResource(
		resource,
		accountResourceType,
		resource,
	)

	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (a *accountBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	req := &servicespb.ListAccessibleCustomersRequest{}
	headers := metadata.Pairs(
		"developer-token", a.developerToken,
	)
	if a.customerID != "" {
		headers.Append("login-customer-id", a.customerID)
	}

	ctx = metadata.NewOutgoingContext(ctx, headers)
	// TODO: possible issue with credentials, need to test with different account
	c, err := clients.NewCustomerClient(ctx, option.WithCredentialsFile(a.credentialsJSON))
	if err != nil {
		return nil, "", nil, fmt.Errorf("error creating customer client: %w", err)
	}
	defer c.Close()

	accounts, err := c.ListAccessibleCustomers(ctx, req)
	if err != nil {
		return nil, "", nil, fmt.Errorf("error listing accounts: %w", err)
	}

	var rv []*v2.Resource
	for _, resource := range accounts.ResourceNames {
		account, err := accountResource(resource)
		if err != nil {
			return nil, "", nil, fmt.Errorf("error creating account resource for account %s: %w", resource, err)
		}
		rv = append(rv, account)
	}
	return rv, "", nil, nil
}

// Entitlements always returns an empty slice for users.
func (a *accountBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (a *accountBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newAccountBuilder(developerToken, customerID, credentialsJSON string) *accountBuilder {
	return &accountBuilder{
		resourceType:    accountResourceType,
		developerToken:  developerToken,
		customerID:      customerID,
		credentialsJSON: credentialsJSON,
	}
}
