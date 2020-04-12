/*

Dumbo makes database user management easy.

Features:

	* Create schemas
	* Create users with read-write or read-only access to a schema
	* Grant or revoke read-write or read-only access to a schema
	* List managed schemas
	* List users with managed roles

Currently only PostgreSQL is supported.

Usage:

	dumbo create schema <schema>
	dumbo create user <user> in <schema> [--password <password>] [--readonly]
	dumbo list schema[s]
	dumbo list users in <schema> [--boring]
	dumbo grant access to <schema> to <user> [--readonly]
	dumbo revoke access to <schema> from <user>
	dumbo -h | --help
	dumbo --version

Options:

	-h --help
		show help
	--version
		show version
	--password
		user's login credential; defaults to empty string
	--readonly
		grant only read access to the schema
	--boring
		print results as CSV intead of pretty printing

dumbo looks for the database connection string in $DUMBO_CONN.

*/
package main

const usage = `Database User Management, Boring Old.

Usage:
  dumbo create schema <schema>
  dumbo create user <user> in <schema> [--password <password>] [--readonly]
  dumbo list schema
  dumbo list schemas
  dumbo list users in <schema> [--boring]
  dumbo grant access to <schema> to <user> [--readonly]
  dumbo revoke access to <schema> from <user>
  dumbo -h | --help
  dumbo --version

Options:
  -h --help   Show this screen.
  --version   Show version.
  --password  User's login credential. Defaults to empty string.
  --readonly  Grant only read access to the schema.
  --boring    Print results as CSV intead of pretty printing.`
