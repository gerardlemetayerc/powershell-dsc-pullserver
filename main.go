package main

import (
	"fmt"
	"net/http"
	"runtime"
	"os"
	"path/filepath"
	"go-dsc-pull/utils"
	"go-dsc-pull/internal/routes"
	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal"
	"go-dsc-pull/internal/logs"
	"go-dsc-pull/internal/auth"
	"go-dsc-pull/internal/service"
	"golang.org/x/crypto/bcrypt"
)

func main() {

	   runApp := func() {
		   // Résout dynamiquement les chemins de cert/key si nécessaire
		   resolveCertKeyPath := func(certFile, keyFile string) (string, string) {
			   if filepath.IsAbs(certFile) && filepath.IsAbs(keyFile) {
				   return certFile, keyFile
			   }
			   exePath, err := utils.ExePath()
			   baseDir := ""
			   if err == nil {
				   baseDir = filepath.Dir(exePath)
			   }
			   if !filepath.IsAbs(certFile) {
				   certFile = filepath.Join(baseDir, certFile)
			   }
			   if !filepath.IsAbs(keyFile) {
				   keyFile = filepath.Join(baseDir, keyFile)
			   }
			   return certFile, keyFile
		   }
		   // Chargement de la configuration globale (incluant SAML)
		   appCfg, err := internal.LoadAppConfig("config.json")
		   if err != nil {
			   logs.WriteLogFile(fmt.Sprintf("ERROR [INIT] Failed to load global config: %v", err))
			   os.Exit(1)
		   }

		   // Initialisation SAML Service Provider si activé
		   samlMiddleware, err := auth.InitSamlMiddleware(appCfg)
		   if err != nil {
			   logs.WriteLogFile(fmt.Sprintf("ERROR [SAML] %v", err))
			   os.Exit(1)
		   }

		   // Initialisation automatique de la base (CREATE IF NOT EXISTS)
		   dbPath := appCfg.Database
		   if !filepath.IsAbs(dbPath) {
			   exePath, err := utils.ExePath()
			   baseDir := ""
			   if err == nil {
				   baseDir = filepath.Dir(exePath)
				   dbPath = filepath.Join(baseDir, dbPath)
			   }
		   }
		   dbCfg := &db.DBConfig{
			   Driver:   appCfg.Driver,
			   Server:   appCfg.Server,
			   Port:     appCfg.Port,
			   User:     appCfg.User,
			   Password: appCfg.Password,
			   Database: dbPath,
		   }
		   db.InitDB(dbCfg)
		   dbConn, err := db.OpenDB(dbCfg)
		   if err != nil {
			   logs.WriteLogFile(fmt.Sprintf("ERROR [INITDB] Failed to open DB: %v", err))
			   os.Exit(1)
		   }

		   if appCfg.SAML.Enabled {
			   logs.WriteLogFile(fmt.Sprintf("INFO [SAML] SAML Authentication activated (EntityID: %s)", appCfg.SAML.EntityID))
		   } else {
			   logs.WriteLogFile("INFO [SAML] SAML Authentication deactivated (local auth)")
		   }

		   // Vérifie si la table users est vide et insère le compte admin si besoin
		   {
			   var count int
			   err := dbConn.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
			   if err == nil && count == 0 {
				   hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
				   _, err := dbConn.Exec("INSERT INTO users (first_name, last_name, email, password_hash, is_active) VALUES (?, ?, ?, ?, ?)", "Admin", "User", "admin@localhost", string(hash), 1)
				   if err != nil {
					   logs.WriteLogFile(fmt.Sprintf("ERROR [INITDB] Failed to insert admin: %v", err))
				   } else {
					   logs.WriteLogFile("INFO [INITDB] Admin account created: admin@localhost / password")
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
				   logs.WriteLogFile(fmt.Sprintf("INFO [DSC Core Server] DSC Server (HTTPS) on :%d ...", appCfg.DSCPullServer.Port))
				   certFile, keyFile := resolveCertKeyPath(appCfg.DSCPullServer.CertFile, appCfg.DSCPullServer.KeyFile)
				   err := http.ListenAndServeTLS(
					   fmt.Sprintf(":%d", appCfg.DSCPullServer.Port),
					   certFile,
					   keyFile,
					   dscHandler,
				   )
				   if err != nil {
					   logs.WriteLogFile(fmt.Sprintf("ERROR [DSC Core Server] Error starting HTTPS server: %v", err))
					   os.Exit(1)
				   }
			   } else {
				   logs.WriteLogFile(fmt.Sprintf("INFO [DSC WebUi] DSC Server (HTTP) on :%d ...", appCfg.DSCPullServer.Port))
				   err := http.ListenAndServe(fmt.Sprintf(":%d", appCfg.DSCPullServer.Port), dscHandler)
				   if err != nil {
					   logs.WriteLogFile(fmt.Sprintf("ERROR [DSC WebUi] Error starting HTTP server: %v", err))
					   os.Exit(1)
				   }
			   }
		   }()

		   if appCfg.WebUI.EnableHTTPS {
			   logs.WriteLogFile(fmt.Sprintf("INFO [DSC WebUI] IHM/API (HTTPS) on :%d ... (IHM on /web/)", appCfg.WebUI.Port))
			   certFile, keyFile := resolveCertKeyPath(appCfg.WebUI.CertFile, appCfg.WebUI.KeyFile)
			   err := http.ListenAndServeTLS(
				   fmt.Sprintf(":%d", appCfg.WebUI.Port),
				   certFile,
				   keyFile,
				   webHandler,
			   )
			   if err != nil {
				   logs.WriteLogFile(fmt.Sprintf("ERROR [DSC WebUI] Error starting HTTPS server: %v", err))
				   os.Exit(1)
			   }
		   } else {
			   logs.WriteLogFile(fmt.Sprintf("INFO [DSC WebUI] IHM/API (HTTP) on :%d ... (IHM on /web/)", appCfg.WebUI.Port))
			   err := http.ListenAndServe(fmt.Sprintf(":%d", appCfg.WebUI.Port), webHandler)
			   if err != nil {
				   logs.WriteLogFile(fmt.Sprintf("ERROR [DSC WebUI] Error starting HTTP server: %v", err))
				   os.Exit(1)
			   }
		   }
	   }

	   if runtime.GOOS == "windows" {
		   err := service.StartWindowsService("DSCPullServer", runApp)
		   if err != nil {
			   logs.WriteLogFile(fmt.Sprintf("ERROR [SERVICE] %v", err))
			   os.Exit(1)
		   }
	   } else {
		   runApp()
	   }
}


