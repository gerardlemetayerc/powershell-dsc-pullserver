package db

import (
	"time"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"go-dsc-pull/utils"
	_ "modernc.org/sqlite"
)

// Met à jour la date de dernière connexion pour un utilisateur
func UpdateLastLogon(db *sql.DB, userId interface{}) error {
	_, err := db.Exec("UPDATE users SET last_logon_date=? WHERE id=?", time.Now().Format("2006-01-02 15:04:05"), userId)
	return err
}

type DBConfig struct {
	Driver   string `json:"driver"`
	Server   string `json:"server"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

func LoadDBConfig(path string) (*DBConfig, error) {
	exeDir, err := utils.ExePath()
	if err != nil { return nil, err }
	absPath := filepath.Join(filepath.Dir(exeDir), path)
	f, err := os.Open(absPath)
	if err != nil { return nil, err }
	defer f.Close()
	var cfg DBConfig
	if err := json.NewDecoder(f).Decode(&cfg); err != nil { return nil, err }
	return &cfg, nil
}

func OpenDB(cfg *DBConfig) (*sql.DB, error) {
	   dsn := cfg.Database
	   if cfg.Driver == "sqlite" && !filepath.IsAbs(dsn) {
		   exePath, err := utils.ExePath()
		   if err == nil {
			   baseDir := filepath.Dir(exePath)
			   dsn = filepath.Join(baseDir, dsn)
		   }
	   }
	   if cfg.Driver == "mysql" {
		   dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", cfg.User, cfg.Password, cfg.Server, cfg.Port, cfg.Database)
	   } else if cfg.Driver == "postgres" {
		   dsn = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", cfg.User, cfg.Password, cfg.Server, cfg.Port, cfg.Database)
	   }
	   return sql.Open(cfg.Driver, dsn)
}