package auth

import (
	"github.com/secnex/auth/database"
	"github.com/secnex/auth/server"
)

type AuthProvider struct {
	Database *database.Database
}

func NewAuthProvider(database *database.Database) *AuthProvider {
	return &AuthProvider{Database: database}
}

func (p *AuthProvider) RunServer() {
	s := server.NewApiServer(nil, nil, "localhost", 5432, "postgres", "postgres", "auth")
	s.Start()
}
