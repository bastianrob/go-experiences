package rbac

import (
	"fmt"
	"net/http"
)

// Ensurer data model
// Can either ensure query, header, or path
type Ensurer struct {
	Query  []Rule `yaml:"query"`
	Header []Rule `yaml:"header"`
	Path   []Rule `yaml:"path"`
}

// QueryComplies check whether query request complies with rules
func (ens Ensurer) QueryComplies(r *http.Request) error {
	if ens.Query == nil || len(ens.Query) <= 0 {
		return nil
	}

	ctx := r.Context()
	for _, rule := range ens.Query {
		actual := r.URL.Query().Get(rule.Key)
		expected := rule.FromContext(ctx)

		if !rule.Comply(expected, actual) {
			return fmt.Errorf("Query rule violation: ensure '%s' %s '%v', instead got: '%s'",
				rule.Key, rule.Operator, expected, actual)
		}
	}

	// all query complies with rules
	return nil
}
