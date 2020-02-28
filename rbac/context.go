package rbac

// Collection of accepted RBAC context
const (
	ContextKeyRole  = ContextKey("role")
	ContextKeyEmail = ContextKey("email")
)

// ContextKey is typed alias to a string for use in golang context
type ContextKey string

func (c ContextKey) String() string {
	return "rbac_context_" + string(c)
}
