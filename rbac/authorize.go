package rbac

import (
	"net/http"
)

// Authorize a request based on its role, resource, and endpoint
func (rbac RBAC) Authorize(r *http.Request, role, resource, endpoint string) error {
	permission, exists := rbac[role][resource][endpoint]
	if !exists {
		return ErrRoleUnknown
	}

	if !permission.Allow {
		return ErrForbidden
	}

	// Ensure query compliance
	err := permission.Ensure.QueryComplies(r)
	if err != nil {
		return err
	}

	// Enforce query compliance
	err = permission.Enforce.QueryComplies(r)
	if err != nil {
		return err
	}

	return nil
}
