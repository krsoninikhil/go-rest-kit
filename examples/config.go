package main

import (
	"fmt"

	"github.com/krsoninikhil/go-rest-kit/auth"
	"github.com/krsoninikhil/go-rest-kit/config"
	"github.com/krsoninikhil/go-rest-kit/integrations/twilio"
	"github.com/krsoninikhil/go-rest-kit/pgdb"
)

// Set environment variables or add a .env with following values
// ENV=development
// DB_PASSWORD=gorestkit
// AUTH_SECRETKEY=secret
// AUTH_TWILIO_AUTHTOKEN=your-twillio-auth-token

type Config struct {
	DB     pgdb.Config
	Twilio twilio.Config
	Auth   auth.Config
	config.BaseConfig
}

func (c *Config) EnvPath() string    { return "../.env" }
func (c *Config) SourcePath() string { return fmt.Sprintf("./%s.yml", c.Env) }
