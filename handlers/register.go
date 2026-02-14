package handlers

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"
	utils "go-dsc-pull/utils"
	internalutils "go-dsc-pull/internal/utils"
	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal/schema"
	"go-dsc-pull/internal/global"
	"io"
)

// RegisterHandler gère l'enregistrement initial (POST /PSDSCPullServer.svc/Nodes)
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// --- Contrôle de la signature Authorization DSC ---
	authHeader := r.Header.Get("Authorization")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusInternalServerError)
		return
	}
	// Charger la config pour récupérer la clé d'enregistrement
	registrationKeyPlain := global.AppConfig.DSCPullServer.RegistrationKey
	if registrationKeyPlain == "" {
		log.Printf("[REGISTER][CONFIG] registrationKey missing in config file, server stopped.")
		http.Error(w, "Server configuration error: registrationKey missing", http.StatusInternalServerError)
		return
	}
	xmsDate := r.Header.Get("x-ms-date")
	valid, logMsg := internalutils.ValidateDSCRegistrationKey(body, xmsDate, authHeader, registrationKeyPlain)
	if !valid {
		log.Print(logMsg)
		http.Error(w, "Unauthorized: invalid signature", http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var req map[string]interface{}
	_ = json.Unmarshal(body, &req) // Body peut être vide en PUT
	// Extraire l'AgentId du segment d'URL si présent
	raw := r.PathValue("node")
	agentId := utils.ExtractAgentId(raw)
	if agentId == "" {
		agentId = generateAgentId()
	}

	// --- Insertion en base ---
	// Charger la config DB
	database, err := db.OpenDB(&global.AppConfig.Database)
	if err != nil {
		log.Printf("[REGISTER][DB] Erreur ouverture DB: %v", err)
		// On continue, mais pas d'insertion DB
	} else {
			   if err != nil {
				   log.Printf("[REGISTER][DB] Erreur ouverture DB: %v", err)
			   } else {
				   defer database.Close()
				   driver := global.AppConfig.Database.Driver
				   // Insertion ou mise à jour agent principal
				   agentInfo := req["AgentInformation"].(map[string]interface{})
				   nodeName := agentInfo["NodeName"].(string)
				   lcmVersion := agentInfo["LCMVersion"].(string)
				   regInfo := req["RegistrationInformation"].(map[string]interface{})
				   certInfo := regInfo["CertificateInformation"].(map[string]interface{})
				   registrationType := regInfo["RegistrationMessageType"].(string)
				   thumbprint := certInfo["Thumbprint"].(string)
				   subject := certInfo["Subject"].(string)
				   issuer := certInfo["Issuer"].(string)
				   notbefore := certInfo["NotBefore"].(string)
				   notafter := certInfo["NotAfter"].(string)

				   // Vérifie si NodeName existe avec un AgentId TEMP
				   var tempAgentId string
				   err = database.QueryRow(`SELECT agent_id FROM agents WHERE node_name = ? AND agent_id LIKE 'TEMP-%'`, nodeName).Scan(&tempAgentId)
				   if err == nil && tempAgentId != "" {
					   // Mise à jour : change l'agent_id et les infos
					   _, err = database.Exec(`UPDATE agents SET agent_id = ?, lcm_version = ?, registration_type = ?, certificate_thumbprint = ?, certificate_subject = ?, certificate_issuer = ?, certificate_notbefore = ?, certificate_notafter = ? WHERE agent_id = ?`,
						   agentId, lcmVersion, registrationType, thumbprint, subject, issuer, notbefore, notafter, tempAgentId)
					   if err != nil {
						   log.Printf("[REGISTER][DB] Erreur update agent TEMP: %v", err)
					   }
					   // Met à jour l'agent_id dans la table agent_tags
					   _, err = database.Exec(`UPDATE agent_tags SET agent_id = ? WHERE agent_id = ?`, agentId, tempAgentId)
					   if err != nil {
						   log.Printf("[REGISTER][DB] Erreur update agent_tags TEMP: %v", err)
					   }
				   } else {
						  // Insertion normale, compatible SQLite/MSSQL
						  if driver == "sqlite" {
							  _, err = database.Exec(`INSERT OR REPLACE INTO agents (agent_id, node_name, lcm_version, registration_type, certificate_thumbprint, certificate_subject, certificate_issuer, certificate_notbefore, certificate_notafter, state) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
								  agentId, nodeName, lcmVersion, registrationType, thumbprint, subject, issuer, notbefore, notafter, "pending_apply")
						  } else {
							  _, err = database.Exec(`IF NOT EXISTS (SELECT 1 FROM agents WHERE agent_id = ?) INSERT INTO agents (agent_id, node_name, lcm_version, registration_type, certificate_thumbprint, certificate_subject, certificate_issuer, certificate_notbefore, certificate_notafter, registered_at, last_communication, has_error_last_report) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
								  agentId, agentId, nodeName, lcmVersion, registrationType, thumbprint, subject, issuer, notbefore, notafter, nil, nil, 0)
						  }
						  if err != nil {
							  log.Printf("[REGISTER][DB] Erreur insertion agent: %v", err)
						  }
				   }

				   // Only proceed with configuration insertion if agent insert/update succeeded
				   if err == nil {
					   // Supprimer toutes les configurations existantes pour cet agent
					   _, err := database.Exec(`DELETE FROM agent_configurations WHERE agent_id = ?`, agentId)
					   if err != nil {
						   log.Printf("[REGISTER][DB] Erreur suppression configs existantes: %v", err)
					   }
					   // Insertion des nouveaux noms de configuration (compatible SQLite/MSSQL)
					   if configNames, ok := req["ConfigurationNames"]; ok {
						   switch vv := configNames.(type) {
						   case []interface{}:
							   for _, n := range vv {
								   if s, ok := n.(string); ok {
									   if driver == "sqlite" {
										   _, err := database.Exec(`INSERT OR REPLACE INTO agent_configurations (agent_id, configuration_name) VALUES (?, ?)`, agentId, s)
										   if err != nil {
											   log.Printf("[REGISTER][DB] Erreur insertion config: %v (agentId=%s, config=%s)", err, agentId, s)
										   }
									   } else {
										   _, err := database.Exec(`IF NOT EXISTS (SELECT 1 FROM agent_configurations WHERE agent_id = ? AND configuration_name = ?) INSERT INTO agent_configurations (agent_id, configuration_name) VALUES (?, ?)`, agentId, s, agentId, s)
										   if err != nil {
											   log.Printf("[REGISTER][DB] Erreur insertion config: %v (agentId=%s, config=%s)", err, agentId, s)
										   }
									   }
								   }
							   }
						   case string:
							   if driver == "sqlite" {
								   _, err := database.Exec(`INSERT OR REPLACE INTO agent_configurations (agent_id, configuration_name) VALUES (?, ?)`, agentId, vv)
								   if err != nil {
									   log.Printf("[REGISTER][DB] Erreur insertion config: %v (agentId=%s, config=%s)", err, agentId, vv)
								   }
							   } else {
								   _, err := database.Exec(`IF NOT EXISTS (SELECT 1 FROM agent_configurations WHERE agent_id = ? AND configuration_name = ?) INSERT INTO agent_configurations (agent_id, configuration_name) VALUES (?, ?)`, agentId, vv, agentId, vv)
								   if err != nil {
									   log.Printf("[REGISTER][DB] Erreur insertion config: %v (agentId=%s, config=%s)", err, agentId, vv)
								   }
							   }
						   }
					   }
				   }
			   }
	}

	resp := schema.RegisterResponse{AgentId: agentId}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ProtocolVersion", "2.0")
	_ = json.NewEncoder(w).Encode(resp)
}

func generateAgentId() string {
	rand.Seed(time.Now().UnixNano())
	return randomHex(8) + "-" + randomHex(4) + "-" + randomHex(4) + "-" + randomHex(4) + "-" + randomHex(12)
}

func randomHex(n int) string {
	const letters = "0123456789ABCDEF"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
