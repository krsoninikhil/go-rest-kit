package pgdb

import (
	"context"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Name            string
	Host            string
	Port            int
	User            string
	Password        string `log:"-"`
	DebugMigrations bool
}

type PGDB struct {
	db *gorm.DB
}

func (d *PGDB) DB(ctx context.Context) *gorm.DB {
	return d.db
}

func NewPGConnection(ctx context.Context, config Config) *PGDB {
	dns := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.Name)
	db, err := gorm.Open(postgres.Open(dns), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect postgres", err)
	}
	return &PGDB{db: db}
}

func (db *PGDB) Migrate(ctx context.Context, models []any) {
	if err := db.DB(ctx).AutoMigrate(models...); err != nil {
		log.Fatal(ctx, "could not migrate", err)
	}
}
