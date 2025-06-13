package models

import (
	"time"

	"github.com/google/uuid"
)

// CreateOrganizationRequest represents a request to create a new organization.
// @Description Request body containing the name for the new organization.
// @name CreateOrganizationRequest
type CreateOrganizationRequest struct {
	// The desired name for the new organization. Must be unique.
	Name string `json:"name" binding:"required" example:"1st MARDIV"`
}

// UpdateOrganizationRequest represents a request to update an organization's properties.
// @Description Request body for modifying an existing organization's name.
// @name UpdateOrganizationRequest
type UpdateOrganizationRequest struct {
	// The new name for the organization. Must be unique.
	Name string `json:"name" binding:"required" example:"1st MARDIV"`
}

// OrganizationResponse represents the detailed view of an organization.
// @Description Contains the full details of an organization, including its ID, name, and timestamps.
// @name OrganizationResponse
type OrganizationResponse struct {
	// The unique identifier of the organization.
	ID uuid.UUID `json:"id" swaggertype:"string,uuid" example:"b7c5e4d3-a2b1-4f0e-8d9c-1a2b3c4d5e6f"`
	// The name of the organization.
	Name string `json:"name" example:"1st MARDIV"`
	// Timestamp indicating when the organization was created.
	CreatedAt time.Time `json:"created_at" example:"2024-03-10T10:00:00Z" format:"date-time"`
	// Timestamp indicating when the organization was last updated.
	UpdatedAt time.Time `json:"updated_at" example:"2024-03-11T15:30:00Z" format:"date-time"`
}

// CreateOrganizationInvitationRequest represents a request to invite a user to an organization.
// @Description Request body containing the email and desired role for the user being invited.
// @name CreateOrganizationInvitationRequest
type CreateOrganizationInvitationRequest struct {
	// The email address of the user to invite. An invitation email will be sent here.
	Email string `json:"email" binding:"required,email" example:"new.member@example.com" format:"email"`
	// The role to assign the user within the organization upon acceptance.
	OrganizationRole OrgRole `json:"organization_role" binding:"required" swaggertype:"string" enums:"ADMIN,FULL_ACCESS,LIMITED"`
}

// OrganizationInvitationResponse represents the detailed view of an organization invitation.
// @Description Contains the full details of an invitation sent to a user to join an organization.
// @name OrganizationInvitationResponse
type OrganizationInvitationResponse struct {
	// The unique identifier of the invitation.
	ID uuid.UUID `json:"id" swaggertype:"string,uuid" example:"c8d6f5e4-b3c2-5g1f-9e0d-2b3c4d5e6f7g"`
	// The unique identifier of the organization the invitation is for.
	OrganizationID uuid.UUID `json:"organization_id" swaggertype:"string,uuid" example:"b7c5e4d3-a2b1-4f0e-8d9c-1a2b3c4d5e6f"`
	// The name of the organization the invitation is for.
	OrganizationName string `json:"organization_name" example:"Example Innovations Inc."`
	// The email address of the invited user.
	Email string `json:"email" example:"new.member@example.com" format:"email"`
	// The role assigned to the user upon acceptance.
	OrganizationRole OrgRole `json:"organization_role" swaggertype:"string" enums:"ADMIN,FULL_ACCESS,LIMITED"`
	// The unique token used to accept or decline the invitation. Included only when listing invitations for the invited user.
	Token string `json:"token,omitempty" example:"eyJhbGciOiJIUzI1NiIsI[...]unique_token_part[...]"`
	// Timestamp indicating when the invitation expires.
	ExpiresAt time.Time `json:"expires_at" example:"2024-03-17T10:00:00Z" format:"date-time"`
	// Timestamp indicating when the invitation was accepted. Null if not yet accepted.
	AcceptedAt *time.Time `json:"accepted_at,omitempty" example:"2024-03-12T09:05:00Z" format:"date-time"`
	// Timestamp indicating when the invitation was created.
	CreatedAt time.Time `json:"created_at" example:"2024-03-10T10:05:00Z" format:"date-time"`
	// Timestamp indicating when the invitation was last updated.
	UpdatedAt time.Time `json:"updated_at" example:"2024-03-10T10:05:00Z" format:"date-time"`
}

// AcceptDeclineOrganizationInvitationRequest represents a request to accept or decline an invitation.
// @Description Request body containing the token required to accept or decline an organization invitation.
// @name AcceptDeclineOrganizationInvitationRequest
type AcceptDeclineOrganizationInvitationRequest struct {
	// The invitation token received via email used to verify the action.
	Token string `json:"token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsI[...]unique_token_part[...]" minlength:"10"`
}

// OrganizationUserResponse represents a user's membership within an organization.
// @Description Contains details about a user's association with a specific organization, including their role.
// @name OrganizationUserResponse
type OrganizationUserResponse struct {
	// The unique identifier of the user.
	UserId uuid.UUID `json:"user_id" swaggertype:"string,uuid" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef"`
	// The user's email address.
	Email *string `json:"email,omitempty" example:"existing.member@example.com" format:"email"`
	// The user's first name.
	FirstName *string `json:"first_name,omitempty" example:"Existing"`
	// The user's last name.
	LastName *string `json:"last_name,omitempty" example:"Member"`
	// If the current user is support
	IsSupport bool `json:"is_support" example:"false"`
	// The unique identifier of the organization.
	OrganizationId uuid.UUID `json:"organization_id" swaggertype:"string,uuid" example:"b7c5e4d3-a2b1-4f0e-8d9c-1a2b3c4d5e6f"`
	// The name of the organization.
	OrganizationName *string `json:"organization_name,omitempty" example:"Example Innovations Inc."`
	// The user's role within this specific organization.
	OrganizationRole OrgRole `json:"organization_role" swaggertype:"string" enums:"ADMIN,FULL_ACCESS,LIMITED"`
	// Timestamp indicating when the user joined the organization.
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T11:00:00Z" format:"date-time"`
	// Timestamp indicating when the user's membership details were last updated.
	UpdatedAt time.Time `json:"updated_at" example:"2024-03-11T16:00:00Z" format:"date-time"`
}

// OrganizationUserWithOrgDetailsResponse represents organization membership details for a user.
// @Description Contains details about an organization a user belongs to, including their role within it.
// @deprecated This seems redundant with OrganizationUserResponse if user details are included there. Consider merging or clarifying purpose.
// @name OrganizationUserWithOrgDetailsResponse
type OrganizationUserWithOrgDetailsResponse struct {
	// The unique identifier of the organization.
	OrganizationId uuid.UUID `json:"organization_id" swaggertype:"string,uuid" example:"b7c5e4d3-a2b1-4f0e-8d9c-1a2b3c4d5e6f"`
	// The unique identifier of the user.
	UserId uuid.UUID `json:"user_id" swaggertype:"string,uuid" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef"`
	// If the current user is support
	IsSupport bool `json:"is_support" example:"false"`
	// The name of the organization.
	OrganizationName string `json:"organization_name" example:"Example Innovations Inc."`
	// The user's role within this specific organization.
	OrganizationRole OrgRole `json:"organization_role" swaggertype:"string" enums:"ADMIN,FULL_ACCESS,LIMITED"`
	// Timestamp indicating when the user joined the organization.
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T11:00:00Z" format:"date-time"`
	// Timestamp indicating when the user's membership details were last updated.
	UpdatedAt time.Time `json:"updated_at" example:"2024-03-11T16:00:00Z" format:"date-time"`
}

// CreateOrganizationUserRequest represents a request to manually add a user to an organization.
// @Description Request body for directly adding an existing user to an organization with a specific role. Use invitations for new users.
// @name CreateOrganizationUserRequest
type CreateOrganizationUserRequest struct {
	// The unique identifier of the organization to add the user to.
	OrganizationID string `json:"organization_id" binding:"required,uuid" example:"b7c5e4d3-a2b1-4f0e-8d9c-1a2b3c4d5e6f" format:"uuid"`
	// The unique identifier of the user to add.
	UserID string `json:"user_id" binding:"required,uuid" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef" format:"uuid"`
	// The role to assign to the user within the organization.
	OrganizationRole OrgRole `json:"organization_role" binding:"required" swaggertype:"string" enums:"ADMIN,FULL_ACCESS,LIMITED"`
}

// UpdateOrganizationUserRequest represents a request to update a user's role in an organization.
// @Description Request body for changing the role of an existing member within an organization.
// @name UpdateOrganizationUserRequest
type UpdateOrganizationUserRequest struct {
	// The new role to assign to the user.
	OrganizationRole OrgRole `json:"organization_role" binding:"required" swaggertype:"string" enums:"ADMIN,FULL_ACCESS,LIMITED"`
}

// OrganizationUserPaginatedResponse represents a paginated response of organization users
// @Description A paginated list of organization users
// @name OrganizationUserPaginatedResponse
type OrganizationUserPaginatedResponse struct {
	// Results is a slice of items of type OrganizationUserResponse.
	Results []OrganizationUserResponse `json:"results"`
	// TotalCount is the total number of items available.
	TotalCount int `json:"total_count" swaggertype:"integer" example:"100"`
	// Paging contains optional paging information.
	Paging Paging `json:"paging,omitempty"`
}
