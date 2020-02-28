package rbac

import (
	"net/http"
)

// Enforcer structure is just like an Ensurer
type Enforcer Ensurer

// QueryComplies enforce query request from rule
func (enf Enforcer) QueryComplies(r *http.Request) error {
	q := r.URL.Query()
	ctx := r.Context()
	for _, rule := range enf.Query {
		expected := rule.FromContext(ctx)
		valueStr, isString := expected.(string)
		if !isString {
			return ErrNotString
		}

		q.Set(rule.Key, valueStr)
	}

	r.URL.RawQuery = q.Encode()
	// all query enforced with rules
	return nil
}
