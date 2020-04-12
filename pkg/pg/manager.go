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

// Package pg is the PostgreSQL implementation for dumbo.
package pg

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // PostgreSQL magic
	"github.com/quells/dumbo/pkg/dumbo"
)

type manager struct {
	db *sql.DB
}

// Connect to a PostgreSQL server and configure a Manager.
func Connect(ctx context.Context, conn string) (m dumbo.Manager, err error) {
	var db *sql.DB
	db, err = sql.Open("postgres", conn)
	if err != nil {
		return
	}

	if err = db.PingContext(ctx); err != nil {
		return
	}

	m = &manager{db}
	return
}

const (
	createSchema = "CREATE SCHEMA IF NOT EXISTS "
	checkRole    = "SELECT FROM pg_catalog.pg_roles WHERE rolname=$1"
	createRole   = "CREATE ROLE "
	getDBName    = "SELECT current_database()"
	grantConnect = "GRANT CONNECT ON DATABASE %s TO %s"

	grantUsageRW      = "GRANT USAGE, CREATE ON SCHEMA %s TO %s"
	grantTablesRW     = "GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA %s TO %s"
	alterDefaultRW    = "ALTER DEFAULT PRIVILEGES IN SCHEMA %s GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO %s"
	grantSeqRW        = "GRANT USAGE ON ALL SEQUENCES IN SCHEMA %s TO %s"
	alterDefaultSeqRW = "ALTER DEFAULT PRIVILEGES IN SCHEMA %s GRANT USAGE ON SEQUENCES TO %s"

	grantUsageRO   = "GRANT USAGE ON SCHEMA %s TO %s"
	grantTablesRO  = "GRANT SELECT ON ALL TABLES IN SCHEMA %s TO %s"
	alterDefaultRO = "ALTER DEFAULT PRIVILEGES IN SCHEMA %s GRANT SELECT ON TABLES TO %s"
)

func (m *manager) CreateSchema(ctx context.Context, name string) (err error) {
	var tx *sql.Tx
	tx, err = m.db.BeginTx(ctx, nil)
	if err != nil {
		tx.Rollback()
		return
	}

	row := tx.QueryRow(getDBName)
	if row == nil {
		tx.Rollback()
		return
	}
	var dbName string
	if err = row.Scan(&dbName); err != nil {
		tx.Rollback()
		return
	}

	if _, err = tx.Exec(createSchema + name); err != nil {
		tx.Rollback()
		return
	}

	createRoleIfNotExists := func(roleName string) (e error) {
		var rows *sql.Rows
		rows, e = tx.Query(checkRole, roleName)
		if e != nil {
			return
		}

		// continue if role already exists
		defer rows.Close()
		if rows.Next() {
			return
		}

		_, e = tx.Exec(createRole + roleName)
		return
	}

	rwName := name + "_rw"
	if err = createRoleIfNotExists(rwName); err != nil {
		tx.Rollback()
		return
	}
	if _, err = tx.Exec(fmt.Sprintf(grantConnect, dbName, rwName)); err != nil {
		tx.Rollback()
		return
	}
	for _, q := range []string{grantUsageRW, grantTablesRW, alterDefaultRW, grantSeqRW, alterDefaultSeqRW} {
		if _, err = tx.Exec(fmt.Sprintf(q, name, rwName)); err != nil {
			tx.Rollback()
			return
		}
	}

	roName := name + "_ro"
	createRoleIfNotExists(roName)
	if _, err = tx.Exec(fmt.Sprintf(grantConnect, dbName, roName)); err != nil {
		tx.Rollback()
		return
	}
	for _, q := range []string{grantUsageRO, grantTablesRO, alterDefaultRO} {
		if _, err = tx.Exec(fmt.Sprintf(q, name, roName)); err != nil {
			tx.Rollback()
			return
		}
	}

	err = tx.Commit()
	return
}

const (
	createUser = "CREATE USER %s WITH PASSWORD '%s'"
	grantRole  = "GRANT %s TO %s"
)

func (m *manager) CreateUser(ctx context.Context, schema, name, password string, readOnly bool) (err error) {
	var tx *sql.Tx
	tx, err = m.db.BeginTx(ctx, nil)
	if err != nil {
		tx.Rollback()
		return
	}

	// check if role with that name already exists
	var rows *sql.Rows
	rows, err = tx.Query(checkRole, name)
	if err != nil {
		tx.Rollback()
		return
	}
	var alreadyExists bool
	if rows.Next() {
		alreadyExists = true
	}
	rows.Close()

	if !alreadyExists {
		if _, err = tx.Exec(fmt.Sprintf(createUser, name, password)); err != nil {
			tx.Rollback()
			return
		}
	}

	var roleName string
	if readOnly {
		roleName = schema + "_ro"
	} else {
		roleName = schema + "_rw"
	}
	if _, err = tx.Exec(fmt.Sprintf(grantRole, roleName, name)); err != nil {
		tx.Rollback()
		return
	}

	err = tx.Commit()
	return
}

const getManagedSchemas = `WITH
	schemas AS (
		SELECT
			schema_name AS name
		FROM information_schema.schemata
	),
	managed AS (
		SELECT
			DISTINCT left(role_name, -3) AS schema_name
		FROM information_schema.enabled_roles
		WHERE
			role_name LIKE '%_rw' OR
			role_name LIKE '%_ro'
	)
SELECT
	name
FROM schemas JOIN managed ON schemas.name=managed.schema_name
ORDER BY name ASC`

func (m *manager) ListSchemas(ctx context.Context) (schemas []string, err error) {
	var rows *sql.Rows
	rows, err = m.db.QueryContext(ctx, getManagedSchemas)
	if err != nil {
		return
	}

	var name string
	for rows.Next() {
		if err = rows.Scan(&name); err != nil {
			return
		}
		schemas = append(schemas, name)
	}

	return
}

const getManagedUsers = `SELECT
	grantee, role_name
FROM information_schema.applicable_roles
WHERE
	role_name=$1 OR
	role_name=$2
ORDER BY grantee ASC`

func (m *manager) ListUsers(ctx context.Context, schema string) (users []dumbo.User, err error) {
	rwName := schema + "_rw"
	roName := schema + "_ro"
	var rows *sql.Rows
	rows, err = m.db.QueryContext(ctx, getManagedUsers, rwName, roName)
	if err != nil {
		return
	}

	var grantee, roleName string
	for rows.Next() {
		if err = rows.Scan(&grantee, &roleName); err != nil {
			return
		}
		user := dumbo.User{
			Name: grantee,
			Role: roleName,
		}
		users = append(users, user)
	}

	return
}

func (m *manager) GrantAccess(ctx context.Context, schema, user string, readOnly bool) (err error) {
	var roleName string
	if readOnly {
		roleName = schema + "_ro"
	} else {
		roleName = schema + "_rw"
	}

	_, err = m.db.ExecContext(ctx, fmt.Sprintf(grantRole, roleName, user))
	return
}

const revokeRole = "REVOKE %s FROM %s"

func (m *manager) RevokeAccess(ctx context.Context, schema, user string) (revokedRW, revokedRO bool, err error) {
	rwName := schema + "_rw"
	roName := schema + "_ro"

	_, err = m.db.ExecContext(ctx, fmt.Sprintf(revokeRole, rwName, user))
	if err == nil {
		revokedRW = true
	}

	_, err = m.db.ExecContext(ctx, fmt.Sprintf(revokeRole, roName, user))
	if err == nil {
		revokedRO = true
	}
	return
}
