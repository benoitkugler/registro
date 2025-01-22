package config

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
)

// DB provides access to a database
type DB struct {
	Host     string
	User     string
	Password string
	Name     string // of the database
	Port     int    // default to 5432
}

// NewDB uses env variables to build DB credentials :
// DB_HOST, DB_USER, DB_PASSWORD, DB_NAME, DB_PORT (optional)
func NewDB() (out DB, err error) {
	out.Host = os.Getenv("DB_HOST")
	if out.Host == "" {
		return DB{}, errors.New("missing env DB_HOST")
	}

	out.User = os.Getenv("DB_USER")
	if out.User == "" {
		return DB{}, errors.New("missing env DB_USER")
	}

	out.Password = os.Getenv("DB_PASSWORD")
	if out.Password == "" {
		return DB{}, errors.New("missing env DB_PASSWORD")
	}

	out.Name = os.Getenv("DB_NAME")
	if out.Name == "" {
		return DB{}, errors.New("missing env DB_NAME")
	}

	out.Port = 5432
	if port := os.Getenv("DB_PORT"); port != "" {
		out.Port, err = strconv.Atoi(port)
		if err != nil {
			return DB{}, fmt.Errorf("invalid DB_PORT: %s", err)
		}
	}

	return out, nil
}

// ConnectPostgres builds a connection string and
// connect using postgres as driver name.
func (db DB) ConnectPostgres() (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		db.Host, db.Port, db.User, db.Password, db.Name)
	return sql.Open("postgres", connStr)
}
