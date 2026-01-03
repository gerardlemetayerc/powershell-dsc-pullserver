package handlers

import (
	"database/sql"
	"net/http"
)

// Handler DSC: fournit le module DSC (zip) si nom, version, checksum correspondent
func ModuleDownloadHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("ModuleName")
		version := r.URL.Query().Get("ModuleVersion")
		checksum := r.URL.Query().Get("Checksum")
		if name == "" || version == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Missing ModuleName or ModuleVersion"))
			return
		}
		// On stocke le checksum en base64, mais certains clients l'envoient en minuscule ou sans padding
		// On tolère les variantes
		var zipBlob []byte
		var dbChecksum string
		err := db.QueryRow(`SELECT zip_blob, checksum FROM modules WHERE name = ? AND version = ?`, name, version).Scan(&zipBlob, &dbChecksum)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Module not found"))
			return
		}
		if checksum != "" {
			if checksum != dbChecksum {
				// Tolérance : padding base64
				if len(checksum) < len(dbChecksum) {
					pad := len(dbChecksum) - len(checksum)
					checksum += string(make([]byte, pad))
				}
				if checksum != dbChecksum {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("Checksum mismatch"))
					return
				}
			}
		}
		w.Header().Set("Checksum", dbChecksum)
		w.Header().Set("ChecksumAlgorithm", "SHA-256")
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+name+"_"+version+".zip\"")
		w.Header().Set("ProtocolVersion", "2.0")
		w.WriteHeader(http.StatusOK)
		w.Write(zipBlob)
	}
}
