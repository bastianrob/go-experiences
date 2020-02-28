package rbac_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/bastianrob/go-experiences/rbac"
	"github.com/stretchr/testify/assert"
)

func TestEnforcer_QueryComplies(t *testing.T) {
	type args struct {
		method string
		url    string
	}
	tests := []struct {
		given    string
		then     string
		enforcer rbac.Enforcer
		context  func() context.Context
		args     args
		want     map[string]string
		wantErr  bool
	}{{
		given: "Query: id=nil&name=nil and Rule: id=0001&name=ctx.name and ctx.name=John",
		then:  "QueryComplies must not return error, and query must be re-written by enforcer",
		args: args{
			url: "http://api.example.com/resources?id=nil&name=nil",
		},
		enforcer: rbac.Enforcer{
			Query: []rbac.Rule{
				// query: {id: "0001", name: "John"}
				{Key: "id", Value: "0001"},
				{Key: "name", Value: "ctx.name"},
			},
		},
		context: func() context.Context {
			// we give the context.name = "John"
			return context.WithValue(context.Background(), rbac.ContextKey("name"), "John")
		},
		want: map[string]string{
			"id":   "0001",
			"name": "John",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.given, func(t *testing.T) {
			r, _ := http.NewRequest(tt.args.method, tt.args.url, nil)
			r = r.WithContext(tt.context())

			err := tt.enforcer.QueryComplies(r)
			if tt.wantErr {
				assert.Error(t, err, tt.given)
			} else {
				assert.NoError(t, err, tt.given)
				assert.Equal(t, len(tt.want), len(r.URL.Query()))
				for key, val := range tt.want {
					assert.Equal(t, val, r.URL.Query().Get(key), tt.then)
				}
			}
		})
	}
}
