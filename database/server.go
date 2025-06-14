package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	Connection DatabaseConnection
	Database   *gorm.DB
}

type DatabaseConnection struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

func (c *DatabaseConnection) String() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", c.Host, c.Port, c.User, c.Password, c.Database)
}

func NewConnection(host string, port int, user string, password string, database string) *Database {
	connection := DatabaseConnection{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
	}
	db, err := gorm.Open(postgres.Open(connection.String()), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	return &Database{
		Connection: connection,
		Database:   db,
	}
}
