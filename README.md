# dumbo

Dumbo makes database user management easy.

Features:

* Create schemas
* Create users with read-write or read-only access to a schema
* Grant or revoke read-write or read-only access to a schema
* List managed schemas
* List users with managed roles

Currently only PostgreSQL is supported.

## Usage

The `DUMBO_CONN` environment variable must be set to a connection string for the database you want to manage. For example:

```
export DUMBO_CONN=postgres://username:password@host:port/database?sslmode=disable
```

Note that this exposes your password in plaintext as an environment variable.

```
dumbo create schema <schema>
dumbo create user <user> in <schema> [--password <password>] [--readonly]
dumbo list schema[s]
dumbo list users in <schema> [--boring]
dumbo grant access to <schema> to <user> [--readonly]
dumbo revoke access to <schema> from <user>
dumbo -h | --help
dumbo --version

Options:
  -h --help   Show help.
  --version   Show version.
  --password  User's login credential. Defaults to empty string.
  --readonly  Grant only read access to the schema.
  --boring    Print results as CSV intead of pretty printing.
```
