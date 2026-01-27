package postgres

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Options struct {
	DBName   string
	DBUser   string
	Password string
	Host     string
	Port     string
	SSLMode  bool
}

func NewConnection(opts Options) (*gorm.DB, error) {
	sslmode := "disable"
	if opts.SSLMode {
		sslmode = "enable"
	}

	datasource := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		opts.Host, opts.Port, opts.DBUser, opts.Password, opts.DBName, sslmode,
	)

	return gorm.Open(postgres.Open(datasource), &gorm.Config{})
}
