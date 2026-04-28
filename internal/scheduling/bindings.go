package scheduling

import (
	"go-dsc-pull/internal/schema"
	"strings"
	"time"
)

// ParseDBTime parses supported date formats from DB payloads.
func ParseDBTime(value string) (time.Time, bool) {
	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05.000",
		"2006-01-02 15:04:05.9999999",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05.000",
		"2006-01-02T15:04:05.9999999",
		"2006-01-02 15:04",
		"2006-01-02T15:04",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t.UTC(), true
		}
	}
	return time.Time{}, false
}

func normalizedWindowMinutes(binding schema.AgentConfigurationBinding) int {
	windowMinutes := binding.WindowMinutes
	if windowMinutes <= 0 {
		windowMinutes = 30
	}
	return windowMinutes
}

func normalizedScheduleType(binding schema.AgentConfigurationBinding) string {
	return strings.ToLower(strings.TrimSpace(binding.ScheduleType))
}

// IsScheduleDue returns whether a binding is due now and the matching occurrence start.
func IsScheduleDue(binding schema.AgentConfigurationBinding, now time.Time) (bool, time.Time) {
	scheduleType := normalizedScheduleType(binding)
	if scheduleType != "oneshot" && scheduleType != "recurring" {
		return false, time.Time{}
	}
	if !binding.Enabled || binding.ScheduledAt == nil || *binding.ScheduledAt == "" {
		return false, time.Time{}
	}

	startAt, ok := ParseDBTime(*binding.ScheduledAt)
	if !ok {
		return false, time.Time{}
	}

	window := time.Duration(normalizedWindowMinutes(binding)) * time.Minute
	if now.Before(startAt) {
		return false, time.Time{}
	}

	if scheduleType == "oneshot" {
		if binding.ScheduledLastAppliedAt != nil && *binding.ScheduledLastAppliedAt != "" {
			return false, time.Time{}
		}
		if now.After(startAt.Add(window)) {
			return false, time.Time{}
		}
		return true, startAt
	}

	if binding.RecurrenceMinutes == nil || *binding.RecurrenceMinutes <= 0 {
		return false, time.Time{}
	}

	interval := time.Duration(*binding.RecurrenceMinutes) * time.Minute
	elapsed := now.Sub(startAt)
	occurrenceCount := int64(elapsed / interval)
	occurrenceStart := startAt.Add(time.Duration(occurrenceCount) * interval)
	if now.After(occurrenceStart.Add(window)) {
		return false, time.Time{}
	}

	if binding.ScheduledLastAppliedAt != nil && *binding.ScheduledLastAppliedAt != "" {
		if lastApplied, ok := ParseDBTime(*binding.ScheduledLastAppliedAt); ok {
			if !lastApplied.Before(occurrenceStart) {
				return false, time.Time{}
			}
		}
	}

	return true, occurrenceStart
}

// ComputeNextExecutionWindow computes the next visible execution window for a binding.
// Returns startUTC, endUTC, completed, ok.
func ComputeNextExecutionWindow(binding schema.AgentConfigurationBinding, now time.Time) (time.Time, time.Time, bool, bool) {
	if !binding.Enabled {
		return time.Time{}, time.Time{}, false, false
	}

	scheduleType := normalizedScheduleType(binding)
	if scheduleType == "none" {
		return time.Time{}, time.Time{}, false, false
	}
	if binding.ScheduledAt == nil || *binding.ScheduledAt == "" {
		return time.Time{}, time.Time{}, false, false
	}

	startAt, ok := ParseDBTime(*binding.ScheduledAt)
	if !ok {
		return time.Time{}, time.Time{}, false, false
	}

	window := time.Duration(normalizedWindowMinutes(binding)) * time.Minute

	if scheduleType == "oneshot" {
		if binding.ScheduledLastAppliedAt != nil && *binding.ScheduledLastAppliedAt != "" {
			return time.Time{}, time.Time{}, true, true
		}
		return startAt, startAt.Add(window), false, true
	}

	if scheduleType != "recurring" || binding.RecurrenceMinutes == nil || *binding.RecurrenceMinutes <= 0 {
		return time.Time{}, time.Time{}, false, false
	}

	interval := time.Duration(*binding.RecurrenceMinutes) * time.Minute
	occurrenceStart := startAt
	if now.After(occurrenceStart) {
		elapsed := now.Sub(occurrenceStart)
		occurrenceCount := int64(elapsed / interval)
		occurrenceStart = occurrenceStart.Add(time.Duration(occurrenceCount) * interval)
	}
	occurrenceEnd := occurrenceStart.Add(window)

	currentWindowAlreadyApplied := false
	if binding.ScheduledLastAppliedAt != nil && *binding.ScheduledLastAppliedAt != "" {
		if lastApplied, ok := ParseDBTime(*binding.ScheduledLastAppliedAt); ok {
			currentWindowAlreadyApplied = !lastApplied.Before(occurrenceStart) && !lastApplied.After(occurrenceEnd)
		}
	}

	if now.After(occurrenceEnd) || currentWindowAlreadyApplied {
		occurrenceStart = occurrenceStart.Add(interval)
		occurrenceEnd = occurrenceStart.Add(window)
	}

	return occurrenceStart, occurrenceEnd, false, true
}
