package middleware

import (
    "net/http"
    "database/sql"
    "go-dsc-pull/internal/auth"
)

// WebAdminOnly returns a middleware that restricts access to admins (calls renderDenied on failure)
func WebAdminOnly(dbConn *sql.DB, renderDenied func(http.ResponseWriter)) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !auth.IsAdmin(r, dbConn) {
                w.WriteHeader(http.StatusForbidden)
                renderDenied(w)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
