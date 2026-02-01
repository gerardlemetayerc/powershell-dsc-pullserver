package routes

import (
	"net/http"
	"database/sql"
	"go-dsc-pull/handlers"
	"path/filepath"
	"go-dsc-pull/utils"
	samlsp "github.com/crewjam/saml/samlsp"
)

// RegisterWebRoutes sets up all web/API endpoints on the provided mux
func RegisterWebRoutes(mux *http.ServeMux, dbConn *sql.DB, jwtAuthMiddleware func(http.Handler) http.Handler, samlMiddleware http.Handler) {
		// Endpoint pour la liste des profils utilisateurs disponibles
		mux.Handle("GET /api/v1/user_roles", jwtAuthMiddleware(http.HandlerFunc(handlers.UserRolesHandler())))
	exeDir, err := utils.ExePath()
	if err != nil {
		panic("Failed to get executable path: " + err.Error())
	}
	baseDir := filepath.Dir(exeDir)
	mux.Handle("GET /api/v1/my", jwtAuthMiddleware(http.HandlerFunc(handlers.MyUserInfoHandler(dbConn))))
	mux.Handle("/web/profile", handlers.WebJWTAuthMiddleware(http.HandlerFunc(handlers.ProfileHandler)))
	// API tokens utilisateur
	mux.Handle("GET /api/v1/users/{id}/tokens", jwtAuthMiddleware(http.HandlerFunc(handlers.ListUserAPITokensHandler(dbConn))))
	mux.Handle("POST /api/v1/users/{id}/tokens", jwtAuthMiddleware(http.HandlerFunc(handlers.CreateUserAPITokenHandler(dbConn))))
	mux.Handle("POST /api/v1/users/{id}/tokens/{tokenid}/revoke", jwtAuthMiddleware(http.HandlerFunc(handlers.RevokeUserAPITokenHandler(dbConn))))
	mux.Handle("DELETE /api/v1/users/{id}/tokens/{tokenid}", jwtAuthMiddleware(http.HandlerFunc(handlers.DeleteUserAPITokenHandler(dbConn))))
	// SAML endpoints
	mux.Handle("GET /api/v1/saml/userinfo", http.HandlerFunc(handlers.SAMLUserInfoHandler))
	mux.Handle("GET /api/v1/saml/enabled", http.HandlerFunc(handlers.SAMLStatusHandler))
	// SAML config management (admin only)
	mux.Handle("GET /api/v1/saml/config", jwtAuthMiddleware(http.HandlerFunc(handlers.SAMLConfigAPIHandler)))
	mux.Handle("PUT /api/v1/saml/config", jwtAuthMiddleware(http.HandlerFunc(handlers.SAMLConfigAPIHandler)))
	mux.Handle("POST /api/v1/saml/upload_sp_keycert", jwtAuthMiddleware(http.HandlerFunc(handlers.SAMLUploadSPKeyCertHandler)))
	// Web UI for SAML config (admin only)
	mux.Handle("GET /web/saml_config", handlers.WebJWTAuthMiddleware(http.HandlerFunc(handlers.WebSAMLConfigHandler)))
	if samlMiddleware != nil {
		mux.Handle("/saml/", samlMiddleware)
	}
	// API REST endpoints (agents, configs, reports, modules, properties, configuration_models, users, login)
	mux.Handle("GET /api/v1/agents", jwtAuthMiddleware(http.HandlerFunc(handlers.AgentAPIHandler)))
	mux.Handle("POST /api/v1/agents/preenroll", jwtAuthMiddleware(http.HandlerFunc(handlers.PreEnrollAgentHandler)))
	mux.Handle("GET /api/v1/agents/{id}/configs", jwtAuthMiddleware(http.HandlerFunc(handlers.AgentConfigsAPIHandler)))
	mux.Handle("POST /api/v1/agents/{id}/configs", jwtAuthMiddleware(http.HandlerFunc(handlers.AgentConfigsAPIHandlerPostDelete)))
	mux.Handle("DELETE /api/v1/agents/{id}/configs", jwtAuthMiddleware(http.HandlerFunc(handlers.AgentConfigsAPIHandlerPostDelete)))
	mux.Handle("GET /api/v1/agents/{id}/reports", jwtAuthMiddleware(http.HandlerFunc(handlers.AgentReportsListHandler)))
	mux.Handle("GET /api/v1/agents/{id}/reports/latest", jwtAuthMiddleware(http.HandlerFunc(handlers.AgentReportsLatestHandler)))
	mux.Handle("GET /api/v1/agents/{id}/reports/{jobid}", jwtAuthMiddleware(http.HandlerFunc(handlers.AgentReportsByJobIdHandler)))
	mux.Handle("GET /api/v1/agents/{id}", jwtAuthMiddleware(http.HandlerFunc(handlers.AgentByIdAPIHandler)))
	mux.Handle("POST /api/v1/modules/upload", jwtAuthMiddleware(http.HandlerFunc(handlers.ModuleUploadHandler(dbConn))))
	mux.Handle("GET /api/v1/modules", jwtAuthMiddleware(http.HandlerFunc(handlers.ModuleListHandler(dbConn))))
	mux.Handle("DELETE /api/v1/modules/delete", jwtAuthMiddleware(http.HandlerFunc(handlers.ModuleDeleteHandler(dbConn))))
	mux.Handle("GET /api/v1/properties", jwtAuthMiddleware(http.HandlerFunc(handlers.PropertiesListHandler)))
	mux.Handle("POST /api/v1/properties", jwtAuthMiddleware(http.HandlerFunc(handlers.PropertiesCreateHandler)))
	mux.Handle("GET /api/v1/properties/{id}", jwtAuthMiddleware(http.HandlerFunc(handlers.PropertiesGetHandler)))
	mux.Handle("PUT /api/v1/properties/{id}", jwtAuthMiddleware(http.HandlerFunc(handlers.PropertiesUpdateHandler)))
	mux.Handle("DELETE /api/v1/properties/{id}", jwtAuthMiddleware(http.HandlerFunc(handlers.PropertiesDeleteHandler)))
	mux.Handle("POST /api/v1/configuration_models", jwtAuthMiddleware(http.HandlerFunc(handlers.CreateConfigurationModelHandler)))
	mux.Handle("GET /api/v1/configuration_models", jwtAuthMiddleware(http.HandlerFunc(handlers.ListConfigurationModelsHandler)))
	mux.Handle("GET /api/v1/configuration_models/{id}", jwtAuthMiddleware(http.HandlerFunc(handlers.GetConfigurationModelHandler)))
	mux.Handle("PUT /api/v1/configuration_models/{id}", jwtAuthMiddleware(http.HandlerFunc(handlers.UpdateConfigurationModelHandler)))
	mux.Handle("DELETE /api/v1/configuration_models", jwtAuthMiddleware(http.HandlerFunc(handlers.DeleteConfigurationModelHandler)))
	mux.Handle("POST /api/v1/login", handlers.LoginHandler(dbConn))
	mux.HandleFunc("/web/login", WebLoginHandler)
	if smw, ok := samlMiddleware.(*samlsp.Middleware); ok {
		mux.Handle("/web/login/saml", smw.RequireAccount(handlers.SAMLLoginHandler(dbConn)))
		mux.HandleFunc("/saml/login", func(w http.ResponseWriter, r *http.Request) {smw.ServeHTTP(w, r)})
		mux.HandleFunc("/saml/acs", func(w http.ResponseWriter, r *http.Request) {smw.ServeHTTP(w, r)})
		mux.HandleFunc("/saml/metadata", func(w http.ResponseWriter, r *http.Request) {smw.ServeHTTP(w, r)})
	}
	
	// Web GUI endpoints (login, index, static, node, modules, configuration_model, properties, users, user_edit, user_password)
	mux.Handle("/web", handlers.WebJWTAuthMiddleware(http.HandlerFunc(handlers.WebIndexHandler)))
	// Custom static file handler to prevent directory listing under /web/
	mux.Handle("/web/", http.StripPrefix("/web/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "" || path == "/" {
			handlers.WebJWTAuthMiddleware(http.HandlerFunc(handlers.WebIndexHandler)).ServeHTTP(w, r)
			return
		}
		fullPath := filepath.Join(baseDir, "web", path)
		info, err := handlers.StatFile(fullPath)
		if err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, fullPath)
	})))
	mux.Handle("/web/node/", handlers.WebJWTAuthMiddleware(http.HandlerFunc(handlers.WebNodeHandler)))
	mux.Handle("/web/modules", handlers.WebJWTAuthMiddleware(http.HandlerFunc(handlers.WebModulesHandler)))
	mux.Handle("/web/configuration_model", handlers.WebJWTAuthMiddleware(http.HandlerFunc(handlers.WebConfigurationModelHandler)))
	mux.Handle("/templates/properties.tmpl", handlers.WebJWTAuthMiddleware(http.HandlerFunc(handlers.WebPropertiesHandler)))
	mux.Handle("/web/users", handlers.WebJWTAuthMiddleware(http.HandlerFunc(WebUsersHandler)))
	mux.Handle("/web/user_edit", handlers.WebJWTAuthMiddleware(http.HandlerFunc(WebUserEditHandler)))
	mux.Handle("/web/user_password", handlers.WebJWTAuthMiddleware(http.HandlerFunc(WebUserPasswordHandler)))
	mux.Handle("POST /api/v1/users/{id}/password", jwtAuthMiddleware(http.HandlerFunc(handlers.ChangeUserPasswordHandler(dbConn))))
	mux.Handle("GET /api/v1/users", jwtAuthMiddleware(http.HandlerFunc(handlers.ListUsersHandler(dbConn))))
	mux.Handle("GET /api/v1/users/{id}", jwtAuthMiddleware(http.HandlerFunc(handlers.GetUserHandler(dbConn))))
	mux.Handle("POST /api/v1/users", jwtAuthMiddleware(http.HandlerFunc(handlers.CreateUserHandler(dbConn))))
	mux.Handle("PUT /api/v1/users/{id}", jwtAuthMiddleware(http.HandlerFunc(handlers.UpdateUserHandler(dbConn))))
	mux.Handle("DELETE /api/v1/users/{id}", jwtAuthMiddleware(http.HandlerFunc(handlers.DeleteUserHandler(dbConn))))
	mux.Handle("POST /api/v1/users/{id}/active", jwtAuthMiddleware(http.HandlerFunc(handlers.SetUserActiveHandler(dbConn))))
	// API tags agents
	mux.Handle("GET /api/v1/agents/{id}/tags", jwtAuthMiddleware(http.HandlerFunc(handlers.AgentTagsListHandler)))
	mux.Handle("PUT /api/v1/agents/{id}/tags", jwtAuthMiddleware(http.HandlerFunc(handlers.AgentTagsSetHandler)))
	mux.Handle("DELETE /api/v1/agents/{id}/tags", jwtAuthMiddleware(http.HandlerFunc(handlers.AgentTagsDeleteHandler)))

}