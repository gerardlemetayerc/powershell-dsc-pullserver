package handlers

import (
	"log"
	"net/http"
	"os"
	"fmt"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"go-dsc-pull/utils"
)

// ConfigurationContentHandler gère GET /PSDSCPullServer.svc/{node}/{config}/ConfigurationContent
func ConfigurationContentHandler(w http.ResponseWriter, r *http.Request) {
	nodeRaw := r.PathValue("node")
	configRaw := r.PathValue("config")
	agentId := utils.ExtractAgentId(nodeRaw)
	configName := utils.ExtractConfigName(configRaw)
	       log.Printf("[CONFIGURATIONCONTENT] AgentId=%s ConfigName=%s", agentId, configName)
	       mofPath := "configs/" + configName + ".mof"
	       data, err := os.ReadFile(mofPath)
	       if err != nil {
		       http.Error(w, "Configuration not found", http.StatusNotFound)
		       return
	       }
	       // Calcul du checksum SHA256
	       hash := sha256.Sum256(data)
	       checksum := hex.EncodeToString(hash[:])
	       w.Header().Set("Content-Type", "application/octet-stream")
	       w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		   w.Header().Set("Checksum", strings.ToUpper(checksum))
	       w.Header().Set("ChecksumAlgorithm", "SHA-256")
	       //w.Header().Set("Content-Disposition", "attachment; filename=\"configuration.mof\"")
	       w.Header().Set("ProtocolVersion", "2.0")
	       w.WriteHeader(http.StatusOK)
	       w.Write(data)
}
