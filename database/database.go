package database

import (
	"database/sql"
	"embed"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"

	"github.com/orthanc/feedgenerator/feeddb"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

type Database struct {
	DB *sql.DB
	Queries *feeddb.Queries
}

func New() *Database {
	db, err := sql.Open("sqlite3", "feed-data.sqllite")
	if err != nil {
		panic(err)
	}

	queries := feeddb.New(db)

	database := Database{
		DB: db,
		Queries: queries,
	}
	return &database;
}

func (database *Database) Migrate() {
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		panic(err)
	}
	if err := goose.Up(database.DB, "migrations"); err != nil {
		panic(err)
	}
}