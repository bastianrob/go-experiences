package rbac_test

import (
	"context"
	"testing"

	"github.com/bastianrob/go-experiences/rbac"
	"github.com/stretchr/testify/assert"
)

func TestRule_FromContext(t *testing.T) {
	tests := []struct {
		given  string
		then   string
		rule   rbac.Rule
		ctx    func() context.Context
		want   interface{}
		panics bool
	}{{
		given: "Non ctx rule.Value", then: "return value should be rule.Value as is",
		rule: rbac.Rule{Value: "something"},
		ctx:  func() context.Context { return context.Background() },
		want: "something",
	}, {
		given: "rule.Value with ctx", then: "return value should be taken from ctx",
		rule: rbac.Rule{Value: "ctx.email"},
		ctx: func() context.Context {
			return context.WithValue(context.Background(), rbac.ContextKey("email"), "someone@email.com")
		},
		want: "someone@email.com",
	}, {
		given: "rule.Value with deep nested ctx", then: "return value should be taken from ctx",
		rule: rbac.Rule{Value: "ctx.access.id"},
		ctx: func() context.Context {
			return context.WithValue(context.Background(), rbac.ContextKey("access"), map[string]interface{}{
				"id": "IDX-0001",
			})
		},
		want: "IDX-0001",
	}, {
		given: "rule.Value with deep nested ctx, but at 4th level its not a map", then: "code should panic",
		rule: rbac.Rule{Value: "ctx.access.id.name"},
		ctx: func() context.Context {
			return context.WithValue(context.Background(), rbac.ContextKey("access"), map[string]interface{}{
				"id": "IDX-0001",
			})
		},
		panics: true,
	}, {
		given: "rule.Value with deep nested ctx, but does not exists", then: "code should panic",
		rule: rbac.Rule{Value: "ctx.something.not.exists"},
		ctx: func() context.Context {
			return context.Background()
		},
		panics: true,
	}}
	for _, tt := range tests {
		t.Run(tt.given, func(t *testing.T) {
			if !tt.panics {
				got := tt.rule.FromContext(tt.ctx())
				assert.Equal(t, tt.want, got, tt.then)
			} else {
				assert.Panics(t, func() {
					tt.rule.FromContext(tt.ctx())
				}, tt.given)
			}
		})
	}
}

func TestRule_Comply(t *testing.T) {
	type args struct {
		expected interface{}
		actual   interface{}
	}
	tests := []struct {
		given string
		then  string
		rule  rbac.Rule
		args  args
		want  bool
	}{{
		given: "With rule: actual must be = expected", then: "query complies with our rule",
		rule: rbac.Rule{
			Operator: "=",
		},
		args: args{
			expected: "something",
			actual:   "something",
		},
		want: true,
	}, {
		given: "With rule: actual must be != expected", then: "query complies with our rule",
		rule: rbac.Rule{
			Operator: "!=",
		},
		args: args{
			expected: "something",
			actual:   "another",
		},
		want: true,
	}, {
		given: "With rule operator not known", then: "query does not complies",
		rule: rbac.Rule{
			Operator: "unknwon",
		},
		want: false,
	}}
	for _, tt := range tests {
		t.Run(tt.given, func(t *testing.T) {
			got := tt.rule.Comply(tt.args.expected, tt.args.actual)
			assert.Equal(t, tt.want, got, tt.then)
		})
	}
}
