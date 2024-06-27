package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	ent "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	grant "github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	clients "github.com/shenzhencenter/google-ads-pb/clients"
	servicespb "github.com/shenzhencenter/google-ads-pb/services"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/metadata"
)

const roleMembership = "member"

var roles = map[string]string{
	"UNKNOWN":    "unknown",
	"ADMIN":      "admin",
	"STANDARD":   "standard",
	"READ_ONLY":  "read_only",
	"EMAIL_ONLY": "email_only",
}

type roleBuilder struct {
	resourceType    *v2.ResourceType
	developerToken  string
	customerID      string
	credentialsJSON string
}

func (r *roleBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return r.resourceType
}

// Create a new connector resource for Google Ads role.
func roleResource(key, role string) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"role_name": role,
		"role_id":   key,
	}

	roleOptions := []rs.RoleTraitOption{
		rs.WithRoleProfile(profile),
	}

	ret, err := rs.NewRoleResource(
		role,
		roleResourceType,
		key,
		roleOptions,
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

// List returns all the roles from the database as resource objects.
// Roles include a RoleTrait because they are the 'shape' of a standard role.
func (r *roleBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var rv []*v2.Resource
	for key, role := range roles {
		ur, err := roleResource(key, role)
		if err != nil {
			return nil, "", nil, fmt.Errorf("error creating role resource for role %s: %w", role, err)
		}
		rv = append(rv, ur)
	}

	return rv, "", nil, nil
}

func (r *roleBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement
	assignmentOptions := []ent.EntitlementOption{
		ent.WithGrantableTo(userResourceType),
		ent.WithDisplayName(fmt.Sprintf("%s Role %s", resource.DisplayName, roleMembership)),
		ent.WithDescription(fmt.Sprintf("Member of %s role", resource.DisplayName)),
	}

	rv = append(rv, ent.NewAssignmentEntitlement(
		resource,
		roleMembership,
		assignmentOptions...,
	))

	return rv, "", nil, nil
}

func (r *roleBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	query := fmt.Sprintf(`SELECT customer_user_access.user_id, customer_user_access.email_address, customer_user_access.access_role FROM customer_user_access WHERE customer_user_access.access_role = "%s"`, resource.Id.Resource)
	headers := metadata.Pairs(
		"developer-token", r.developerToken,
	)

	if r.customerID != "" {
		headers.Append("login-customer-id", r.customerID)
	}

	ctx = metadata.NewOutgoingContext(ctx, headers)
	// TODO: possible issue with credentials, need to test with different account
	c, err := clients.NewGoogleAdsClient(ctx, option.WithCredentialsFile(r.credentialsJSON))
	if err != nil {
		return nil, "", nil, fmt.Errorf("error creating google ads client: %w", err)
	}
	defer c.Close()

	req := &servicespb.SearchGoogleAdsRequest{
		CustomerId: r.customerID,
		Query:      query,
		PageToken:  pToken.Token,
	}

	it := c.Search(ctx, req)
	var rv []*v2.Grant
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, "", nil, fmt.Errorf("error iterating over google ads results: %w", err)
		}

		userAccess := resp.CustomerUserAccess
		ur, err := userResource(userAccess, nil)
		if err != nil {
			return nil, "", nil, fmt.Errorf("error creating user resource: %w", err)
		}
		gr := grant.NewGrant(resource, roleMembership, ur.Id)
		rv = append(rv, gr)
	}
	return rv, "", nil, nil
}

func newRoleBuilder(developerToken, customerID, credentialsJSON string) *roleBuilder {
	return &roleBuilder{
		resourceType:    roleResourceType,
		developerToken:  developerToken,
		customerID:      customerID,
		credentialsJSON: credentialsJSON,
	}
}
