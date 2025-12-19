package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	_ "modernc.org/sqlite"
)

type DBConfig struct {
	Driver   string `json:"driver"`
	Server   string `json:"server"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

func LoadDBConfig(path string) (*DBConfig, error) {
	f, err := os.Open(path)
	if err != nil { return nil, err }
	defer f.Close()
	var cfg DBConfig
	if err := json.NewDecoder(f).Decode(&cfg); err != nil { return nil, err }
	return &cfg, nil
}

func OpenDB(cfg *DBConfig) (*sql.DB, error) {
	dsn := cfg.Database
	if cfg.Driver == "mysql" {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", cfg.User, cfg.Password, cfg.Server, cfg.Port, cfg.Database)
	} else if cfg.Driver == "postgres" {
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", cfg.User, cfg.Password, cfg.Server, cfg.Port, cfg.Database)
	}
	return sql.Open(cfg.Driver, dsn)
}