// Package dumbo provides database user management helpers to support the dumbo utility.
package dumbo

import "context"

// A Manager handles managed database schemas, roles, and users.
type Manager interface {
	// CreateSchema with default access roles.
	// Both read-write and read-only roles will be created.
	// Roles are namespaced by the schema name.
	CreateSchema(ctx context.Context, name string) (err error)

	// CreateUser with an access role.
	// Access can be read-write or read-only.
	CreateUser(ctx context.Context, schema, name, password string, readOnly bool) (err error)

	// ListSchemas which are probably managed.
	// Only schemas which match namespaced roles are returned.
	ListSchemas(ctx context.Context) (schemas []string, err error)

	// ListUsers with access to a schema who are probably managed.
	// Only users with namespaced roles are returned.
	ListUsers(ctx context.Context, schema string) (users []User, err error)

	// GrantAccess to a schema to a user.
	GrantAccess(ctx context.Context, schema, user string, readOnly bool) (err error)

	// RevokeAccess to a schema from a user.
	RevokeAccess(ctx context.Context, schema, user string) (revokedRW, revokedRO bool, err error)
}

// A User managed by dumbo.
type User struct {
	Name string
	Role string
}
