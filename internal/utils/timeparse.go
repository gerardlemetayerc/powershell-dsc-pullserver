package utils

import "time"

// ParseTimeFlexible gère plusieurs formats de date (SQLite, ISO8601, etc.)
func ParseTimeFlexible(s string) time.Time {
       if s == "" {
               return time.Time{}
       }
       // Essaye format classique
       if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
               return t
       }
       // Essaye format ISO8601 (avec T et Z)
       if t, err := time.Parse(time.RFC3339, s); err == nil {
               return t
       }
       // Essaye format sans Z
       if t, err := time.Parse("2006-01-02T15:04:05", s); err == nil {
               return t
       }
       return time.Time{}
}