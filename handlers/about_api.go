package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"go-dsc-pull/internal/buildinfo"
	internaldb "go-dsc-pull/internal/db"
	"go-dsc-pull/internal/global"
	"go-dsc-pull/internal/logs"
	"go-dsc-pull/internal/service"
)

type aboutResponse struct {
	BuildVersion     string `json:"build_version"`
	DBVersion        string `json:"db_version,omitempty"`
	LatestRelease    string `json:"latest_release,omitempty"`
	LatestReleaseURL string `json:"latest_release_url,omitempty"`
	UpdateAvailable  bool   `json:"update_available"`
	ReleaseCheckOK   bool   `json:"release_check_ok"`
}

// AboutAPIHandler returns application version details from build info and database metadata.
func AboutAPIHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := aboutResponse{BuildVersion: buildinfo.Version}
		driverName := "sqlite"
		if global.AppConfig != nil && global.AppConfig.Database.Driver != "" {
			driverName = global.AppConfig.Database.Driver
		}

		var dbVersion sql.NullString
		err := db.QueryRow("SELECT db_version FROM dsc_infra_info WHERE id = 1").Scan(&dbVersion)
		if err != nil && err != sql.ErrNoRows {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if dbVersion.Valid {
			response.DBVersion = dbVersion.String
		}

		storedState, storedErr := internaldb.GetReleaseCheckState(db)
		if storedErr == nil {
			if storedState.LatestRelease != "" {
				response.LatestRelease = storedState.LatestRelease
			}
			if storedState.LatestReleaseURL != "" {
				response.LatestReleaseURL = storedState.LatestReleaseURL
			}
			response.UpdateAvailable = storedState.UpdateAvailable
			response.ReleaseCheckOK = storedState.ReleaseCheckOK
		}

		releaseInfo, releaseErr := service.CheckLatestRelease(buildinfo.Version)
		if releaseErr == nil && releaseInfo.LatestRelease != "" {
			response.ReleaseCheckOK = true
			response.LatestRelease = releaseInfo.LatestRelease
			response.LatestReleaseURL = releaseInfo.LatestReleaseURL
			response.UpdateAvailable = releaseInfo.UpdateAvailable
			_ = internaldb.PersistReleaseCheckSuccess(db, driverName, releaseInfo.LatestRelease, releaseInfo.LatestReleaseURL, releaseInfo.UpdateAvailable)
			logs.WriteLogFile("INFO [ABOUT] GitHub latest release check OK: latest=" + releaseInfo.LatestRelease + ", current=" + buildinfo.Version)
		} else {
			_ = internaldb.PersistReleaseCheckFailure(db, driverName)
			if releaseErr != nil {
				logs.WriteLogFile("WARN [ABOUT] GitHub latest release check failed: " + releaseErr.Error())
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}
}