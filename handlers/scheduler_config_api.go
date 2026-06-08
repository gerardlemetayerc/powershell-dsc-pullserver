package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"go-dsc-pull/internal/buildinfo"
	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal/global"
	"go-dsc-pull/internal/logs"
	"go-dsc-pull/internal/schema"
	"go-dsc-pull/internal/service"
	"go-dsc-pull/utils"
)

type SchedulerConfigDTO struct {
	EnableReportAutoCleanup bool `json:"enable_report_auto_cleanup"`
	ReportRetentionDays     int  `json:"report_retention_days"`
	ReportCleanupIntervalMins int `json:"report_cleanup_interval_mins"`
	EnableReleaseCheck bool `json:"enable_release_check"`
	ReleaseCheckIntervalMins int `json:"release_check_interval_mins"`
}

func resolveConfigPath() (string, error) {
	exeDir, err := utils.ExePath()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(exeDir), "config.json"), nil
}

func SchedulerConfigAPIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		cfg := global.AppConfig
		if cfg == nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to load config"})
			return
		}
		_ = json.NewEncoder(w).Encode(SchedulerConfigDTO{
			EnableReportAutoCleanup:  cfg.DSCPullServer.EnableReportAutoCleanup,
			ReportRetentionDays:      cfg.DSCPullServer.ReportRetentionDays,
			ReportCleanupIntervalMins: cfg.DSCPullServer.ReportCleanupIntervalMins,
			EnableReleaseCheck: cfg.DSCPullServer.EnableReleaseCheck,
			ReleaseCheckIntervalMins: cfg.DSCPullServer.ReleaseCheckIntervalMins,
		})

	case http.MethodPut:
		var req SchedulerConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
			return
		}
		if req.ReportRetentionDays <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "report_retention_days must be > 0"})
			return
		}
		if req.ReportCleanupIntervalMins <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "report_cleanup_interval_mins must be > 0"})
			return
		}
		if req.ReleaseCheckIntervalMins <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "release_check_interval_mins must be > 0"})
			return
		}

		configPath, err := resolveConfigPath()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to resolve config path"})
			return
		}

		fileData, err := os.ReadFile(configPath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to read config file"})
			return
		}

		var fullConfig map[string]interface{}
		if err := json.Unmarshal(fileData, &fullConfig); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to decode config"})
			return
		}

		dscPullRaw, ok := fullConfig["dsc_pullserver"].(map[string]interface{})
		if !ok {
			dscPullRaw = map[string]interface{}{}
		}
		dscPullRaw["enable_report_auto_cleanup"] = req.EnableReportAutoCleanup
		dscPullRaw["report_retention_days"] = req.ReportRetentionDays
		dscPullRaw["report_cleanup_interval_mins"] = req.ReportCleanupIntervalMins
		dscPullRaw["enable_release_check"] = req.EnableReleaseCheck
		dscPullRaw["release_check_interval_mins"] = req.ReleaseCheckIntervalMins
		fullConfig["dsc_pullserver"] = dscPullRaw

		data, err := json.MarshalIndent(fullConfig, "", "  ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to marshal config"})
			return
		}
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to write config"})
			return
		}

		// Hot-reload in-memory values for immediate API visibility.
		global.AppConfig.DSCPullServer.EnableReportAutoCleanup = req.EnableReportAutoCleanup
		global.AppConfig.DSCPullServer.ReportRetentionDays = req.ReportRetentionDays
		global.AppConfig.DSCPullServer.ReportCleanupIntervalMins = req.ReportCleanupIntervalMins
		global.AppConfig.DSCPullServer.EnableReleaseCheck = req.EnableReleaseCheck
		global.AppConfig.DSCPullServer.ReleaseCheckIntervalMins = req.ReleaseCheckIntervalMins

		w.WriteHeader(http.StatusNoContent)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func SchedulerRunCleanupAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	database, err := db.OpenDB(&global.AppConfig.Database)
	if err != nil {
		http.Error(w, "DB open error", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	retentionDays := global.AppConfig.DSCPullServer.ReportRetentionDays
	if retentionDays <= 0 {
		http.Error(w, "Invalid report retention configuration", http.StatusBadRequest)
		return
	}

	deleted, err := db.CleanupOldReports(database, global.AppConfig.Database.Driver, retentionDays)
	if err != nil {
		http.Error(w, "Cleanup failed", http.StatusInternalServerError)
		return
	}

	logs.WriteLogFile("INFO [REPORT CLEANUP] Manual cleanup triggered from admin")
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]int64{"deleted": deleted})
}

func SchedulerConfigFromApp(cfg *schema.AppConfig) SchedulerConfigDTO {
	return SchedulerConfigDTO{
		EnableReportAutoCleanup:  cfg.DSCPullServer.EnableReportAutoCleanup,
		ReportRetentionDays:      cfg.DSCPullServer.ReportRetentionDays,
		ReportCleanupIntervalMins: cfg.DSCPullServer.ReportCleanupIntervalMins,
		EnableReleaseCheck: cfg.DSCPullServer.EnableReleaseCheck,
		ReleaseCheckIntervalMins: cfg.DSCPullServer.ReleaseCheckIntervalMins,
	}
}

func SchedulerRunReleaseCheckAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	result, err := service.CheckLatestRelease(buildinfo.Version)
	driverName := "sqlite"
	if global.AppConfig != nil && global.AppConfig.Database.Driver != "" {
		driverName = global.AppConfig.Database.Driver
	}
	if err != nil {
		http.Error(w, "Release check failed", http.StatusInternalServerError)
		logs.WriteLogFile("WARN [RELEASE CHECK] Manual release check failed: " + err.Error())
		database, dbErr := db.OpenDB(&global.AppConfig.Database)
		if dbErr == nil {
			_ = db.PersistReleaseCheckFailure(database, driverName)
			database.Close()
		}
		return
	}
	database, dbErr := db.OpenDB(&global.AppConfig.Database)
	if dbErr == nil {
		_ = db.PersistReleaseCheckSuccess(database, driverName, result.LatestRelease, result.LatestReleaseURL, result.UpdateAvailable)
		database.Close()
	}

	if result.UpdateAvailable {
		logs.WriteLogFile("INFO [RELEASE CHECK] Manual check: update available latest=" + result.LatestRelease + ", current=" + result.CurrentVersion)
	} else {
		logs.WriteLogFile("INFO [RELEASE CHECK] Manual check: already up to date current=" + result.CurrentVersion)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"current_version": result.CurrentVersion,
		"latest_release": result.LatestRelease,
		"latest_release_url": result.LatestReleaseURL,
		"update_available": result.UpdateAvailable,
	})
}
