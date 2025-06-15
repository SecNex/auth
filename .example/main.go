package main

import (
	"github.com/secnex/auth"
	"github.com/secnex/auth/database"
)

func main() {
	p := auth.NewAuthProvider(database.NewConnection("localhost", 5432, "postgres", "postgres", "auth"))
	p.RunServer()
}
