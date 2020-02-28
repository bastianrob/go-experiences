# RBAC in REST API

Role Based Access Control, or RBAC.
Might be one of the more interesting challenge I've faced. It is:

* Needed by business
* Complicated by nature
* But quite rewarding to solve while felt over-engineered at the same time

## Dahell is RBAC

Imagine a scenario where you have 4 roles inside your company, which have access to 1 type of resource.
Let's say:

* `Manager`
* `Ops`
* `CS`
* `Client`

These roles have different kind of interaction with one type of resource. Let's say it's an `Inquiry`, where:

* `Client` can submit `New` `Inquiry`
* `CS` views `Inquiry` from `Client` an then assign it to an `Ops`
* `Client` can only view his own `Inquiry`
* `CS` can only view `Inquiry` with status = `New`
* `Ops` can only view `Inquiry` assigned to them
* `Manager` can view all `Inquiry` which have been assigned to `Ops`

This is what `RBAC` is, where we try to `Control` `Access` to a resource/action `Based` on `Role`

Hence the name: `Role Based Access Control`

## Minimum Effort Implementation

The easiest way to implement this is to filter resource access through `Frontend`:

* As a `Client` I must **ONLY** be able to view my `Inquiry`
* As a `CS` I must **ONLY** be able to view `New` `Inquiry`
* As an `Ops` I must **ONLY** be able to view `Inquiry` assigned to me
* As a `Manager` I must **ONLY** be able to view `Inquiry` assigned to `Ops`

Neat right? Looks like some `JIRA` stories for `Frontend`.
BUT NOT SO FAST...

If there's no restriction enforced by `Backend`, then:

* Let's say `Ops A` knows how to construct a query, therefore
* All `Inquiry` data can be leaked just by querying `GET /inquiries`
* There could be sensitive / confidential data which must never be able to seen by other than specified role
* e.g: `Client's` profile, address, phone, etc...

This is why `RBAC` must be implemented by `Backend`

## REST API

In typical `REST API` we usually creates endpoint based on a `Resource` (hence the `R` from `REST`... J.K. It's not!)

Example:

```bash
'GET /inquiries?param1={v1}&param2={v2}...' # Get list of inquiries based on some query parameter
'POST /inquiries'                           # Create a new inquiry
'POST /inquiries/{id}/assign'               # Assign an inquiry to a someone
```

With `Inquiry` data structure:

```json
{
    "id": "INQ-0001",
    "created_by": "client@email.com",
    "status": "Assigned",
    "assignee": "ops.one@company.email"
}
```

And we want to enforce `RBAC` on these endpoints

## Mapping the rules

Now we have:

* Four roles
* Lots of rules
* Three endpoints, and
* `Inquiry` data structure

Next we'll have to map permission to a role based on guideline below:

> `what's my role? what resource am I trying to access? at what endpoint? is it allowed? if it is, what's the rule?`

### So, let's start from `Client's` rule

We'll use `YAML` to structure our rule

```yaml
client:
  inquiry:
    get:
      allow: true
      ensure:
        query:
          - key: created_by
            operator: "="
            value: "ctx.email"

    create:
      allow: true

    assign:
      allow: false
```

Now, we have a rule for `Client`, how do we read it?

```markdown
* If I am a `client`, want to access `inquiry`:
  * For endpoint `get`, I am allowed, only when `url query`
    * Contains a key named `create_by`
    * And the value must be `=` to value of `ctx.email`
  * For endpoint `create`, I am allowed, without restriction
  * For endpoint `assign`, I am not allowed to access
```

---

### Next, `CS'` rule

```yaml
cs:
  inquiry:
    get:
      allow: true
      enforce:
        query:
          - key: status
            value: "New"

    create:
      allow: false

    assign:
      allow: true
```

We read it as:

```markdown
* If I am a `cs`, want to access `inquiry`:
  * For endpoint `get`, I am allowed, but `url query`
    * Will be enforced with a key named `status`
    * And the value is `status=New`
  * For endpoint `create`, I am not allowed to access
  * For endpoint `assign`, I am allowed, without restriction
```

---

### Next, `OPS'` rule

```yaml
ops:
  inquiry:
    get:
      allow: true
      ensure:
        query:
          - key: assignee
            operator: "="
            value: "ctx.email"

    create:
      allow: false

    assign:
      allow: false
```

We read it as:

```markdown
* If I am an `ops`, want to access `inquiry`:
  * For endpoint `get`, I am allowed, only when `url query`
    * Contains a key name `assignee`
    * And the value must be `=` to value of `ctx.email`
  * For endpoint `create`, I am not allowed to access
  * For endpoint `assign`, I am not allowed to access
```

---

### Lastly, `Manager's` rule

```yaml
manager:
  inquiry:
    get:
      allow: true
      enforce:
        query:
          - key: status
            value: "Assigned"

    create:
      allow: false

    assign:
      allow: true
```

We read it as:

```markdown
* If I am a `manager`, want to access `inquiry`:
  * For endpoint `get`, I am allowed, but `url query`
    * Will be enforced with a key named `status`
    * And the value is `status=Assigned`
  * For endpoint `create`, I am not allowed to access
  * For endpoint `assign`, I am allowed, without restriction
```

> `ctx` is `context.Context` object in golang which we usually pass around from end to end

## RBAC engine

Now comes the part where we should code the RBAC engine to check those rules we have listed.

---

### Let's start from Rule

Rule is the smallest unit in our `YAML`
It is used to compare `actual` incoming request, VS `expectation` we have set in `Rule`

`YAML` example:

```yaml
key: status
operator: "="
value: "New"
---
key: created_by
operator: "="
value: "ctx.email"
```

<details>
  <summary>Click if you want to see some very unimportant and boring technical details</summary>

  Go code equivalent of a `Rule`:

  ```go
  // Rule of a permission
  type Rule struct {
      Key      string `yaml:"key"`
      Operator string `yaml:"operator"`
      Value    string `yaml:"value"`
  }
  ```

  From the `YAML` we have 2 types of `Rule`

  1. `Rule.Value` = `{a string}` e.g: `status=New`
  2. `Rule.Value` = `ctx.{field}` e.g: `created_by=ctx.email`

  For the first type, we take the rule value as is
  But for the 2nd type, we have to take the `expected` value from `context` object, splitted by `.`

  So we have to prepare a `method`, owned by `Rule`, to get `expected` value, from `context`.
  Let's name it `FromContext`

  ```go
  // FromContext get actual rule.Value from ctx if rule.Value starts with ctx
  // otherwise, return rule.Value as is
  func (rule Rule) FromContext(ctx context.Context) interface{} {
      ...
  }
  ```

  Next, we write `unit tests` scenario to ensure `FromContext` behave as we wanted it to be:

  ```go
  // Semi BDD style unit testing
  // I think the code is self explanatory, we just wanted to:
  // call rule.FromContext(ctx), and
  //   want it to either panics, or
  //   produce a correct result
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
  ```

  Next, we actually write the code to satisfy these test scenario

  ```go
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
  ```

  Running the test (from VS Code)

  ```bash
  Running tool: /usr/local/opt/go/libexec/bin/go test -timeout 30s github.com/bastianrob/go-experiences/rbac -run ^(TestRule_FromContext)$

  ok      github.com/bastianrob/go-experiences/rbac
  Success: Tests passed.
  ```

  Now imagine we have to actually check whether a `Request` complies with given `Rule` or not
  Let's make another `method` for `Rule` and name it `Comply`

  ```go
  // Comply checks does request value complies with our rule
  func (rule Rule) Comply(expected, actual interface{}) bool {
      ...
  }
  ```

  We write another `unit tests` scenario to ensure `Comply` behave as we wanted it to be:

  ```go
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
  ```

  Again, we actually write the code to satisfy these test scenario

  ```go
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
  ```

  Running the test (from VS Code)

  ```bash
  Running tool: /usr/local/opt/go/libexec/bin/go test -timeout 30s github.com/bastianrob/go-experiences/rbac -run ^(TestRule_Comply)$

  ok      github.com/bastianrob/go-experiences/rbac
  Success: Tests passed.
  ```

</details>

---

### Ensurer and Enforcer

Next, we have `Ensurer` and `Enforcer`. As the name implies:

* `Ensurer` ensure that request must complies with all the rules attached in `Ensurer`
* `Enforcer` enforce that no matter what you request, we'll always enforce all the rules attached in `Enforcer`
* Both can be targeted to either `query`, `header`, or `path`
* And both contains list of `Rules` we have defined in the previous point

<details>
<summary>Don't Click! to avoid another boring technical detail</summary>

```go
// Ensurer data model
// Can either ensure query, header, or path
type Ensurer struct {
    Query  []Rule `yaml:"query"`
    Header []Rule `yaml:"header"`
    Path   []Rule `yaml:"path"`
}


// Enforcer structure is just like an Ensurer
type Enforcer Ensurer
```

For this example, we'll only implement `Query` `Ensurer` and `Enforcer`

```go
// QueryComplies check whether query request complies with rules
func (ens Ensurer) QueryComplies(r *http.Request) error {
    ...
}

// QueryComplies enforce query request from rule
func (enf Enforcer) QueryComplies(r *http.Request) error {
    ...
}
```

Then, we write `unit tests` to ensure `QueryComplies` behave as we wanted it to be:

```go
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
            // query: {id: "0001", name: "John"}
            url: "http://api.example.com/resources?id=0001&name=John",
        },
        ensurer: rbac.Ensurer{
            Query: []rbac.Rule{
                // id IS 0001, and name EQUALS to value stored in context.name
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
```

Then, we write actual `QueryComplies` code which satisfy our `unit tests`:

```go
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
```

And the test result is:

```bash
Running tool: /usr/local/opt/go/libexec/bin/go test -timeout 30s github.com/bastianrob/go-experiences/rbac -run ^(TestEnsurer_QueryComplies)$

ok      github.com/bastianrob/go-experiences/rbac   0.014s
Success: Tests passed.

---

Running tool: /usr/local/opt/go/libexec/bin/go test -timeout 30s github.com/bastianrob/go-experiences/rbac -run ^(TestEnforcer_QueryComplies)$

ok      github.com/bastianrob/go-experiences/rbac    0.014s
Success: Tests passed.
```

</details>

---

### Stitching it all up

Now we have:

* `Ensurer` and `Enforcer`
* Each have their own set of `Rules`
* Lastly we need to stitch them all together in an `RBAC` object

```go
// Error collection
var (
    ErrNotString   = errors.New("Expected value is not a string")
    ErrNoRole      = errors.New("You have no role assigned to you")
    ErrRoleUnknown = errors.New("You have an unknown role assigned to you")
    ErrForbidden   = errors.New("You are not allowed to access specified resource")
)


// Permission of an endpoint
type Permission struct {
    Allow   bool     `yaml:"allow"`
    Ensure  Ensurer  `yaml:"ensure,omitempty"`
    Enforce Enforcer `yaml:"enforce,omitempty"`
}

// Endpoint is a map of {endpoint: permission}
type Endpoint map[string]Permission

// Resource is a map of {resource: endpoint}
type Resource map[string]Endpoint

// RBAC is a map of {role: resource}
type RBAC map[string]Resource

// FromFile creates a new RBAC object from .yaml file
func FromFile(path string) *RBAC {
    f, err := ioutil.ReadFile(path)
    if err != nil {
        return nil
    }

    rbac := &RBAC{}
    err = yaml.Unmarshal(f, rbac)
    if err != nil {
        return nil
    }

    return rbac
}
```

Now we have an actual object called `RBAC`, we can parse `YAML` rule into `RBAC` using `FromFile` factory function.

Next, we have to write `Authorize` method for `RBAC`

```go
// Authorize a request based on its role, resource, and endpoint
func (rbac RBAC) Authorize(r *http.Request, role, resource, endpoint string) error {
    ...
}
```

And we write a `unit tests` for `Authorize`:

```go
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
```

Last, we write `Authorize` code to satisf our `unit tests`

```go
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
```

And the test results shows:

```bash
Running tool: /usr/local/opt/go/libexec/bin/go test -timeout 30s github.com/bastianrob/go-experiences/rbac -run ^(TestRBAC_Authorize)$

ok      github.com/bastianrob/go-experiences/rbac    (cached)
Success: Tests passed.
```

## So, Again What the Hell is `ctx`

We see a lot of `ctx` being thrown around so, let me explain it in chronologically, from end-to-end

```markdown
* Think of `ctx` as browser `cookies`
* As a user logged in our application, we respond to them by setting `cookie` into the browser
* This `cookie` is typically a string of encrypted data about logged in user's profile (think `JWT`)
* This `cookie` will always get sent to server for each and every request to our domain
* At the start of the request, our code have to:
  * Check existence of `cookie`
    * Returns 401 if not exists (unauthorized)
    * Proceed if exists
* Unwrap the `cookie` value, which in our case typpically contains `role` and `email` of the logged in user
* And then, we set all of the unwrapped `cookie` value into HTTP request `context`
* In this case we set `ctx.name`
* Then, we call the `RBAC` engine to `Authorize` the request
  * `RBAC` ensures the request satisfy all rules inside the `Ensurer`
  * `RBAC` enforce all rules inside the `Enforcer` by rewriting the HTTP request object
* If HTTP request pass all the rules in `RBAC` engine, we then proceed to pass the HTTP request to `Business Service Layer`
* Because `RBAC` engine relies heavily on the HTTP query, `Business Service Layer` must be written to always converts the HTTP query into actual `Database Query`. e.g:
  * `/inquiries?created_by=client@email.com` must be translated to `SELECT * FROM inquiries WHERE created_by = @url.query.created_by`
  * `/inquiries?status=New` must be translated to `SELECT * FROM inquiries WHERE status = @url.query.status`
```

## What does this all means? This is nuts

Let us take a journey and assume position as each of the roles

### As a `Client`

```yaml
client:
  inquiry:
    get:
      allow: true
      ensure:
        query:
          - key: created_by
            operator: "="
            value: "ctx.email"

    create:
      allow: true

    assign:
      allow: false
```

```markdown
* I am currently logged in as `client.one@email.com`
* I am trying to access `GET /inquiries`
  * `RBAC` states: `ensure query: created_by = ctx.email`
  * `GET /inquiries` doesn't have any query, therefore rule is not satisfied
  * Returned as `Unauthorized`
* I am trying to access `GET /inquiries?created_by=client.other@email.com`
  * `RBAC` states: `ensure query: created_by = ctx.email`
  * `GET /inquiries?created_by=client.other@email.com` have the `created_by` query and will be compared against `ctx.email`
  * `ctx.email` is `client.one@gmail.com`, but supplied query is `client.other@email.com`, therefore rule is not satisfied
  * Returned as `Unauthorized`
```

So now as a `Client` even though I can read and construct a query, I still can't get any data from other people

---

### As a `CS`

```yaml
cs:
  inquiry:
    get:
      allow: true
      enforce:
        query:
          - key: status
            value: "New"

    create:
      allow: false

    assign:
      allow: true
```

```markdown
* I am currently logged in as `cs.one@company.com`
* I am trying to access `GET /inquiries`
  * `RBAC` states: `enforce query: status = New`
  * `GET /inquiries` doesn't have any query, but `Enforcer` will forcefully write it as `?status=New`
* I am trying to access `GET /inquiries?status=Assigned`
  * `RBAC` states: `enforce query: status = New`
  * `GET /inquiries?status=Assigned` have the `status` query and the value is `Assigned`
  * `Enforcer` doesn't care with the value requested, and will will forcefully re-write it as `?status=New`
```

So now as a `CS` even though I can read and construct a query, I can only get `inquiries` with `status=New`

---

### As an `Ops`

```yaml
ops:
  inquiry:
    get:
      allow: true
      ensure:
        query:
          - key: assignee
            operator: "="
            value: "ctx.email"

    create:
      allow: false

    assign:
      allow: false
```

```markdown
* I am currently logged in as `ops.one@company.com`
* I am trying to access `GET /inquiries`
  * `RBAC` states: `ensure query: assignee = ctx.email`
  * `GET /inquiries` doesn't have any query, therefore rule is not satisfied
  * Returned as `Unauthorized`
* I am trying to access `GET /inquiries?assignee=ops.other@company.com`
  * `RBAC` states: `ensure query: status = ctx.email`
  * `GET /inquiries?status=ops.other@company.com` have the `created_by` query and will be compared against `ctx.email`
  * `ctx.email` is `ops.one@company.com`, but supplied query is `ops.other@company.com`, therefore rule is not satisfied
  * Returned as `Unauthorized`
```

So now as an `Ops` even though I can read and construct a query, I still can't get any data from any other `Ops`

### As a `Manager`

```yaml
manager:
  inquiry:
    get:
      allow: true
      enforce:
        query:
          - key: status
            value: "Assigned"

    create:
      allow: false

    assign:
      allow: true
```

```markdown
* I am currently logged in as `manager@company.com`
* I am trying to access `GET /inquiries`
  * `RBAC` states: `enforce query: status = Assigned`
  * `GET /inquiries` doesn't have any query, but `Enforcer` will forcefully write it as `?status=Assigned`
* I am trying to access `GET /inquiries?status=New`
  * `RBAC` states: `enforce query: status = Assigned`
  * `GET /inquiries?status=New` have the `status` query and the value is `New`
  * `Enforcer` doesn't care with the value requested, and will will forcefully re-write it as `?status=Assigned`
```

So now as a `Manager` even though I can read and construct a query, I can only get `inquiries` with `status=Assigned`
