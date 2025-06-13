package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// The set of roles that can be used.
var (
	ADMIN       = NewOrgRole("ADMIN")
	LIMITED     = NewOrgRole("LIMITED")
	FULL_ACCESS = NewOrgRole("FULL_ACCESS")
)

// =============================================================================

// Set of known roles.
var orgRoles = make(map[string]OrgRole)

// OrgRole represents a role in the system.
// @Description: A role is a set of permissions that can be assigned to a user for an organization.
type OrgRole struct {
	// value is the string representation of the role.
	// @Description: The string representation of the role.
	value string
}

func NewOrgRole(role string) OrgRole {
	r := OrgRole{role}
	orgRoles[role] = r
	return r
}

// String returns the name of the role.
func (r *OrgRole) String() string {
	return r.value
}

// Equal provides support for the go-cmp package and testing.
func (r *OrgRole) Equal(r2 OrgRole) bool {
	return r.value == r2.value
}

// MarshalText provides support for logging and any marshal needs.
func (r *OrgRole) MarshalText() ([]byte, error) {
	return []byte(r.value), nil
}

// MarshalJSON implements the json.Marshaler interface.
func (r OrgRole) MarshalJSON() ([]byte, error) {
	return []byte(`"` + r.value + `"`), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (r *OrgRole) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseOrgRole(s)
	if err != nil {
		return err
	}
	*r = parsed
	return nil
}

// Scan Implement the sql.Scanner interface.
func (r *OrgRole) Scan(src any) error {
	// Handle NULL database values (if your column can ever be NULL)
	if src == nil {
		*r = OrgRole{}
		return nil
	}

	var strVal string
	switch v := src.(type) {
	case string:
		strVal = v
	case []byte:
		strVal = string(v)
	default:
		return fmt.Errorf("cannot scan type %T into OrgRole", src)
	}

	parsedRole, err := ParseOrgRole(strVal)
	if err != nil {
		return err
	}
	*r = parsedRole
	return nil
}

// Value Implement the driver.Valuer interface.
func (r OrgRole) Value() (driver.Value, error) {
	return r.value, nil
}

// =============================================================================

// ParseOrgRole ParseAppRole parses the string value and returns a role if one exists.
func ParseOrgRole(value string) (OrgRole, error) {
	role, exists := orgRoles[value]
	if !exists {
		return OrgRole{}, fmt.Errorf("invalid role %q", value)
	}

	return role, nil
}

// MustParseOrgRole parses the string value and returns an org role if one exists. If
// an error occurs the function panics.
func MustParseOrgRole(value string) OrgRole {
	role, err := ParseOrgRole(value)
	if err != nil {
		panic(err)
	}

	return role
}

// ParseOrgRoleToString takes a collection of org roles and converts them to
// a slice of string.
func ParseOrgRoleToString(orgRoles []OrgRole) []string {
	roles := make([]string, len(orgRoles))
	for i, role := range orgRoles {
		roles[i] = (&role).String()
	}

	return roles
}

// ParseManyOrgRoles takes a collection of strings and converts them to a slice
// of roles.
func ParseManyOrgRoles(roles []string) ([]OrgRole, error) {
	orgRoles := make([]OrgRole, len(roles))
	for i, roleStr := range roles {
		role, err := ParseOrgRole(roleStr)
		if err != nil {
			return nil, err
		}
		orgRoles[i] = role
	}

	return orgRoles, nil
}

func (r *OrgRole) UnmarshalText(text []byte) error {
	s := string(text)
	if s == "" {
		// Set a default role if none is provided.
		return fmt.Errorf("role cannot be empty")
	}
	switch s {
	case "ADMIN", "LIMITED", "FULL_ACCESS":
		*r = MustParseOrgRole(s)
		return nil
	default:
		return fmt.Errorf("invalid role %q", s)
	}
}
