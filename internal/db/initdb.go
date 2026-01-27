
package db

import (
    "os"
    "go-dsc-pull/internal/logs"
    "io/ioutil"
    "path/filepath"
    "go-dsc-pull/utils"
)


// InitDB initialise la base de données à partir du schéma SQL (CREATE IF NOT EXISTS)
func InitDB(cfg *DBConfig) {
    database, err := OpenDB(cfg)
    if err != nil {
        logs.WriteLogFile("ERROR [INITDB] Failed to open DB: " + err.Error())
        os.Exit(1)
    }
    defer database.Close()

    exeDir, err := utils.ExePath()
    if err != nil {
        logs.WriteLogFile("ERROR [INITDB] Failed to get executable path: " + err.Error())
        os.Exit(1)
    }
    schemaPath := filepath.Join(filepath.Dir(exeDir), "internal", "db", "schema.sql")
    schema, err := ioutil.ReadFile(schemaPath)
    if err != nil {
        logs.WriteLogFile("ERROR [INITDB] Failed to read schema.sql: " + err.Error())
        os.Exit(1)
    }

    _, err = database.Exec(string(schema))
    if err != nil {
        logs.WriteLogFile("ERROR [INITDB] Failed to execute SQL schema: " + err.Error())
        os.Exit(1)
    }

    logs.WriteLogFile("INFO [INITDB] DB schema checked/created.")
}