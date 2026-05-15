package handlers

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"
	"fmt"
	"strings"
	utils "go-dsc-pull/utils"
	internalutils "go-dsc-pull/internal/utils"
	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal/schema"
	"go-dsc-pull/internal/global"
	"go-dsc-pull/internal/logs"
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
		logs.WriteLogFile(fmt.Sprintf("[REGISTER][CONFIG] registrationKey missing in config file, server stopped."))
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
				   logs.WriteLogFile(fmt.Sprintf("[ERROR][REGISTER][DB] Erreur ouverture DB: %v", err))
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
						   logs.WriteLogFile(fmt.Sprintf("[ERROR][REGISTER][DB] Erreur update agent TEMP: %v", err))
					   }
					   // Met à jour l'agent_id dans la table agent_tags
					   _, err = database.Exec(`UPDATE agent_tags SET agent_id = ? WHERE agent_id = ?`, agentId, tempAgentId)
					   if err != nil {
						   logs.WriteLogFile(fmt.Sprintf("[ERROR][REGISTER][DB] Erreur update agent_tags TEMP: %v", err))
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
							  logs.WriteLogFile(fmt.Sprintf("[ERROR][REGISTER][DB] Erreur insertion agent: %v", err))
						  }
				   }

				   // Only proceed with configuration insertion if agent insert/update succeeded
				   if err == nil {
					   incomingConfigs := extractConfigurationNamesFromRegisterPayload(req)

					   previousMains := make([]string, 0)
					   if driver == "sqlite" {
						   rows, qErr := database.Query(`
							   SELECT configuration_name
							   FROM agent_configurations
							   WHERE agent_id = ?
							     AND LOWER(COALESCE(NULLIF(TRIM(schedule_type), ''), 'none')) = 'none'
							   ORDER BY configuration_name
						   `, agentId)
						   if qErr == nil {
							   for rows.Next() {
								   var cfg string
								   if scanErr := rows.Scan(&cfg); scanErr == nil {
									   previousMains = append(previousMains, cfg)
								   }
							   }
							   rows.Close()
						   }
					   } else {
						   rows, qErr := database.Query(`
							   SELECT configuration_name
							   FROM agent_configurations
							   WHERE agent_id = ?
							     AND LOWER(COALESCE(NULLIF(LTRIM(RTRIM(schedule_type)), ''), 'none')) = 'none'
							   ORDER BY configuration_name
						   `, agentId)
						   if qErr == nil {
							   for rows.Next() {
								   var cfg string
								   if scanErr := rows.Scan(&cfg); scanErr == nil {
									   previousMains = append(previousMains, cfg)
								   }
							   }
							   rows.Close()
						   }
					   }

					   mainConfigs := make([]string, 0)
					   if len(incomingConfigs) > 0 {
						   mainConfigs = incomingConfigs
					   } else {
						   mainConfigs = previousMains
					   }

					   if len(incomingConfigs) == 0 {
						   logs.WriteLogFile(fmt.Sprintf("[WARN][REGISTER][DB] Aucune ConfigurationName detectee dans le payload register pour agentId=%s", agentId))
					   }

					   if len(mainConfigs) == 0 {
						   logs.WriteLogFile(fmt.Sprintf("[WARN][REGISTER][DB] Aucune configuration main resolue pour agentId=%s, conservation des liens existants", agentId))
					   } else {
						   // Reset des liens seulement quand une cible explicite est disponible.
						   _, err := database.Exec(`DELETE FROM agent_configurations WHERE agent_id = ?`, agentId)
						   if err != nil {
							   logs.WriteLogFile(fmt.Sprintf("[ERROR][REGISTER][DB] Erreur suppression configs existantes: %v", err))
						   }

						   for _, mainConfig := range mainConfigs {
							   if driver == "sqlite" {
								   logs.WriteLogFile(fmt.Sprintf("[INFO][REGISTER][DB] Insertion config main (SQLite): agentId=%s, config=%s", agentId, mainConfig))
								   _, err := database.Exec(`
									   INSERT OR REPLACE INTO agent_configurations (agent_id, configuration_name, schedule_type, enabled)
									   VALUES (?, ?, 'none', 1)
								   `, agentId, mainConfig)
								   if err != nil {
									   logs.WriteLogFile(fmt.Sprintf("[ERROR][REGISTER][DB] Erreur insertion config main: %v (agentId=%s, config=%s)", err, agentId, mainConfig))
								   }
							   } else {
								   logs.WriteLogFile(fmt.Sprintf("[INFO][REGISTER][DB] Insertion config main (MSSQL): agentId=%s, config=%s", agentId, mainConfig))
								   _, err := database.Exec(`
									   INSERT INTO agent_configurations (agent_id, configuration_name, schedule_type, enabled)
									   VALUES (?, ?, 'none', 1)
								   `, agentId, mainConfig)
								   if err != nil {
									   logs.WriteLogFile(fmt.Sprintf("[ERROR][REGISTER][DB] Erreur insertion config main: %v (agentId=%s, config=%s)", err, agentId, mainConfig))
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

func appendUniqueConfigName(raw string, seen map[string]bool, names *[]string) {
	s := strings.TrimSpace(raw)
	if s == "" || s == "<nil>" {
		return
	}
	key := strings.ToLower(s)
	if seen[key] {
		return
	}
	seen[key] = true
	*names = append(*names, s)
}

func extractConfigurationNamesFromRegisterPayload(payload map[string]interface{}) []string {
	seen := make(map[string]bool)
	names := make([]string, 0)
	collectConfigurationNamesFromValue(payload, seen, &names)
	return names
}

func collectConfigurationNamesFromValue(value interface{}, seen map[string]bool, names *[]string) {
	switch vv := value.(type) {
	case map[string]interface{}:
		for k, nested := range vv {
			trimmedKey := strings.TrimSpace(k)
			switch {
			case strings.EqualFold(trimmedKey, "ConfigurationNames"):
				switch cfgs := nested.(type) {
				case []interface{}:
					for _, item := range cfgs {
						if s, ok := item.(string); ok {
							appendUniqueConfigName(s, seen, names)
						}
					}
				case string:
					appendUniqueConfigName(cfgs, seen, names)
				}
			case strings.EqualFold(trimmedKey, "ConfigurationName"):
				if s, ok := nested.(string); ok {
					appendUniqueConfigName(s, seen, names)
				}
			case strings.EqualFold(trimmedKey, "PartialConfigurations"):
				if partials, ok := nested.([]interface{}); ok {
					for _, p := range partials {
						if pm, ok := p.(map[string]interface{}); ok {
							if d, ok := pm["Description"].(string); ok {
								appendUniqueConfigName(d, seen, names)
							}
							if cn, ok := pm["ConfigurationName"].(string); ok {
								appendUniqueConfigName(cn, seen, names)
							}
						}
					}
				}
			}
			collectConfigurationNamesFromValue(nested, seen, names)
		}
	case []interface{}:
		for _, item := range vv {
			collectConfigurationNamesFromValue(item, seen, names)
		}
	}
}
