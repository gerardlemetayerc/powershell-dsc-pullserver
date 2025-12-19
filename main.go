
package main

import(
	"log"
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

		       mux := http.NewServeMux()
	       // Handler pour l'enregistrement initial
	       mux.HandleFunc("PUT /PSDSCPullServer.svc/{node}", handlers.RegisterHandler)
	       // Handler pour GET /PSDSCPullServer.svc/{node}/{config}/ConfigurationContent
	       mux.HandleFunc("GET /PSDSCPullServer.svc/{node}/{config}/ConfigurationContent", handlers.ConfigurationContentHandler)
	       // Handler pour POST /PSDSCPullServer.svc/{node}/SendReport
	       mux.HandleFunc("POST /PSDSCPullServer.svc/{node}/SendReport", handlers.SendReportHandler)
	       // Handler pour GetDscAction dynamique (segment id = (AgentId='...'))
	       mux.HandleFunc("POST /PSDSCPullServer.svc/{id}/GetDscAction", func(w http.ResponseWriter, r *http.Request) {
		       raw := r.PathValue("id")
		       agentId := utils.ExtractAgentId(raw)
		       handlers.GetDscActionNodeHandlerWithId(w, r, agentId)
	       })
		// API REST: liste des agents
		mux.HandleFunc("GET /api/v1/agents", handlers.AgentAPIHandler)
			   // API REST: configurations d'un agent
			   mux.HandleFunc("GET /api/v1/agents/{id}/configs", handlers.AgentConfigsAPIHandler)
			   mux.HandleFunc("POST /api/v1/agents/{id}/configs", handlers.AgentConfigsAPIHandlerPostDelete)
			   mux.HandleFunc("DELETE /api/v1/agents/{id}/configs", handlers.AgentConfigsAPIHandlerPostDelete)
		// API REST: infos d'un agent
		mux.HandleFunc("GET /api/v1/agents/{id}", handlers.AgentByIdAPIHandler)
		// Servir l'IHM web statique sur /web/
		mux.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))

		// Wrap mux with logging middleware
		handler := loggingMiddleware(mux)
		log.Println("Serveur DSC Pull - endpoint d'enregistrement sur :8080 ... (IHM sur /web/)")
		log.Fatal(http.ListenAndServe(":8080", handler))
}
