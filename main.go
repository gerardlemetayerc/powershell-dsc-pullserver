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
	webMux.HandleFunc("GET /api/v1/agents", handlers.AgentAPIHandler)
	// API REST: configurations d'un agent
	webMux.HandleFunc("GET /api/v1/agents/{id}/configs", handlers.AgentConfigsAPIHandler)
	webMux.HandleFunc("POST /api/v1/agents/{id}/configs", handlers.AgentConfigsAPIHandlerPostDelete)
	webMux.HandleFunc("DELETE /api/v1/agents/{id}/configs", handlers.AgentConfigsAPIHandlerPostDelete)
	// API REST: rapports d'un agent
	webMux.HandleFunc("GET /api/v1/agents/{id}/reports", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[API] /api/v1/agents/%s/reports?%s", r.PathValue("id"), r.URL.RawQuery)
		handlers.AgentReportsListHandler(w, r)
	})
	webMux.HandleFunc("GET /api/v1/agents/{id}/reports/latest", handlers.AgentReportsLatestHandler)
	webMux.HandleFunc("GET /api/v1/agents/{id}/reports/{jobid}", handlers.AgentReportsByJobIdHandler)
	// API REST: infos d'un agent
	webMux.HandleFunc("GET /api/v1/agents/{id}", handlers.AgentByIdAPIHandler)
	// API REST: modules DSC
	webMux.HandleFunc("POST /api/v1/modules/upload", handlers.ModuleUploadHandler(dbConn))
	webMux.HandleFunc("GET /api/v1/modules", handlers.ModuleListHandler(dbConn))
	webMux.HandleFunc("DELETE /api/v1/modules/delete", handlers.ModuleDeleteHandler(dbConn))

	// API REST: properties
	webMux.HandleFunc("GET /api/v1/properties", handlers.PropertiesListHandler)
	webMux.HandleFunc("POST /api/v1/properties", handlers.PropertiesCreateHandler)
	webMux.HandleFunc("GET /api/v1/properties/{id}", handlers.PropertiesGetHandler)
	webMux.HandleFunc("PUT /api/v1/properties/{id}", handlers.PropertiesUpdateHandler)
	webMux.HandleFunc("DELETE /api/v1/properties/{id}", handlers.PropertiesDeleteHandler)

	// API REST: node_properties
	webMux.HandleFunc("GET /api/v1/agents/{nodename}/properties", handlers.NodePropertiesListHandler)
	webMux.HandleFunc("POST /api/v1/agents/{nodename}/properties", handlers.NodePropertiesCreateHandler)
	webMux.HandleFunc("GET /api/v1/agents/{nodename}/properties/{property_id}", handlers.NodePropertyGetHandler)
	webMux.HandleFunc("PUT /api/v1/agents/{nodename}/properties/{property_id}", handlers.NodePropertyUpdateHandler)
	webMux.HandleFunc("DELETE /api/v1/agents/{nodename}/properties/{property_id}", handlers.NodePropertyDeleteHandler)

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
