// Package config implements config struct, which is used in all layers
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/goccy/go-yaml"
)

// TODO: add validator

type app struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Env     string `yaml:"env" env-default:"dev"`
}

type grpc struct {
	Port     string `yaml:"port" env-default:"50051"`
	Host     string `yaml:"host" env-default:"localhost"`
	Protocol string `yaml:"protocol" env-default:"tcp"`
}

type jwt struct {
	TTL       time.Duration `yaml:"ttl" env-default:"24h"`
	SecretKey string        `yaml:"key"`
}

type postgres struct {
	DSN             string        `yaml:"dsn"`
	MaxOpenConn     int           `yaml:"max_open_conn" env-default:"15"`
	MaxIdleConn     int           `yaml:"max_idle_conn" env-default:"5"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env-default:"30m"`
}

type redis struct {
	TTL      time.Duration `yaml:"ttl"`
	Addr     string        `yaml:"addr"`
	DB       int           `yaml:"db"`
	Password string        `yaml:"password"`
}

type Config struct {
	App    app `yaml:"app"`
	Server struct {
		GRPC      grpc          `yaml:"grpc"`
		RWTimeout time.Duration `yaml:"rw_timeout" env-default:"2s"`
	} `yaml:"server"`
	Secrets struct {
		JWT      jwt      `yaml:"jwt"`
		Postgres postgres `yaml:"postgres"`
		Redis    redis    `yaml:"redis"`
	} `yaml:"secrets"`
}

// Validate implements validation of config fields (dsn, secret key)
func (c *Config) Validate() error {
	if c.Secrets.JWT.SecretKey == "" {
		return errors.New("jwt secret key is empty")
	}
	// DSN is validated when gorm connecting to pg
	if c.Secrets.Postgres.DSN == "" {
		return errors.New("invalid postgres dsn")
	}

	if c.Secrets.JWT.TTL < time.Minute {
		return errors.New("jwt ttl is too short")
	}

	if c.Secrets.Redis.Addr == "" {
		return errors.New("invalid addr for redis")
	}

	if c.Secrets.Redis.TTL < time.Second {
		return errors.New("redis ttl is too short")
	}

	return nil
}

// Load builds up config
func Load(path string) (*Config, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	bytes = []byte(os.ExpandEnv(string(bytes)))

	var cfg Config
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// MustLoad is just an add-on to Load(), it just doesn't return an error.
func MustLoad(path string) *Config {
	cfg, err := Load(path)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	return cfg
}
