
package db

import (
    "log"
    "io/ioutil"
    "path/filepath"
    "go-dsc-pull/utils"
)


// InitDB initialise la base de données à partir du schéma SQL (CREATE IF NOT EXISTS)
func InitDB(cfg *DBConfig) {
    database, err := OpenDB(cfg)
    if err != nil {
        log.Fatalf("[INITDB] Erreur ouverture DB: %v", err)
    }
    defer database.Close()

    exeDir, err := utils.ExePath()
    if err != nil {
        log.Fatalf("[INITDB] Erreur récupération chemin exécutable: %v", err)
    }
    schemaPath := filepath.Join(filepath.Dir(exeDir), "internal", "db", "schema.sql")
    schema, err := ioutil.ReadFile(schemaPath)
    if err != nil {
        log.Fatalf("[INITDB] Erreur lecture schema.sql: %v", err)
    }

    _, err = database.Exec(string(schema))
    if err != nil {
        log.Fatalf("[INITDB] Erreur exécution du schéma SQL: %v", err)
    }

    log.Println("[INITDB] Schéma DB vérifié/créé.")
}