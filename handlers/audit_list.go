package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal/global"
	"go-dsc-pull/internal/schema"
)

func auditSearchClause(driverName string) string {
	if driverName == "mssql" || driverName == "sqlserver" {
		return ` WHERE
			LOWER(ISNULL(user_email, '')) LIKE ? OR
			LOWER(ISNULL(action, '')) LIKE ? OR
			LOWER(ISNULL(target, '')) LIKE ? OR
			LOWER(ISNULL(details, '')) LIKE ? OR
			LOWER(CONVERT(NVARCHAR(30), created_at, 120)) LIKE ?`
	}
	return ` WHERE
		LOWER(COALESCE(user_email, '')) LIKE ? OR
		LOWER(COALESCE(action, '')) LIKE ? OR
		LOWER(COALESCE(target, '')) LIKE ? OR
		LOWER(COALESCE(details, '')) LIKE ? OR
		LOWER(COALESCE(CAST(created_at AS TEXT), '')) LIKE ?`
}

// AuditListHandler returns paginated audit entries (GET /api/v1/audit).
func AuditListHandler(w http.ResponseWriter, r *http.Request) {
	draw := 0
	if rawDraw := r.URL.Query().Get("draw"); rawDraw != "" {
		if parsed, err := strconv.Atoi(rawDraw); err == nil && parsed >= 0 {
			draw = parsed
		}
	}

	limit := 20
	offset := 0
	if rawLength := r.URL.Query().Get("length"); rawLength != "" {
		parsed, err := strconv.Atoi(rawLength)
		if err != nil || parsed <= 0 {
			http.Error(w, "Invalid length", http.StatusBadRequest)
			return
		}
		if parsed > 200 {
			parsed = 200
		}
		limit = parsed
	} else if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed <= 0 {
			http.Error(w, "Invalid limit", http.StatusBadRequest)
			return
		}
		if parsed > 200 {
			parsed = 200
		}
		limit = parsed
	}

	if rawStart := r.URL.Query().Get("start"); rawStart != "" {
		parsed, err := strconv.Atoi(rawStart)
		if err != nil || parsed < 0 {
			http.Error(w, "Invalid start", http.StatusBadRequest)
			return
		}
		offset = parsed
	} else if rawOffset := r.URL.Query().Get("offset"); rawOffset != "" {
		parsed, err := strconv.Atoi(rawOffset)
		if err != nil || parsed < 0 {
			http.Error(w, "Invalid offset", http.StatusBadRequest)
			return
		}
		offset = parsed
	}

	search := strings.TrimSpace(r.URL.Query().Get("search[value]"))
	if search == "" {
		search = strings.TrimSpace(r.URL.Query().Get("search"))
	}

	database, err := db.OpenDB(&global.AppConfig.Database)
	if err != nil {
		http.Error(w, "DB open error", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	var total int64
	if err := database.QueryRow("SELECT COUNT(*) FROM audit").Scan(&total); err != nil {
		http.Error(w, "DB count error", http.StatusInternalServerError)
		return
	}

	filteredTotal := total
	whereClause := ""
	filterArgs := make([]interface{}, 0)
	if search != "" {
		like := "%" + strings.ToLower(search) + "%"
		whereClause = auditSearchClause(global.AppConfig.Database.Driver)
		filterArgs = []interface{}{like, like, like, like, like}

		countQuery := "SELECT COUNT(*) FROM audit" + whereClause
		if err := database.QueryRow(countQuery, filterArgs...).Scan(&filteredTotal); err != nil {
			http.Error(w, "DB filtered count error", http.StatusInternalServerError)
			return
		}
	}

	query := "SELECT id, user_email, action, target, details, ip_address, created_at FROM audit" + whereClause + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args := append(filterArgs, limit+1, offset)
	if global.AppConfig.Database.Driver == "mssql" || global.AppConfig.Database.Driver == "sqlserver" {
		query = fmt.Sprintf("SELECT id, user_email, action, target, details, ip_address, created_at FROM audit%s ORDER BY created_at DESC OFFSET ? ROWS FETCH NEXT ? ROWS ONLY", whereClause)
		args = append(filterArgs, offset, limit+1)
	}

	rows, err := database.Query(query, args...)
	if err != nil {
		http.Error(w, "DB query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	audits := make([]schema.Audit, 0)
	for rows.Next() {
		var a schema.Audit
		if err := rows.Scan(&a.ID, &a.UserEmail, &a.Action, &a.Target, &a.Details, &a.IPAddress, &a.CreatedAt); err == nil {
			audits = append(audits, a)
		}
	}

	hasMore := len(audits) > limit
	if hasMore {
		audits = audits[:limit]
	}

	w.Header().Set("Content-Type", "application/json")
	payload := map[string]interface{}{
		"items":    audits,
		"limit":    limit,
		"offset":   offset,
		"has_more": hasMore,
		"total":    total,
		"filtered": filteredTotal,
	}
	if draw > 0 {
		payload["draw"] = draw
		payload["recordsTotal"] = total
		payload["recordsFiltered"] = filteredTotal
		payload["data"] = audits
	}
	_ = json.NewEncoder(w).Encode(payload)
}

// AuditExportCSVHandler exports audit rows as CSV and applies optional search filter.
func AuditExportCSVHandler(w http.ResponseWriter, r *http.Request) {
	database, err := db.OpenDB(&global.AppConfig.Database)
	if err != nil {
		http.Error(w, "DB open error", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	search := strings.TrimSpace(r.URL.Query().Get("search"))
	whereClause := ""
	args := make([]interface{}, 0)
	if search != "" {
		like := "%" + strings.ToLower(search) + "%"
		whereClause = auditSearchClause(global.AppConfig.Database.Driver)
		args = []interface{}{like, like, like, like, like}
	}

	query := "SELECT user_email, action, target, details, created_at FROM audit" + whereClause + " ORDER BY created_at DESC"
	rows, err := database.Query(query, args...)
	if err != nil {
		http.Error(w, "DB query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"audit_export.csv\"")

	cw := csv.NewWriter(w)
	defer cw.Flush()
	_ = cw.Write([]string{"Email", "Action", "Target", "Details", "Date"})

	for rows.Next() {
		var email, action, target, details, createdAt string
		if err := rows.Scan(&email, &action, &target, &details, &createdAt); err != nil {
			continue
		}
		_ = cw.Write([]string{email, action, target, details, createdAt})
	}
}
