
package main

import(
	"log"
	"strings"
	"os"
	"fmt"
	"encoding/json"
	"net/http"
	"go-dsc-pull/handlers"
	"go-dsc-pull/utils"
	"go-dsc-pull/internal/db"
    "time"
    "github.com/golang-jwt/jwt/v5"
)

type RegisterRequest struct {
	// À adapter selon la spec, ici minimal
	NodeName string `json:"NodeName"`
}

type RegisterResponse struct {
	AgentId string `json:"AgentId"`
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
	// Initialisation automatique de la base (CREATE IF NOT EXISTS)
	dbCfg, err := db.LoadDBConfig("config.json")
	if err != nil {
		log.Fatalf("[INITDB] Erreur chargement config DB: %v", err)
	}
	db.InitDB(dbCfg)
	dbConn, err := db.OpenDB(dbCfg)
	if err != nil {
		log.Fatalf("Erreur ouverture DB: %v", err)
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
	webMux := http.NewServeMux()
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

	// Endpoint de login pour JWT
	webMux.HandleFunc("POST /api/v1/login", func(w http.ResponseWriter, r *http.Request) {
		type LoginRequest struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		type LoginResponse struct {
			Token string `json:"token"`
			ExpiresAt int64 `json:"expires_at"`
		}
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		// Auth simplifiée : à remplacer par vérif DB/LDAP
		if req.Username != "admin" || req.Password != "password" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		// Générer JWT
		secret := []byte("supersecretkey") // À stocker ailleurs en prod
		expiresAt := time.Now().Add(60 * time.Minute).Unix()
		claims := jwt.MapClaims{
			"sub": req.Username,
			"exp": expiresAt,
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString(secret)
		if err != nil {
			http.Error(w, "Token error", http.StatusInternalServerError)
			return
		}
		resp := LoginResponse{Token: signed, ExpiresAt: expiresAt}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})


	// Handler pour /web/login
	webMux.HandleFunc("/web/login", func(w http.ResponseWriter, r *http.Request) {
		// Vérifie la présence d'un token JWT dans le cookie (optionnel, car le JS stocke dans localStorage)
		// Ici, on laisse toujours afficher la page login, la redirection sera gérée côté JS après login
		http.ServeFile(w, r, "web/login.tmpl")
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