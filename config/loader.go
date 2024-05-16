package config

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type BaseConfig struct {
	Env string
}

func (c *BaseConfig) SetEnv(env string) { c.Env = env }

type AppConfig interface {
	SourcePath() string
	EnvPath() string
	SetEnv(string)
}

func Load(ctx context.Context, target AppConfig) {
	// override from env
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := godotenv.Load(target.EnvPath()); err == nil {
		fmt.Print(ctx, "Environment variables set from", target.EnvPath())
	}

	// read from config files
	env := viper.GetString("env")
	if env == "" {
		env = "development"
	}
	target.SetEnv(strings.ToLower(env))
	viper.SetConfigFile(target.SourcePath())
	if err := viper.MergeInConfig(); err != nil {
		log.Fatal(ctx, "error reading environment config file", err)
	}

	if err := viper.Unmarshal(target); err != nil {
		log.Fatal(ctx, "error parsing config", err)
	}
	// log.Printf("Loaded config %+v", target)
}
