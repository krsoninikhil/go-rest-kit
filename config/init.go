package config

import (
	"context"
	"go/build"
	"log"
	"os"
)

func init() {
	GetOrSetGoPath()
}

func GetOrSetGoPath() string {
	if custom := os.Getenv("GOPATH"); custom != "" {
		return custom
	}
	if err := os.Setenv("GOPATH", build.Default.GOPATH); err != nil {
		log.Fatal(context.Background(), "error setting GOPATH", err)
	}
	return os.Getenv("GOPATH")
}
