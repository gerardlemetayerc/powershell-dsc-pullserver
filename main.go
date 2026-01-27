package main

import (
	"log"
	"fmt"
	"net/http"
	"go-dsc-pull/internal/routes"
	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal"
	"go-dsc-pull/internal/logs"
	"go-dsc-pull/internal/auth"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	
	// Chargement de la configuration globale (incluant SAML)
	appCfg, err := internal.LoadAppConfig("config.json")
	if err != nil {
		log.Fatalf("[INIT] Erreur chargement config globale: %v", err)
	}
	
	   // Initialisation SAML Service Provider si activé
	   samlMiddleware, err := auth.InitSamlMiddleware(appCfg)
	   if err != nil {
		   log.Fatalf("[SAML] %v", err)
	   }

	   // Initialisation automatique de la base (CREATE IF NOT EXISTS)
		dbCfg := &db.DBConfig{
			Driver:   appCfg.Driver,
			Server:   appCfg.Server,
			Port:     appCfg.Port,
			User:     appCfg.User,
			Password: appCfg.Password,
			Database: appCfg.Database,
		}
		db.InitDB(dbCfg)
		dbConn, err := db.OpenDB(dbCfg)
		if err != nil {
			log.Fatalf("Erreur ouverture DB: %v", err)
		}
		
		if appCfg.SAML.Enabled {
		       log.Printf("[SAML] Authentification SAML activée (EntityID: %s)", appCfg.SAML.EntityID)
	       } else {
		       log.Printf("[SAML] Authentification SAML désactivée (auth locale)")
	       }

	// Vérifie si la table users est vide et insère le compte admin si besoin
	{
		var count int
		err := dbConn.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
		if err == nil && count == 0 {
			hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
			_, err := dbConn.Exec("INSERT INTO users (first_name, last_name, email, password_hash, is_active) VALUES (?, ?, ?, ?, ?)", "Admin", "User", "admin@localhost", string(hash), 1)
			if err != nil {
				log.Printf("[INITDB] Erreur insertion admin: %v", err)
			} else {
				log.Printf("[INITDB] Compte admin créé: admin@localhost / password")
			}
		}
	}

	// --- Mux DSC (endpoints MS-DSCPM) ---
	dscMux := http.NewServeMux()
	routes.RegisterDSCRoutes(dscMux, dbConn)
	// --- Mux Web/API/GUI ---
	webMux := http.NewServeMux()
	routes.RegisterWebRoutes(webMux, dbConn, auth.JwtOrAPITokenAuthMiddleware(dbConn), samlMiddleware)

	// Ajout du logging middleware
	dscHandler := logs.LoggingMiddleware(dscMux)
	webHandler := logs.LoggingMiddleware(webMux)

	// Lancer les deux serveurs sur des ports différents avec HTTPS optionnel
	go func() {
		if appCfg.DSCPullServer.EnableHTTPS {
			log.Printf("Serveur DSC (HTTPS) sur :%d ...", appCfg.DSCPullServer.Port)
			log.Fatal(http.ListenAndServeTLS(
				fmt.Sprintf(":%d", appCfg.DSCPullServer.Port),
				appCfg.DSCPullServer.CertFile,
				appCfg.DSCPullServer.KeyFile,
				dscHandler,
			))
		} else {
			log.Printf("Serveur DSC (HTTP) sur :%d ...", appCfg.DSCPullServer.Port)
			log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", appCfg.DSCPullServer.Port), dscHandler))
		}
	}()

	if appCfg.WebUI.EnableHTTPS {
		log.Printf("Serveur IHM/API (HTTPS) sur :%d ... (IHM sur /web/)", appCfg.WebUI.Port)
		log.Fatal(http.ListenAndServeTLS(
			fmt.Sprintf(":%d", appCfg.WebUI.Port),
			appCfg.WebUI.CertFile,
			appCfg.WebUI.KeyFile,
			webHandler,
		))
	} else {
		log.Printf("Serveur IHM/API (HTTP) sur :%d ... (IHM sur /web/)", appCfg.WebUI.Port)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", appCfg.WebUI.Port), webHandler))
	}
}


