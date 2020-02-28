package rbac

import (
	"context"
	"reflect"
	"strings"
)

// Rule of a permission
type Rule struct {
	Key      string `yaml:"key"`
	Operator string `yaml:"operator"`
	Value    string `yaml:"value"`
}

// FromContext get actual rule.Value from ctx if rule.Value starts with ctx
// otherwise, return rule.Value as is
func (rule Rule) FromContext(ctx context.Context) interface{} {
	if !strings.HasPrefix(rule.Value, "ctx") {
		return rule.Value
	}

	paths := strings.Split(rule.Value, ".")
	var ctxval interface{}

	// starts from 1, as we exclude the ctx part
	for i := 1; i < len(paths); i++ {
		ctxkey := paths[i]

		//Get current context index
		if i == 1 {
			ctxval = ctx.Value(ContextKey(ctxkey))
		} else {
			// if rule.Value is nested more than 1 level, we assume the context value is of type map[string]interface{}
			// otherwise, panic
			var ok bool
			kvp := ctxval.(map[string]interface{})
			ctxval, ok = kvp[ctxkey]
			if !ok || ctxval == nil {
				ctxval = nil
			}
		}
	}

	return ctxval
}

// Comply checks does request value complies with our rule
func (rule Rule) Comply(expected, actual interface{}) bool {
	switch rule.Operator {
	case "!=":
		return !reflect.DeepEqual(expected, actual)
	case "=":
		return reflect.DeepEqual(expected, actual)
	}

	// doesn't comply if we don't recognize the rule operator
	return false
}
