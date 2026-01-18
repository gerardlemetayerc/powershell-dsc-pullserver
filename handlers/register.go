package handlers

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"
	"go-dsc-pull/utils"
	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal/schema"
	"os"
	"io"
)

// RegisterHandler gère l'enregistrement initial (POST /PSDSCPullServer.svc/Nodes)
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	body, _ := io.ReadAll(r.Body)
	var req map[string]interface{}
	_ = json.Unmarshal(body, &req) // Body peut être vide en PUT
	// Extraire l'AgentId du segment d'URL si présent
	raw := r.PathValue("node")
	agentId := utils.ExtractAgentId(raw)
	if agentId == "" {
		agentId = generateAgentId()
	}
	// Stocker tout le body JSON reçu dans un fichier agents/AgentId.json
	_ = os.MkdirAll("agents", 0755)
	_ = os.WriteFile("agents/"+agentId+".json", body, 0644)

	// --- Insertion en base ---
	// Charger la config DB
	dbCfg, err := db.LoadDBConfig("config.json")
	if err != nil {
		log.Printf("[REGISTER][DB] Erreur chargement config DB: %v", err)
		// On continue, mais pas d'insertion DB
	} else {
		database, err := db.OpenDB(dbCfg)
		if err != nil {
			log.Printf("[REGISTER][DB] Erreur ouverture DB: %v", err)
		} else {
			defer database.Close()
			// Insertion agent principal
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
			_, err = database.Exec(`INSERT OR REPLACE INTO agents (agent_id, node_name, lcm_version, registration_type, certificate_thumbprint, certificate_subject, certificate_issuer, certificate_notbefore, certificate_notafter) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				agentId, nodeName, lcmVersion, registrationType, thumbprint, subject, issuer, notbefore, notafter)
			if err != nil {
				log.Printf("[REGISTER][DB] Erreur insertion agent: %v", err)
			}
			   // Supprimer toutes les configurations existantes pour cet agent
			   _, err := database.Exec(`DELETE FROM agent_configurations WHERE agent_id = ?`, agentId)
			   if err != nil {
				   log.Printf("[REGISTER][DB] Erreur suppression configs existantes: %v", err)
			   }
			   // Insertion des nouveaux noms de configuration
			   if configNames, ok := req["ConfigurationNames"]; ok {
				   switch vv := configNames.(type) {
				   case []interface{}:
					   for _, n := range vv {
						   if s, ok := n.(string); ok {
							   _, err := database.Exec(`INSERT OR REPLACE INTO agent_configurations (agent_id, configuration_name) VALUES (?, ?)`, agentId, s)
							   if err != nil {
								   log.Printf("[REGISTER][DB] Erreur insertion config: %v", err)
							   }
						   }
					   }
				   case string:
					   _, err := database.Exec(`INSERT OR REPLACE INTO agent_configurations (agent_id, configuration_name) VALUES (?, ?)`, agentId, vv)
					   if err != nil {
						   log.Printf("[REGISTER][DB] Erreur insertion config: %v", err)
					   }
				   }
			   }
		}
	}

	resp := schema.RegisterResponse{AgentId: agentId}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ProtocolVersion", "2.0")
	_ = json.NewEncoder(w).Encode(resp)
	log.Printf("[REGISTER] Infos=%v AgentId=%s", req, agentId)
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
