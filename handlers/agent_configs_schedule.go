package handlers

import (
	"go-dsc-pull/internal/schema"
	"go-dsc-pull/internal/scheduling"
	"time"
)

func decorateBindingsWithNextExecution(bindings []schema.AgentConfigurationBinding, now time.Time) {
	for i := range bindings {
		b := &bindings[i]
		b.NextExecutionStartAt = nil
		b.NextExecutionEndAt = nil
		b.NextExecutionCompleted = false

		start, end, completed, ok := scheduling.ComputeNextExecutionWindow(*b, now)
		if !ok {
			continue
		}
		if completed {
			b.NextExecutionCompleted = true
			continue
		}

		startStr := start.UTC().Format(time.RFC3339)
		endStr := end.UTC().Format(time.RFC3339)
		b.NextExecutionStartAt = &startStr
		b.NextExecutionEndAt = &endStr
	}
}
