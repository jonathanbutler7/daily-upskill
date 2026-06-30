package config

import (
	"time"
)

type AppConfig struct {
	// initialize config with default values
	TestString   string
	TestInt      int
	TestBool     bool
	TestAddress  string `wdns:"true"`
	TestDuration time.Duration
}

func NewDefaultAppConfig() *AppConfig {
	return &AppConfig{
		TestString:   "test-string",
		TestInt:      3,
		TestBool:     true,
		TestAddress:  "google.com:8080",
		TestDuration: time.Second,
	}
}
