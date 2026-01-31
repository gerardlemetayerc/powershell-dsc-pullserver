package main

import(
	"crypto"
	"crypto/tls"
	samlsp "github.com/crewjam/saml/samlsp"
	"log"
	"strings"
	"fmt"
	"context"
	"net/url"
	"net/http"
	"go-dsc-pull/internal/routes"
	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal"
	"go-dsc-pull/internal/schema"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"database/sql"
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

		// Utilise l'entity_id de la config pour l'URL du SP
		spURL := appCfg.SAML.EntityID
		idpMetadataURL := appCfg.SAML.IdpMetadataURL
		idpMetadata, err := samlsp.FetchMetadata(context.Background(), http.DefaultClient, *mustParseURL(idpMetadataURL))
		if err != nil {
			log.Fatalf("[SAML] Erreur récupération metadata IdP: %v", err)
		}
		samlOptions := samlsp.Options{
			URL: *mustParseURL(spURL),
			Key: cert.PrivateKey.(crypto.Signer),
			Certificate: cert.Leaf,
			IDPMetadata:  idpMetadata,
		}
		samlMiddleware, err = samlsp.New(samlOptions)
		if samlMiddleware != nil {
			// Force explicitement l'EntityID à la valeur de la config
			samlMiddleware.ServiceProvider.EntityID = appCfg.SAML.EntityID
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
			res, err := dbConn.Exec("INSERT INTO users (first_name, last_name, email, password_hash, is_active) VALUES (?, ?, ?, ?, ?)", "Admin", "User", "admin@localhost", string(hash), 1)
			if err != nil {
				log.Printf("[INITDB] Erreur insertion admin: %v", err)
			} else {
				adminID, _ := res.LastInsertId()
				// Assign Administrator role to default admin user
				_, err := dbConn.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", adminID, 1)
				if err != nil {
					log.Printf("[INITDB] Erreur assignation rôle admin: %v", err)
				} else {
					log.Printf("[INITDB] Compte admin créé: admin@localhost / password (rôle: Administrator)")
				}
			}
		}
	}


	// Les ports sont maintenant dans appCfg.DSCPullServer.Port et appCfg.WebUI.Port

	// --- Mux DSC (endpoints MS-DSCPM) ---
	dscMux := http.NewServeMux()
	routes.RegisterDSCRoutes(dscMux, dbConn)
	// Register all web/API/GUI routes
	routes.RegisterWebRoutes(webMux, dbConn, jwtOrAPITokenAuthMiddleware(dbConn), samlMiddleware)
	// Wrap mux with logging middleware
	dscHandler := loggingMiddleware(dscMux)
	webHandler := loggingMiddleware(webMux)

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


// Middleware de vérification JWT Bearer
func jwtOrAPITokenAuthMiddleware(dbConn *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				tokenStr := strings.TrimPrefix(auth, "Bearer ")
				secret := []byte("supersecretkey")
				jwtToken, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("Unexpected signing method")
					}
					return secret, nil
				})
				if err == nil && jwtToken.Valid {
					if claims, ok := jwtToken.Claims.(jwt.MapClaims); ok {
						if sub, ok := claims["sub"].(string); ok {
							ctx := context.WithValue(r.Context(), "userId", sub)
							r = r.WithContext(ctx)
						}
					}
					next.ServeHTTP(w, r)
					return
				}
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}
			// Si ce n'est pas Bearer, tente API token
			if strings.HasPrefix(auth, "Token ") {
				tokenStr := strings.TrimPrefix(auth, "Token ")
				userId, apiErr := db.CheckAPIToken(dbConn, tokenStr)
				if apiErr == nil && userId > 0 {
					ctx := context.WithValue(r.Context(), "userId", userId)
					r = r.WithContext(ctx)
					next.ServeHTTP(w, r)
					return
				}
				http.Error(w, "Invalid or expired API token", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		})
	}
}