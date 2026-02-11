package db

import (
	"time"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"go-dsc-pull/utils"
	"go-dsc-pull/internal/schema"
	 _ "modernc.org/sqlite"
	 _ "github.com/denisenkom/go-mssqldb"
)

func UpdateLastLogon(db *sql.DB, userId interface{}) error {

	_, err := db.Exec("UPDATE users SET last_logon_date=? WHERE id=?", time.Now().Format("2006-01-02 15:04:05"), userId)
	return err
}

func UpdateRole(db *sql.DB, userId interface{}, role string) error {
	_, err := db.Exec("UPDATE users SET role=? WHERE id=?", role, userId)
	return err
}

func LoadDBConfig(path string) (*schema.DBConfig, error) {
	var absPath string
	if filepath.IsAbs(path) {
		absPath = path
	} else {
		exeDir, err := utils.ExePath()
		if err != nil { return nil, err }
		absPath = filepath.Join(filepath.Dir(exeDir), path)
	}
	f, err := os.Open(absPath)
	if err != nil { return nil, err }
	defer f.Close()
	var raw map[string]interface{}
	if err := json.NewDecoder(f).Decode(&raw); err != nil { return nil, err }
	dbRaw, ok := raw["database"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("database config not found or invalid format")
	}
	cfg := &schema.DBConfig{}
	if v, ok := dbRaw["driver"].(string); ok { cfg.Driver = v }
	if v, ok := dbRaw["server"].(string); ok { cfg.Server = v }
	if v, ok := dbRaw["port"].(float64); ok { cfg.Port = int(v) }
	if v, ok := dbRaw["user"].(string); ok { cfg.User = v }
	if v, ok := dbRaw["password"].(string); ok { cfg.Password = v }
	if v, ok := dbRaw["name"].(string); ok { cfg.Database = v }
	return cfg, nil
}

func OpenDB(cfg *schema.DBConfig) (*sql.DB, error) {
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
	} else if cfg.Driver == "mssql" || cfg.Driver == "sqlserver" {
		dsn = fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s", cfg.User, cfg.Password, cfg.Server, cfg.Port, cfg.Database)
	}
	return sql.Open(cfg.Driver, dsn)
}