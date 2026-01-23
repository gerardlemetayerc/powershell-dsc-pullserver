
package db

import (
    "log"
    "io/ioutil"
)


// InitDB initialise la base de données à partir du schéma SQL (CREATE IF NOT EXISTS)
func InitDB(cfg *DBConfig) {
    database, err := OpenDB(cfg)
    if err != nil {
        log.Fatalf("[INITDB] Erreur ouverture DB: %v", err)
    }
    defer database.Close()

    schema, err := ioutil.ReadFile("internal\\db\\schema.sql")
    if err != nil {
        log.Fatalf("[INITDB] Erreur lecture schema.sql: %v", err)
    }

    _, err = database.Exec(string(schema))
    if err != nil {
        log.Fatalf("[INITDB] Erreur exécution du schéma SQL: %v", err)
    }

    log.Println("[INITDB] Schéma DB vérifié/créé.")
}