package rbac_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bastianrob/go-experiences/rbac"
)

func TestRBAC_Authorize(t *testing.T) {
	rbo := rbac.FromFile("./test.yaml")
	fmt.Printf("%+v", rbo)

	type args struct {
		req      func() *http.Request
		role     string
		resource string
		endpoint string
	}
	tests := []struct {
		given, when, then string
		args              args
		wantErr           bool
		queryResult       map[string]string
	}{
		// As a client
		{
			given: "Role is Client & email = client.one@email.com",
			when:  "?created_by=client.one@email.com", then: "is allowed",
			args: args{
				req: func() *http.Request {
					req, _ := http.NewRequest("", "http://api.example.com/inquiries?created_by=client.one@email.com", nil)
					ctx := context.WithValue(context.Background(), rbac.ContextKey("email"), "client.one@email.com")

					return req.WithContext(ctx)
				},
				role:     "client",
				resource: "inquiry",
				endpoint: "get",
			},
		}, {
			given: "Role is Client & email = client.one@email.com",
			when:  "?created_by=client.other@email.com", then: "is not allowed",
			args: args{
				req: func() *http.Request {
					req, _ := http.NewRequest("", "http://api.example.com/inquiries?created_by=client.other@email.com", nil)
					ctx := context.WithValue(context.Background(), rbac.ContextKey("email"), "client.one@email.com")

					return req.WithContext(ctx)
				},
				role:     "client",
				resource: "inquiry",
				endpoint: "get",
			},
			wantErr: true,
		}, {
			given: "Role is Client & email = client.one@email.com",
			when:  "query is not given", then: "is not allowed",
			args: args{
				req: func() *http.Request {
					req, _ := http.NewRequest("", "http://api.example.com/inquiries", nil)
					ctx := context.WithValue(context.Background(), rbac.ContextKey("email"), "client.one@email.com")

					return req.WithContext(ctx)
				},
				role:     "client",
				resource: "inquiry",
				endpoint: "get",
			},
			wantErr: true,
		}, {
			given: "Role is Client & email = client.one@email.com",
			when:  "trying to create", then: "is allowed",
			args: args{
				req: func() *http.Request {
					req, _ := http.NewRequest("POST", "http://api.example.com/inquiries", nil)
					ctx := context.WithValue(context.Background(), rbac.ContextKey("email"), "client.one@email.com")

					return req.WithContext(ctx)
				},
				role:     "client",
				resource: "inquiry",
				endpoint: "create",
			},
		}, {
			given: "Role is Client & email = client.one@email.com",
			when:  "trying to assign", then: "is not allowed",
			args: args{
				req: func() *http.Request {
					req, _ := http.NewRequest("POST", "http://api.example.com/inquiries/INQ-0001/assign", nil)
					ctx := context.WithValue(context.Background(), rbac.ContextKey("email"), "client.one@email.com")

					return req.WithContext(ctx)
				},
				role:     "client",
				resource: "inquiry",
				endpoint: "assign",
			},
			wantErr: true,
		},

		// As CS
		{
			given: "Role is CS",
			when:  "query is not given", then: "status=New is enforced",
			args: args{
				req: func() *http.Request {
					req, _ := http.NewRequest("", "http://api.example.com/inquiries", nil)
					ctx := context.WithValue(context.Background(), rbac.ContextKey("email"), "cs.one@company.com")

					return req.WithContext(ctx)
				},
				role:     "cs",
				resource: "inquiry",
				endpoint: "get",
			},
			queryResult: map[string]string{
				"status": "New",
			},
		}, {
			given: "Role is CS",
			when:  "query ?status is given", then: "status=New is still enforced",
			args: args{
				req: func() *http.Request {
					req, _ := http.NewRequest("", "http://api.example.com/inquiries?status=Assigned", nil)
					ctx := context.WithValue(context.Background(), rbac.ContextKey("email"), "cs.one@company.com")

					return req.WithContext(ctx)
				},
				role:     "cs",
				resource: "inquiry",
				endpoint: "get",
			},
			queryResult: map[string]string{
				"status": "New",
			},
		}, {
			given: "Role is CS",
			when:  "trying to create", then: "is not allowed",
			args: args{
				req: func() *http.Request {
					req, _ := http.NewRequest("POST", "http://api.example.com/inquiries", nil)
					ctx := context.WithValue(context.Background(), rbac.ContextKey("email"), "cs.one@company.com")

					return req.WithContext(ctx)
				},
				role:     "cs",
				resource: "inquiry",
				endpoint: "create",
			},
			wantErr: true,
		}, {
			given: "Role is CS",
			when:  "trying to assign", then: "is allowed",
			args: args{
				req: func() *http.Request {
					req, _ := http.NewRequest("POST", "http://api.example.com/inquiries/INQ-0001/assign", nil)
					ctx := context.WithValue(context.Background(), rbac.ContextKey("email"), "cs.one@company.com")

					return req.WithContext(ctx)
				},
				role:     "cs",
				resource: "inquiry",
				endpoint: "assign",
			},
		},

		// As an Ops
		{
			given: "Role is Ops & email = ops.one@company.com",
			when:  "?assignee=ops.one@company.com", then: "is allowed",
			args: args{
				req: func() *http.Request {
					req, _ := http.NewRequest("", "http://api.example.com/inquiries?assignee=ops.one@company.com", nil)
					ctx := context.WithValue(context.Background(), rbac.ContextKey("email"), "ops.one@company.com")

					return req.WithContext(ctx)
				},
				role:     "ops",
				resource: "inquiry",
				endpoint: "get",
			},
		}, {
			given: "Role is Ops & email = ops.one@company.com",
			when:  "?assignee=ops.other@company.com", then: "is not allowed",
			args: args{
				req: func() *http.Request {
					req, _ := http.NewRequest("", "http://api.example.com/inquiries?assignee=ops.other@company.com", nil)
					ctx := context.WithValue(context.Background(), rbac.ContextKey("email"), "ops.one@company.com")

					return req.WithContext(ctx)
				},
				role:     "ops",
				resource: "inquiry",
				endpoint: "get",
			},
			wantErr: true,
		}, {
			given: "Role is Ops & email = ops.one@company.com",
			when:  "query is not supplied", then: "is not allowed",
			args: args{
				req: func() *http.Request {
					req, _ := http.NewRequest("", "http://api.example.com/inquiries", nil)
					ctx := context.WithValue(context.Background(), rbac.ContextKey("email"), "ops.one@company.com")

					return req.WithContext(ctx)
				},
				role:     "ops",
				resource: "inquiry",
				endpoint: "get",
			},
			wantErr: true,
		}, {
			given: "Role is Ops & email = ops.one@company.com",
			when:  "trying to created", then: "is not allowed",
			args: args{
				req: func() *http.Request {
					req, _ := http.NewRequest("POST", "http://api.example.com/inquiries", nil)
					ctx := context.WithValue(context.Background(), rbac.ContextKey("email"), "ops.one@company.com")

					return req.WithContext(ctx)
				},
				role:     "ops",
				resource: "inquiry",
				endpoint: "create",
			},
			wantErr: true,
		}, {
			given: "Role is Ops & email = ops.one@company.com",
			when:  "trying to assign", then: "is not allowed",
			args: args{
				req: func() *http.Request {
					req, _ := http.NewRequest("POST", "http://api.example.com/inquiries/INQ-0001/assign", nil)
					ctx := context.WithValue(context.Background(), rbac.ContextKey("email"), "ops.one@company.com")

					return req.WithContext(ctx)
				},
				role:     "ops",
				resource: "inquiry",
				endpoint: "assign",
			},
			wantErr: true,
		},

		// As a manager
		{
			given: "Role is Manager",
			when:  "query is not given", then: "status=Assigned is enforced",
			args: args{
				req: func() *http.Request {
					req, _ := http.NewRequest("", "http://api.example.com/inquiries", nil)
					ctx := context.WithValue(context.Background(), rbac.ContextKey("email"), "manager@company.com")

					return req.WithContext(ctx)
				},
				role:     "manager",
				resource: "inquiry",
				endpoint: "get",
			},
			queryResult: map[string]string{
				"status": "Assigned",
			},
		}, {
			given: "Role is Manager",
			when:  "query ?status is given", then: "status=Assigned is still enforced",
			args: args{
				req: func() *http.Request {
					req, _ := http.NewRequest("", "http://api.example.com/inquiries?status=New", nil)
					ctx := context.WithValue(context.Background(), rbac.ContextKey("email"), "manager@company.com")

					return req.WithContext(ctx)
				},
				role:     "manager",
				resource: "inquiry",
				endpoint: "get",
			},
			queryResult: map[string]string{
				"status": "Assigned",
			},
		}, {
			given: "Role is Manager",
			when:  "trying to create", then: "is not allowed",
			args: args{
				req: func() *http.Request {
					req, _ := http.NewRequest("POST", "http://api.example.com/inquiries", nil)
					ctx := context.WithValue(context.Background(), rbac.ContextKey("email"), "manager@company.com")

					return req.WithContext(ctx)
				},
				role:     "manager",
				resource: "inquiry",
				endpoint: "create",
			},
			wantErr: true,
		}, {
			given: "Role is Manager",
			when:  "trying to assign", then: "is allowed",
			args: args{
				req: func() *http.Request {
					req, _ := http.NewRequest("POST", "http://api.example.com/inquiries/INQ-0001/assign", nil)
					ctx := context.WithValue(context.Background(), rbac.ContextKey("email"), "manager@company.com")

					return req.WithContext(ctx)
				},
				role:     "manager",
				resource: "inquiry",
				endpoint: "assign",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.given, func(t *testing.T) {
			req := tt.args.req()
			got := rbo.Authorize(req, tt.args.role, tt.args.resource, tt.args.endpoint)
			if tt.wantErr {
				assert.Error(t, got, "when: %s, then: %s", tt.when, tt.then)
			} else {
				assert.NoError(t, got, "when: %s, then: %s", tt.when, tt.then)
				for key, val := range tt.queryResult {
					assert.Equal(t, val, req.URL.Query().Get(key), "when: %s, then: %s", tt.when, tt.then)
				}
			}
		})
	}
}
