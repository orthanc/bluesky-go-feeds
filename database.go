package main

import (
	"database/sql"
	"embed"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"

	"github.com/orthanc/feedgenerator/feeddb"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func createDb() *feeddb.Queries {
	db, err := sql.Open("sqlite3", "feed-data.sqllite")
	if err != nil {
		panic(err)
	}

	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		panic(err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		panic(err)
	}

	queries := feeddb.New(db)

	return queries
}
