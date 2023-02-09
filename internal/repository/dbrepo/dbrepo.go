package dbrepo

import (
	"database/sql"

	"github.com/atom91/bookings/internal/config"
	"github.com/atom91/bookings/internal/repository"
)

type postgresDbRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}



func NewPostgresRepo(con *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	return &postgresDbRepo{
		App: a,
		DB:  con,
	}
}
