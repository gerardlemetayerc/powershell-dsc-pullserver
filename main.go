package main

import(
	"crypto"
	"crypto/tls"
	samlsp "github.com/crewjam/saml/samlsp"
	"log"
	"strings"
	"os"
	"fmt"
	"encoding/json"
	"context"
	"net/url"
	"net/http"
	"go-dsc-pull/internal/routes"
	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal"
	"go-dsc-pull/internal/schema"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Helper pour parser une URL ou panic
func mustParseURL(raw string) *url.URL {
       u, err := url.Parse(raw)
       if err != nil {
	       panic(err)
       }
       return u
}


// StatusRecorder is now in internal/schema

// loggingMiddleware logs all HTTP requests with method, path, remote addr, and status
func loggingMiddleware(next http.Handler) http.Handler {
	   return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		   rec := &schema.StatusRecorder{ResponseWriter: w, Status: 200}
		   next.ServeHTTP(rec, r)
		   log.Printf("[HTTP] %s %s %s %d", r.Method, r.URL.Path, r.RemoteAddr, rec.Status)
	   })
}

func main() {
	
	webMux := http.NewServeMux()
	// Chargement de la configuration globale (incluant SAML)
	appCfg, err := internal.LoadAppConfig("config.json")
	if err != nil {
		log.Fatalf("[INIT] Erreur chargement config globale: %v", err)
	}
	
	// Initialisation SAML Service Provider si activé
	var samlMiddleware *samlsp.Middleware
	if appCfg.SAML.Enabled {
		log.Println("[SAML] Initialisation du Service Provider SAML...")
		// Charge la clé privée et le certificat
		cert, err := tls.LoadX509KeyPair(appCfg.SAML.SPCertFile, appCfg.SAML.SPKeyFile)
		if err != nil {
			log.Fatalf("[SAML] Erreur chargement clé/cert SP: %v", err)
		}

		// Lecture du port web depuis la config
		webPort := appCfg.WebPort
		if webPort == 0 {
			webPort = 80
		}
		// Construction dynamique de l'URL du SP
		spURL := "http://127.0.0.1"
		if webPort != 80 {
			spURL = fmt.Sprintf("http://127.0.0.1:%d", webPort)
		}
				       // Récupération des métadonnées IdP via FetchMetadata (méthode crewjam/samlsp)
				       idpMetadataURL := appCfg.SAML.IdpMetadataURL
					   idpMetadata, err := samlsp.FetchMetadata(context.Background(), http.DefaultClient, *mustParseURL(idpMetadataURL))
				       if err != nil {
					       log.Fatalf("[SAML] Erreur récupération metadata IdP: %v", err)
				       }else{
						log.Printf("[SAML] Métadonnées IdP récupérées depuis %s", idpMetadataURL)

					   }
					       samlOptions := samlsp.Options{
						       URL: *mustParseURL(spURL),
						       Key: cert.PrivateKey.(crypto.Signer),
						       Certificate: cert.Leaf,
						       IDPMetadata:  idpMetadata,
					       }
					       samlMiddleware, err = samlsp.New(samlOptions)
					       // Désactive la vérification de signature SAML pour le dev (ne pas utiliser en prod)
					       if samlMiddleware != nil {
						       log.Printf("[SAML] SP middleware: %+v", samlMiddleware)
						       log.Printf("[SAML] SP EntityID: %s", samlMiddleware.ServiceProvider.EntityID)
						       log.Printf("[SAML] SP ACS URL: %s", samlMiddleware.ServiceProvider.AcsURL.String())
						       log.Printf("[SAML] SP Metadata URL: %s", samlMiddleware.ServiceProvider.MetadataURL.String())
					       }
				       if err != nil {
					       log.Fatalf("[SAML] Erreur initialisation SAML middleware: %v", err)
				       }
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
	       // Log SAML activé
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


	// Lecture des ports depuis config.json
	type portConfig struct {
		DscPort int `json:"dsc_port"`
		WebPort int `json:"web_port"`
	}
	var ports portConfig
	{
		f, err := os.Open("config.json")
		if err == nil {
			defer f.Close()
			json.NewDecoder(f).Decode(&ports)
		}
	}
	if ports.DscPort == 0 {
		ports.DscPort = 8081
	}
	if ports.WebPort == 0 {
		ports.WebPort = 8080
	}

	// --- Mux DSC (endpoints MS-DSCPM) ---
	dscMux := http.NewServeMux()
	routes.RegisterDSCRoutes(dscMux, dbConn)
	// Register all web/API/GUI routes
	routes.RegisterWebRoutes(webMux, dbConn, jwtAuthMiddleware, samlMiddleware)
	// Wrap mux with logging middleware
	dscHandler := loggingMiddleware(dscMux)
	webHandler := loggingMiddleware(webMux)

	// Lancer les deux serveurs sur des ports différents
	go func() {
		log.Printf("Serveur DSC (endpoints agents) sur :%d ...", ports.DscPort)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", ports.DscPort), dscHandler))
	}()
	log.Printf("Serveur IHM/API sur :%d ... (IHM sur /web/)", ports.WebPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", ports.WebPort), webHandler))
}


// Middleware de vérification JWT Bearer
func jwtAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "Missing Bearer token", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		secret := []byte("supersecretkey")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method")
			}
			return secret, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}