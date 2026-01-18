package main

import(
	"crypto"
	"crypto/tls"
	"database/sql"
	samlsp "github.com/crewjam/saml/samlsp"
	"log"
	"strings"
	"os"
	"fmt"
	"time"
	"encoding/json"
	"context"
	"net/url"
	"net/http"
	"html/template"
	"go-dsc-pull/handlers"
	"go-dsc-pull/utils"
	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	// À adapter selon la spec, ici minimal
	NodeName string `json:"NodeName"`
}

type RegisterResponse struct {
	AgentId string `json:"AgentId"`
}

// Helper pour parser une URL ou panic
func mustParseURL(raw string) *url.URL {
       u, err := url.Parse(raw)
       if err != nil {
	       panic(err)
       }
       return u
}


// statusRecorder wraps http.ResponseWriter to capture status code
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// loggingMiddleware logs all HTTP requests with method, path, remote addr, and status
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: 200}
		next.ServeHTTP(rec, r)
		log.Printf("[HTTP] %s %s %s %d", r.Method, r.URL.Path, r.RemoteAddr, rec.status)
	})
}

func main() {
		   // --- Mux IHM/API ---
		   webMux := http.NewServeMux()
		   // Endpoint pour exposer les infos utilisateur SAML mappées
		   webMux.Handle("GET /api/v1/saml/userinfo", http.HandlerFunc(handlers.SAMLUserInfoHandler))
		   // Endpoint pour exposer l'état SAML (pour le bouton login SAML)
		   webMux.Handle("GET /api/v1/saml/enabled", http.HandlerFunc(handlers.SAMLStatusHandler))
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
				       samlMiddleware, err = samlsp.New(samlsp.Options{
					       URL: *mustParseURL(spURL),
					       Key: cert.PrivateKey.(crypto.Signer),
					       Certificate: cert.Leaf,
					       IDPMetadata:  idpMetadata,
				       })
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

	// ...existing code...



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
	dscMux.HandleFunc("PUT /PSDSCPullServer.svc/{node}", handlers.RegisterHandler)
	dscMux.HandleFunc("GET /PSDSCPullServer.svc/{node}/{config}/ConfigurationContent", handlers.ConfigurationContentHandler)
	dscMux.HandleFunc("POST /PSDSCPullServer.svc/{node}/SendReport", handlers.SendReportHandler)
	dscMux.HandleFunc("POST /PSDSCPullServer.svc/{id}/GetDscAction", func(w http.ResponseWriter, r *http.Request) {
		raw := r.PathValue("id")
		agentId := utils.ExtractAgentId(raw)
		handlers.GetDscActionNodeHandlerWithId(w, r, agentId)
	})
	dscMux.HandleFunc("GET /PSDSCPullServer.svc/{module}/ModuleContent", func(w http.ResponseWriter, r *http.Request) {
		// Exemple de segment: Modules(ModuleName='FileContentDsc',ModuleVersion='1.3.0.151')
		moduleSeg := r.PathValue("module")
		var name, version string
		// Extraction robuste avec regex
		// Format attendu: Modules(ModuleName='...',ModuleVersion='...')
		if strings.HasPrefix(moduleSeg, "Modules(") && strings.HasSuffix(moduleSeg, ")") {
			inner := moduleSeg[len("Modules(") : len(moduleSeg)-1]
			parts := strings.Split(inner, ",")
			for _, part := range parts {
				kv := strings.SplitN(part, "=", 2)
				if len(kv) == 2 {
					k := strings.TrimSpace(kv[0])
					v := strings.Trim(kv[1], "'\"")
					if k == "ModuleName" {
						name = v
					} else if k == "ModuleVersion" {
						version = v
					}
				}
			}
		}
		log.Printf("[MODULECONTENT] Agent request ModuleName=%s, ModuleVersion=%s", name, version)
		checksum := r.URL.Query().Get("Checksum")
		q := r.URL.Query()
		q.Set("ModuleName", name)
		q.Set("ModuleVersion", version)
		if checksum != "" { q.Set("Checksum", checksum) }
		r.URL.RawQuery = q.Encode()
		handlers.ModuleDownloadHandler(dbConn)(w, r)
	})
	       // --- Mux IHM/API ---
	       //webMux := http.NewServeMux()
	       // --- Endpoints SAML (placeholders) ---
				       if samlMiddleware != nil {
					       log.Println("[SAML] Montage du handler /saml/ (middleware actif)")
					       webMux.Handle("/saml/", samlMiddleware)
				       } else {
					       log.Println("[SAML] Middleware SAML non initialisé, /saml/ non monté")
				       }
	// API REST: liste des agents
	webMux.Handle("GET /api/v1/agents", jwtAuthMiddleware(http.HandlerFunc(handlers.AgentAPIHandler)))
	// API REST: configurations d'un agent
	webMux.Handle("GET /api/v1/agents/{id}/configs", jwtAuthMiddleware(http.HandlerFunc(handlers.AgentConfigsAPIHandler)))
	webMux.Handle("POST /api/v1/agents/{id}/configs", jwtAuthMiddleware(http.HandlerFunc(handlers.AgentConfigsAPIHandlerPostDelete)))
	webMux.Handle("DELETE /api/v1/agents/{id}/configs", jwtAuthMiddleware(http.HandlerFunc(handlers.AgentConfigsAPIHandlerPostDelete)))
	// API REST: rapports d'un agent
	webMux.Handle("GET /api/v1/agents/{id}/reports", jwtAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[API] /api/v1/agents/%s/reports?%s", r.PathValue("id"), r.URL.RawQuery)
		handlers.AgentReportsListHandler(w, r)
	})))
	webMux.Handle("GET /api/v1/agents/{id}/reports/latest", jwtAuthMiddleware(http.HandlerFunc(handlers.AgentReportsLatestHandler)))
	webMux.Handle("GET /api/v1/agents/{id}/reports/{jobid}", jwtAuthMiddleware(http.HandlerFunc(handlers.AgentReportsByJobIdHandler)))
	// API REST: infos d'un agent
	webMux.Handle("GET /api/v1/agents/{id}", jwtAuthMiddleware(http.HandlerFunc(handlers.AgentByIdAPIHandler)))
	// API REST: modules DSC
	webMux.Handle("POST /api/v1/modules/upload", jwtAuthMiddleware(http.HandlerFunc(handlers.ModuleUploadHandler(dbConn))))
	webMux.Handle("GET /api/v1/modules", jwtAuthMiddleware(http.HandlerFunc(handlers.ModuleListHandler(dbConn))))
	webMux.Handle("DELETE /api/v1/modules/delete", jwtAuthMiddleware(http.HandlerFunc(handlers.ModuleDeleteHandler(dbConn))))

	// API REST: properties
	webMux.Handle("GET /api/v1/properties", jwtAuthMiddleware(http.HandlerFunc(handlers.PropertiesListHandler)))
	webMux.Handle("POST /api/v1/properties", jwtAuthMiddleware(http.HandlerFunc(handlers.PropertiesCreateHandler)))
	webMux.Handle("GET /api/v1/properties/{id}", jwtAuthMiddleware(http.HandlerFunc(handlers.PropertiesGetHandler)))
	webMux.Handle("PUT /api/v1/properties/{id}", jwtAuthMiddleware(http.HandlerFunc(handlers.PropertiesUpdateHandler)))
	webMux.Handle("DELETE /api/v1/properties/{id}", jwtAuthMiddleware(http.HandlerFunc(handlers.PropertiesDeleteHandler)))


	// API REST: configuration_model CRUD
	webMux.Handle("POST /api/v1/configuration_models", jwtAuthMiddleware(http.HandlerFunc(handlers.CreateConfigurationModelHandler)))
	webMux.Handle("GET /api/v1/configuration_models", jwtAuthMiddleware(http.HandlerFunc(handlers.ListConfigurationModelsHandler)))
	webMux.Handle("GET /api/v1/configuration_models/{id}", jwtAuthMiddleware(http.HandlerFunc(handlers.GetConfigurationModelHandler)))
	webMux.Handle("PUT /api/v1/configuration_models/{id}", jwtAuthMiddleware(http.HandlerFunc(handlers.UpdateConfigurationModelHandler)))
	webMux.Handle("DELETE /api/v1/configuration_models", jwtAuthMiddleware(http.HandlerFunc(handlers.DeleteConfigurationModelHandler)))

	// Endpoint de login pour JWT (validation via DB)
	webMux.Handle("POST /api/v1/login", handlers.LoginHandler(dbConn))


	// Handler pour /web/login
	webMux.HandleFunc("/web/login", func(w http.ResponseWriter, r *http.Request) {
		// Vérifie la présence d'un token JWT dans le cookie (optionnel, car le JS stocke dans localStorage)
		// Ici, on laisse toujours afficher la page login, la redirection sera gérée côté JS après login
		http.ServeFile(w, r, "templates/login.tmpl")
	})

	// Servir la page index via le template Go
	webMux.HandleFunc("/web", handlers.WebIndexHandler)
	webMux.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))
	// Handler Go pour /web/node/{id} et /web/node/{nodename}/properties
	webMux.HandleFunc("/web/node/", func(w http.ResponseWriter, r *http.Request) {
	   if strings.HasSuffix(r.URL.Path, "/properties") {
		   handlers.WebNodePropertiesHandler(w, r)
		   return
	   }
	   handlers.WebNodeHandler(w, r)
	})

	// Handler Go pour /web/modules
	webMux.HandleFunc("/web/modules", handlers.WebModulesHandler)

	// Handler Go pour /web/configuration_model
	webMux.HandleFunc("/web/configuration_model", handlers.WebConfigurationModelHandler)

	// Handler Go pour /web/properties.html
	webMux.HandleFunc("/web/properties.html", handlers.WebPropertiesHandler)

				// Handler Go pour /web/users
				webMux.HandleFunc("/web/users", func(w http.ResponseWriter, r *http.Request) {
					tmpl, err := template.New("layout.html").
						ParseFiles(
							"templates/layout.html",
							"templates/head.tmpl",
							"templates/menu.tmpl",
							"templates/footer.tmpl",
							"templates/users.tmpl",
						)
					if err != nil {
						http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
						return
					}
					data := map[string]interface{}{ "Title": "Utilisateurs" }
					tmpl.ExecuteTemplate(w, "layout", data)
				})

				// Handler Go pour /web/user_edit
				webMux.HandleFunc("/web/user_edit", func(w http.ResponseWriter, r *http.Request) {
					tmpl, err := template.New("layout.html").
						ParseFiles(
							"templates/layout.html",
							"templates/head.tmpl",
							"templates/menu.tmpl",
							"templates/footer.tmpl",
							"templates/user_edit.tmpl",
						)
					if err != nil {
						http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
						return
					}
					data := map[string]interface{}{ "Title": "Édition utilisateur" }
					tmpl.ExecuteTemplate(w, "layout", data)
				})

				// Handler Go pour /web/user_password
				webMux.HandleFunc("/web/user_password", func(w http.ResponseWriter, r *http.Request) {
					tmpl, err := template.New("layout.html").
						ParseFiles(
							"templates/layout.html",
							"templates/head.tmpl",
							"templates/menu.tmpl",
							"templates/footer.tmpl",
							"templates/user_password.tmpl",
						)
					if err != nil {
						http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
						return
					}
					data := map[string]interface{}{ "Title": "Mot de passe utilisateur" }
					tmpl.ExecuteTemplate(w, "layout", data)
				})
			webMux.Handle("POST /api/v1/users/{id}/password", jwtAuthMiddleware(http.HandlerFunc(handlers.ChangeUserPasswordHandler(dbConn))))
			// API REST: utilisateurs
			webMux.Handle("GET /api/v1/users", jwtAuthMiddleware(http.HandlerFunc(handlers.ListUsersHandler(dbConn))))
			webMux.Handle("GET /api/v1/users/{id}", jwtAuthMiddleware(http.HandlerFunc(handlers.GetUserHandler(dbConn))))
			webMux.Handle("POST /api/v1/users", jwtAuthMiddleware(http.HandlerFunc(handlers.CreateUserHandler(dbConn))))
			webMux.Handle("PUT /api/v1/users/{id}", jwtAuthMiddleware(http.HandlerFunc(handlers.UpdateUserHandler(dbConn))))
			webMux.Handle("DELETE /api/v1/users/{id}", jwtAuthMiddleware(http.HandlerFunc(handlers.DeleteUserHandler(dbConn))))
			webMux.Handle("POST /api/v1/users/{id}/active", jwtAuthMiddleware(http.HandlerFunc(handlers.SetUserActiveHandler(dbConn))))
	
		   // Endpoint de test protégé SAML pour valider le flux SSO
	   if samlMiddleware != nil {
		   webMux.Handle("/web/login/saml", samlMiddleware.RequireAccount(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			   // Affiche dynamiquement toutes les assertions SAML reçues (clé par clé)
			   log.Println("[SAML] Assertions reçues (toutes clés) :")
			   // Liste des claims potentiels (pour debug)
			   claimKeys := []string{
				   "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
				   "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname",
				   "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname",
				   "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name",
				   "http://schemas.microsoft.com/identity/claims/displayname",
			   }
			   for _, k := range claimKeys {
				   v := samlsp.AttributeFromContext(r.Context(), k)
				   log.Printf("[SAML] %s = %s", k, v)
			   }

			   // Mapping automatique depuis la conf (via internal.GetSAMLUserMapping)
			   mapping, err := internal.GetSAMLUserMapping()
			   if err != nil {
				   log.Printf("[SAML] Erreur lecture mapping SAML: %v", err)
				   http.Error(w, "Erreur mapping SAML", http.StatusInternalServerError)
				   return
			   }
			   getAttr := func(uri string) string {
				   return samlsp.AttributeFromContext(r.Context(), uri)
			   }
			   email := getAttr(mapping["email"])
			   firstName := getAttr(mapping["givenName"])
			   lastName := getAttr(mapping["sn"])
			   if email == "" {
				   log.Printf("[SAML] Aucun email trouvé dans les assertions SAML, accès refusé")
				   http.Error(w, "Email SAML manquant", http.StatusForbidden)
				   return
			   }
			   // Vérifie si l'utilisateur existe déjà
			   var userId int
			   err = dbConn.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userId)
			   if err == sql.ErrNoRows {
				   // Crée l'utilisateur
				   log.Printf("[SAML] Création nouvel utilisateur: %s %s <%s>", firstName, lastName, email)
				   _, err := dbConn.Exec("INSERT INTO users (first_name, last_name, email, password_hash, is_active) VALUES (?, ?, ?, '', 1)", firstName, lastName, email)
				   if err != nil {
					   log.Printf("[SAML] Erreur création utilisateur: %v", err)
					   http.Error(w, "Erreur création utilisateur", http.StatusInternalServerError)
					   return
				   }
			   } else if err != nil && err != sql.ErrNoRows {
				   log.Printf("[SAML] Erreur DB: %v", err)
				   http.Error(w, "Erreur DB", http.StatusInternalServerError)
				   return
			   }
			   // Génère le JWT applicatif comme dans LoginHandler
			   secret := []byte("supersecretkey")
			   expiresAt := time.Now().Add(60 * time.Minute).Unix()
			   claims := jwt.MapClaims{
				   "sub": email,
				   "exp": expiresAt,
			   }
			   token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			   signed, err := token.SignedString(secret)
			   if err != nil {
				   log.Printf("[SAML] Erreur génération JWT: %v", err)
				   http.Error(w, "Erreur JWT", http.StatusInternalServerError)
				   return
			   }
			   // Renvoie une page HTML/JS qui stocke le JWT dans localStorage puis redirige
			   w.Header().Set("Content-Type", "text/html; charset=utf-8")
			   fmt.Fprintf(w, `<!DOCTYPE html><html lang='fr'><head><meta charset='UTF-8'><title>Connexion SAML...</title></head><body>Connexion en cours...<script>
try {
  localStorage.setItem('jwt_token', %q);
  localStorage.setItem('jwt_exp', %q);
  window.location.replace('/web');
} catch(e) {
  window.location.replace('/web/login');
}
</script></body></html>`, signed, fmt.Sprintf("%d", expiresAt))
		   })))
			webMux.HandleFunc("/saml/login", func(w http.ResponseWriter, r *http.Request) {
			   log.Printf("[SAML] /saml/login hit: %s %s", r.Method, r.RemoteAddr)
			   samlMiddleware.ServeHTTP(w, r)
		   })
		   webMux.HandleFunc("/saml/acs", func(w http.ResponseWriter, r *http.Request) {
			   log.Printf("[SAML] /saml/acs hit: %s %s", r.Method, r.RemoteAddr)
			   samlMiddleware.ServeHTTP(w, r)
		   })
		   webMux.HandleFunc("/saml/metadata", func(w http.ResponseWriter, r *http.Request) {
			   log.Printf("[SAML] /saml/metadata hit: %s %s", r.Method, r.RemoteAddr)
			   samlMiddleware.ServeHTTP(w, r)
		   })
	   }
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