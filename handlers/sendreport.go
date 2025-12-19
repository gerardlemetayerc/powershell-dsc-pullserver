package handlers

import (
	"io"
	"log"
	"net/http"
	"os"
	"fmt"
	"time"
)

// SendReportHandler gère POST /PSDSCPullServer.svc/Nodes(AgentId='...')/SendReport
func SendReportHandler(w http.ResponseWriter, r *http.Request) {
	       reportBody, err := io.ReadAll(r.Body)
	       if err != nil {
		       log.Printf("[SENDREPORT] Erreur lecture body: %v", err)
		       w.WriteHeader(http.StatusBadRequest)
		       return
	       }
	       agentId := r.PathValue("node")
	       log.Printf("[SENDREPORT] AgentId=%s, ReportSize=%d", agentId, len(reportBody))
	       // Sauvegarde le rapport dans /reports/AgentId-timestamp.json
	       ts := time.Now().Format("20060102-150405")
	       filename := fmt.Sprintf("reports/%s-%s.json", agentId, ts)
	       if err := os.WriteFile(filename, reportBody, 0644); err != nil {
		       log.Printf("[SENDREPORT] Erreur écriture fichier: %v", err)
	       }
	       w.WriteHeader(http.StatusOK)
	       w.Write([]byte("{}"))
}
