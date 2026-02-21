package sqldb

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	turso "turso.tech/database/tursogo"
)

type SQLiteConfig struct {
	LocalPath       string `validate:"required"`
	RemoteUrl       string `validate:"required"`
	AuthToken       string `validate:"required" log:"-"`
	DebugMigrations bool
	Debug           bool
}

func NewTursoConnection(ctx context.Context, conf SQLiteConfig) (*gorm.DB, error) {

	var conn *sql.DB
	var err error
	if conf.RemoteUrl != "" {
		// Connect a local database to a remote Turso database
		dbSync, err := turso.NewTursoSyncDb(ctx, turso.TursoSyncDbConfig{
			Path:      conf.LocalPath,
			RemoteUrl: conf.RemoteUrl,
			AuthToken: conf.AuthToken,
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		conn, err = dbSync.Connect(ctx)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		conn, err = sql.Open("turso", conf.LocalPath)
		if err != nil {
			log.Fatal(err)
		}
	}
	defer conn.Close()

	db, err := gorm.Open(sqlite.New(sqlite.Config{
		Conn: conn,
	}), &gorm.Config{})

	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	return db, nil
}

type SQLiteDB struct {
	db     *gorm.DB
	config SQLiteConfig
}

func (d *SQLiteDB) DB(ctx context.Context) *gorm.DB {
	if d.config.Debug {
		return d.db.Debug()
	}
	return d.db
}

func (db *SQLiteDB) Migrate(ctx context.Context, models []any) {
	curentLogger := db.db.Logger
	if !db.config.DebugMigrations {
		db.db.Logger = logger.Discard
	}
	if err := db.DB(ctx).AutoMigrate(models...); err != nil {
		log.Fatal(ctx, "could not migrate", err)
	}
	db.db.Logger = curentLogger
}

func (db *SQLiteDB) TableName(m any) string {
	stmt := &gorm.Statement{DB: db.db}
	stmt.Parse(m)
	return stmt.Schema.Table
}
