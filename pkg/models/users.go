package models

import (
	"time"

	"github.com/google/uuid"
)

// UserResponse represents a single user with their properties.
// @Description A complete user profile with personal information and role.
// @name UserResponse
type UserResponse struct {
	// The unique identifier of the user
	ID uuid.UUID `json:"id" swaggertype:"string,uuid" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef"`
	// The user's email address
	Email string `json:"email" example:"john.doe@example.com" format:"email"`
	// The user's first name
	FirstName string `json:"first_name" example:"John"`
	// The user's last name
	LastName string `json:"last_name" example:"Doe"`
	// The user's role in the application
	UserRole string `json:"user_role" swaggertype:"string" enums:"STANDARD,GLOBAL_ADMIN"`
	// When the user was created
	CreatedAt time.Time `json:"created_at" example:"2024-02-16T21:45:33Z" format:"date-time"`
	// When the user was last updated
	UpdatedAt time.Time `json:"updated_at" example:"2024-02-16T21:45:33Z" format:"date-time"`
}

// CreateUserRequest represents a request to create a new user.
// @Description Request body for creating a new user with their personal information.
// @name CreateUserRequest
type CreateUserRequest struct {
	// The user's email address
	Email string `json:"email" binding:"required,email" example:"jane.doe@example.com" format:"email"`
	// The user's first name
	FirstName string `json:"first_name" binding:"required" example:"Jane"`
	// The user's last name
	LastName string `json:"last_name" binding:"required" example:"Doe"`
}

// UpdateUserRequest represents a request to update a user's information.
// @Description Request body for updating an existing user's personal information.
// @name UpdateUserRequest
type UpdateUserRequest struct {
	// The new email address for the user
	Email string `json:"email,omitempty" binding:"omitempty,email" example:"john.doe.updated@example.com" format:"email"`
	// The new first name for the user
	FirstName string `json:"first_name,omitempty" example:"Johnny"`
	// The new last name for the user
	LastName string `json:"last_name,omitempty" example:"Doe Jr."`
}

// UserScopeResponse details a scope assigned to a user.
// @Description Represents a specific permission scope granted to a user, potentially limited to a specific resource.
// @name UserScopeResponse
type UserScopeResponse struct {
	// The name of the scope
	ScopeName string `json:"scope_name" example:"organizations:read"`
	// The ID of the resource to which this scope applies. If empty or omitted, the scope applies globally for its type.
	ResourceID string `json:"resource_id,omitempty" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef" format:"uuid"`
	// When this scope was assigned
	CreatedAt time.Time `json:"created_at" example:"2024-02-16T21:45:33Z" format:"date-time"`
}

// CreateUserScopeRequest represents a request to assign a scope to a user.
// @Description Request body for assigning a permission scope to a user, optionally targeting a specific resource.
// @name CreateUserScopeRequest
type CreateUserScopeRequest struct {
	// The name of the scope to assign
	ScopeName string `json:"scope_name" binding:"required" example:"organizations:write"`
	// The ID of the resource to which this scope applies. Leave empty or omit for a global scope assignment (within its type).
	ResourceID string `json:"resource_id,omitempty" binding:"omitempty,uuid" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef" format:"uuid"`
}
