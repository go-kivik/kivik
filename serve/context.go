package serve

import "golang.org/x/net/context"

type contextKey struct {
	name string
}

var (
	// AuthUserKey is a context key used to access the authenticated user, ifany.
	AuthUserKey = &contextKey{"auth-user"}
	// AuthRolesKey is a context key used to access the authenticated user's roles.
	AuthRolesKey = &contextKey{"auth-roles"}
	// ClientContextKey is a context key used to access the kivik client.
	ClientContextKey = &contextKey{"kivik-client"}
	// ServiceContextKey is a context key used to access the serve.Service struct.
	ServiceContextKey = &contextKey{"kivik-service"}
)

// GetUser gets the authenticated user from the context.
func GetUser(ctx context.Context) (string, bool) {
	u, ok := ctx.Value(AuthUserKey).(string)
	return u, ok
}

// GetRoles gets the authenticated user's roles from the context.
func GetRoles(ctx context.Context) ([]string, bool) {
	r, ok := ctx.Value(AuthRolesKey).([]string)
	return r, ok
}
