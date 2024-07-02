package connector

import (
	"context"
	"errors"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	clients "github.com/shenzhencenter/google-ads-pb/clients"
	"github.com/shenzhencenter/google-ads-pb/resources"
	servicespb "github.com/shenzhencenter/google-ads-pb/services"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/metadata"
)

type userBuilder struct {
	resourceType    *v2.ResourceType
	developerToken  string
	customerID      string
	credentialsJSON string
}

func (u *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return u.resourceType
}

// Create a new connector resource for Google Ads user.
func userResource(user *resources.CustomerUserAccess, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"first_name": user.EmailAddress,
		"login":      user.EmailAddress,
		"user_id":    user.UserId,
	}

	userTraitOptions := []rs.UserTraitOption{
		rs.WithUserProfile(profile),
		rs.WithEmail(*user.EmailAddress, true),
	}

	ret, err := rs.NewUserResource(
		*user.EmailAddress,
		userResourceType,
		user.UserId,
		userTraitOptions,
		rs.WithParentResourceID(parentResourceID),
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (u *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	query := `SELECT customer_user_access.user_id, customer_user_access.email_address, customer_user_access.access_role FROM customer_user_access`
	headers := metadata.Pairs(
		"developer-token", u.developerToken,
	)

	if u.customerID != "" {
		headers.Append("login-customer-id", u.customerID)
	}

	ctx = metadata.NewOutgoingContext(ctx, headers)
	// TODO: possible issue with credentials, need to test with different account
	c, err := clients.NewGoogleAdsClient(ctx, option.WithCredentialsFile(u.credentialsJSON))
	if err != nil {
		return nil, "", nil, fmt.Errorf("error creating google ads client: %w", err)
	}
	defer c.Close()

	req := &servicespb.SearchGoogleAdsRequest{
		CustomerId: u.customerID,
		Query:      query,
		PageToken:  pToken.Token,
	}

	var rv []*v2.Resource
	it := c.Search(ctx, req)
	for {
		resp, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return nil, "", nil, fmt.Errorf("error iterating over google ads results: %w", err)
		}

		userAccess := resp.CustomerUserAccess
		ur, err := userResource(userAccess, parentResourceID)
		if err != nil {
			return nil, "", nil, fmt.Errorf("error creating user resource: %w", err)
		}
		rv = append(rv, ur)
	}

	return rv, "", nil, nil
}

// Entitlements always returns an empty slice for users.
func (u *userBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (u *userBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newUserBuilder(developerToken, customerID, credentialsJSON string) *userBuilder {
	return &userBuilder{
		resourceType:    userResourceType,
		developerToken:  developerToken,
		customerID:      customerID,
		credentialsJSON: credentialsJSON,
	}
}
