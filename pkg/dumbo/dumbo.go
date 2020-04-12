/*
MIT License

Copyright (c) 2020 Kai Wells

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

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
