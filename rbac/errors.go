package rbac

import "errors"

// Error collection
var (
	ErrNotString   = errors.New("Expected value is not a string")
	ErrNoRole      = errors.New("You have no role assigned to you")
	ErrRoleUnknown = errors.New("You have an unknown role assigned to you")
	ErrForbidden   = errors.New("You are not allowed to access specified resource")
)
