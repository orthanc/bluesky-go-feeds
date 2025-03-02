package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"runtime"
	"slices"

	"github.com/mattn/go-sqlite3"
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

type median struct {
	values  []int64
}

func newMedian() *median {
	return &median{}
}

func (s *median) Step(x int64) {
	s.values = append(s.values, x)
}

func (s *median) Done() int64 {
	if len(s.values) == 0 {
		return 0
	}
	slices.Sort(s.values)
	return s.values[len(s.values)/2]
}

func init() {
	sql.Register("sqlite3_custom", &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			if err := conn.RegisterAggregator("median", newMedian, true); err != nil {
				return err
			}
			return nil
		},
	})
}


func connect(ctx context.Context, extraParams string) (*sql.DB, error) {
	dbLocation := os.Getenv("FEEDGEN_SQLITE_LOCATION")
	// https://kerkour.com/sqlite-for-servers
	db, err := sql.Open("sqlite3_custom", fmt.Sprintf("%s?_txlock=immediate&_journal=WAL&_timeout=5000&_sync=1&_cache_size=25000&_fk=true%s", dbLocation, extraParams))
	if err != nil {
		return nil, fmt.Errorf("error creating db at %s: %s", dbLocation, err)
	}
	// https://kerkour.com/sqlite-for-servers
	db.ExecContext(ctx, "PRAGMA temp_store = memory")
	return db, nil
}

func NewDatabase(ctx context.Context) (*Database, error) {
	writeDB, err := connect(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("error creating write db:%s", err)
	}
	writeDB.SetMaxOpenConns(1)

	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, fmt.Errorf("error setting goose dialect: %s", err)
	}

	readDB, err := connect(ctx, "&mode=ro")
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

func (database *Database) MigrateUp(_ context.Context) error {
	if err := goose.Up(database.writeDB, "migrations"); err != nil {
		return fmt.Errorf("error updating schema: %s", err)
	}
	return nil
}

func (database *Database) MigrateDown(_ context.Context) error {
	if err := goose.Down(database.writeDB, "migrations"); err != nil {
		return fmt.Errorf("error updating schema: %s", err)
	}
	return nil
}

func (database *Database) BeginTx(ctx context.Context) (*write.Queries, *sql.Tx, error) {
	tx, err := database.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error starting transaction: %s", err)
	}
	qtx := database.Updates.WithTx(tx)
	return qtx, tx, nil
}

func (database *Database) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return database.readDB.QueryContext(ctx, query, args...)
}

func (database *Database) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return database.readDB.QueryRowContext(ctx, query, args...)
}

func (database *Database) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return database.writeDB.ExecContext(ctx, query, args...)
}
