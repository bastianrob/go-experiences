package rbac_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bastianrob/go-experiences/rbac"
)

func TestEnsurer_QueryComplies(t *testing.T) {
	type args struct {
		method string
		url    string
	}
	tests := []struct {
		given   string
		then    string
		ensurer rbac.Ensurer
		context func() context.Context
		args    args
		wantErr bool
	}{{
		given: "Query: id=0001&name=John and Rule: id=0001&name=ctx.name and ctx.name=John",
		then:  "QueryComplies must not return error",
		args: args{
			url: "http://api.example.com/resources?id=0001&name=John",
		},
		ensurer: rbac.Ensurer{
			Query: []rbac.Rule{
				// query: {id: "0001", name: "John"}
				{Key: "id", Operator: "=", Value: "0001"},
				{Key: "name", Operator: "=", Value: "ctx.name"},
			},
		},
		context: func() context.Context {
			// we give the context.name = "John"
			return context.WithValue(context.Background(), rbac.ContextKey("name"), "John")
		},
	}}
	for _, tt := range tests {
		t.Run(tt.given, func(t *testing.T) {
			r, _ := http.NewRequest(tt.args.method, tt.args.url, nil)
			r = r.WithContext(tt.context())

			err := tt.ensurer.QueryComplies(r)
			if tt.wantErr {
				assert.Error(t, err, tt.given)
			} else {
				assert.NoError(t, err, tt.given)
			}
		})
	}
}
