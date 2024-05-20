package pgdb

import (
	"context"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Name            string
	Host            string
	Port            int
	User            string `log:"-"`
	Password        string `log:"-"`
	SSLRootCertPath string
	DebugMigrations bool
	Debug           bool
}

type PGDB struct {
	db     *gorm.DB
	config Config
}

func (d *PGDB) DB(ctx context.Context) *gorm.DB {
	if d.config.Debug {
		return d.db.Debug()
	}
	return d.db
}

func NewPGConnection(ctx context.Context, config Config) *PGDB {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		config.Host, config.Port, config.User, config.Password, config.Name)
	if config.SSLRootCertPath != "" {
		dsn = fmt.Sprintf("%s sslrootcert=%s", dsn, config.SSLRootCertPath)
	} else {
		dsn = fmt.Sprintf("%s sslmode=disable", dsn)
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect postgres", err)
	}
	return &PGDB{db: db, config: config}
}

func (db *PGDB) Migrate(ctx context.Context, models []any) {
	curentLogger := db.db.Logger
	if !db.config.DebugMigrations {
		db.db.Logger = logger.Discard
	}
	if err := db.DB(ctx).AutoMigrate(models...); err != nil {
		log.Fatal(ctx, "could not migrate", err)
	}
	db.db.Logger = curentLogger
}
