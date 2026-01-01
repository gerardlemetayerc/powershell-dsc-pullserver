
package main

import(
	"log"
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

	// --- Mux IHM/API ---
	webMux := http.NewServeMux()
	// API REST: liste des agents
	webMux.HandleFunc("GET /api/v1/agents", handlers.AgentAPIHandler)
	// API REST: configurations d'un agent
	webMux.HandleFunc("GET /api/v1/agents/{id}/configs", handlers.AgentConfigsAPIHandler)
	webMux.HandleFunc("POST /api/v1/agents/{id}/configs", handlers.AgentConfigsAPIHandlerPostDelete)
	webMux.HandleFunc("DELETE /api/v1/agents/{id}/configs", handlers.AgentConfigsAPIHandlerPostDelete)
	// API REST: rapports d'un agent
	webMux.HandleFunc("GET /api/v1/agents/{id}/reports", handlers.AgentReportsListHandler)
	webMux.HandleFunc("GET /api/v1/agents/{id}/reports/latest", handlers.AgentReportsLatestHandler)
	webMux.HandleFunc("GET /api/v1/agents/{id}/reports/{jobid}", handlers.AgentReportsByJobIdHandler)
	// API REST: infos d'un agent
	webMux.HandleFunc("GET /api/v1/agents/{id}", handlers.AgentByIdAPIHandler)
	// Servir l'IHM web statique sur /web/
	webMux.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))
	// Special handler for /web/node/{id} to always serve node.html (SPA style)
	webMux.HandleFunc("/web/node/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/node.html")
	})

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
