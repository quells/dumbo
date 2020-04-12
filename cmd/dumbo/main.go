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

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/docopt/docopt-go"
	"github.com/quells/dumbo/pkg/dumbo"
	"github.com/quells/dumbo/pkg/pg"
)

const (
	timeout = 30 * time.Second
	connEnv = "DUMBO_CONN"
)

const (
	success int = iota
	showUsage
	missingConnString
	couldNotConnect
	couldNotCreateSchema
	couldNotCreateUser
	couldNotListSchema
	couldNotListUsers
	couldNotGrantAccess
	couldNotRevokeAccess
)

func main() {
	args, _ := docopt.ParseArgs(usage, os.Args[1:], version)

	conn := os.Getenv(connEnv)
	if conn == "" {
		fmt.Fprintf(os.Stderr, "%s must be set\n", connEnv)
		os.Exit(missingConnString)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var err error
	var manager dumbo.Manager
	manager, err = pg.Connect(ctx, conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not connect to database: %v\n", err)
		os.Exit(couldNotConnect)
	}

	schema, _ := args.String("<schema>")
	username, _ := args.String("<user>")
	password, _ := args.String("<password>")
	readonly, _ := args.Bool("--readonly")
	asCsv, _ := args.Bool("--boring")

	switch {
	case matchCmd(args, "create", "schema"):
		if err = manager.CreateSchema(ctx, schema); err != nil {
			fmt.Fprintf(os.Stderr, "could not create schema: %v\n", err)
			os.Exit(couldNotCreateSchema)
		}
	case matchCmd(args, "create", "user"):
		if err = manager.CreateUser(ctx, schema, username, password, readonly); err != nil {
			fmt.Fprintf(os.Stderr, "could not create user: %v\n", err)
			os.Exit(couldNotCreateUser)
		}
	case matchCmd(args, "list", "schema") || matchCmd(args, "list", "schemas"):
		var schemas []string
		schemas, err = manager.ListSchemas(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not get schemas: %v\n", err)
			os.Exit(couldNotListSchema)
		}
		for _, s := range schemas {
			fmt.Println(s)
		}
	case matchCmd(args, "list", "users"):
		var users []dumbo.User
		users, err = manager.ListUsers(ctx, schema)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not get users: %v\n", err)
			os.Exit(couldNotListUsers)
		}
		if asCsv {
			printUsers(users)
		} else {
			prettyPrintUsers(users)
		}
	case matchCmd(args, "grant", "access"):
		if err = manager.GrantAccess(ctx, schema, username, readonly); err != nil {
			fmt.Fprintf(os.Stderr, "could not grant access to user: %v\n", err)
			os.Exit(couldNotGrantAccess)
		}
	case matchCmd(args, "revoke", "access"):
		if _, _, err = manager.RevokeAccess(ctx, schema, username); err != nil {
			fmt.Fprintf(os.Stderr, "could not revoke access: %v\n", err)
			os.Exit(couldNotRevokeAccess)
		}
	default:
		fmt.Println(args)
		fmt.Println(usage)
		os.Exit(1)
	}
}

func matchCmd(args docopt.Opts, keywords ...string) bool {
	if len(keywords) == 0 {
		return false
	}

	for _, k := range keywords {
		if !args[k].(bool) {
			return false
		}
	}

	return true
}

func printUsers(users []dumbo.User) {
	for _, u := range users {
		fmt.Printf("%s,%s\n", u.Name, u.Role)
	}
}

func prettyPrintUsers(users []dumbo.User) {
	var nw, rw int
	for _, u := range users {
		n := len(u.Name)
		r := len(u.Role)
		if n > nw {
			nw = n
		}
		if r > rw {
			rw = r
		}
	}

	hr := "+" + strings.Repeat("-", nw+rw+5) + "+"
	f := "| %" + fmt.Sprintf("%ds", nw) + " | %" + fmt.Sprintf("%ds", rw) + " |\n"
	fmt.Println(hr)
	for _, u := range users {
		fmt.Printf(f, u.Name, u.Role)
	}
	fmt.Println(hr)
}
