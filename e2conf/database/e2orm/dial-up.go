package e2orm

import (
	"strings"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// NewPostgreSQL create new postgresql connect
// Endpoint format:
// host=pg-host port=5432 user=db-user password=db-password dbname=database-name sslmode=require application_name=application-name
func NewPostgreSQL(c *Config) *Connect {
	if !strings.Contains(strings.ToLower(c.Endpoint), "application_name=") {
		c.Endpoint += " application_name=golang-e2util "
	}
	for idx := range c.RoEndpoint {
		rc := c.RoEndpoint[idx]
		if !strings.Contains(strings.ToLower(rc), "application_name=") {
			c.RoEndpoint[idx] += " application_name=golang-e2util "
		}
	}
	return newConnect("postgres", c)
}

// NewMySQL create new postgresql connect
// Endpoint format:
// db-user:db-password@tcp(db-host:db-port)/database-name
func NewMySQL(c *Config) *Connect {
	return newConnect("mysql", c)
}
