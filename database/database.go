package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"

	"github.com/orthanc/feedgenerator/database/read"
	"github.com/orthanc/feedgenerator/database/write"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

type Database struct {
	readDB  *sql.DB
	writeDB *sql.DB
	Queries *read.Queries
	Updates *write.Queries
}

func connect(ctx context.Context) (*sql.DB, error) {
	dbLocation := os.Getenv("FEEDGEN_SQLITE_LOCATION")
	db, err := sql.Open("sqlite3", fmt.Sprintf("%s?_txlock=immediate", dbLocation))
	if err != nil {
		return nil, fmt.Errorf("error creating db at %s: %s", dbLocation, err)
	}
	// https://kerkour.com/sqlite-for-servers
	db.ExecContext(ctx, "PRAGMA journal_mode = WAL")
	db.ExecContext(ctx, "PRAGMA busy_timeout = 5000")
	db.ExecContext(ctx, "PRAGMA synchronous = NORMAL")
	db.ExecContext(ctx, "PRAGMA cache_size = 1000000000")
	db.ExecContext(ctx, "PRAGMA foreign_keys = true")
	db.ExecContext(ctx, "PRAGMA temp_store = memory")
	return db, nil
}

func NewDatabase(ctx context.Context) (*Database, error) {
	writeDB, err := connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating write db:%s", err)
	}
	writeDB.SetMaxOpenConns(1)

	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, fmt.Errorf("error setting goose dialect: %s", err)
	}
	if err := goose.Up(writeDB, "migrations"); err != nil {
		return nil, fmt.Errorf("error updating schema: %s", err)
	}

	readDB, err := connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating read db: %s", err)
	}
	readDB.SetMaxOpenConns(max(4, runtime.NumCPU()))

	database := Database{
		readDB:  readDB,
		writeDB: writeDB,
		Queries: read.New(readDB),
		Updates: write.New(writeDB),
	}
	return &database, nil
}

func (database *Database) BeginTx(ctx context.Context) (*write.Queries, *sql.Tx, error) {
	tx, err := database.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error starting transaction: %s", err)
	}
	qtx := database.Updates.WithTx(tx)
	return qtx, tx, nil
}
