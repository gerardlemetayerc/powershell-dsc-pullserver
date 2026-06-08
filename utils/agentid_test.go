package utils

import "testing"

func TestExtractAgentId(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "classic DSC node segment",
			raw:  "Nodes(AgentId='E7BCBD5C-368C-11EE-9B28-005056B8B9CC')",
			want: "E7BCBD5C-368C-11EE-9B28-005056B8B9CC",
		},
		{
			name: "agent id segment only",
			raw:  "(AgentId='E7BCBD5C-368C-11EE-9B28-005056B8B9CC')",
			want: "E7BCBD5C-368C-11EE-9B28-005056B8B9CC",
		},
		{
			name: "raw guid fallback",
			raw:  "E7BCBD5C-368C-11EE-9B28-005056B8B9CC",
			want: "E7BCBD5C-368C-11EE-9B28-005056B8B9CC",
		},
		{
			name: "raw quoted guid fallback",
			raw:  "'E7BCBD5C-368C-11EE-9B28-005056B8B9CC'",
			want: "E7BCBD5C-368C-11EE-9B28-005056B8B9CC",
		},
		{
			name: "empty input",
			raw:  "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractAgentId(tt.raw)
			if got != tt.want {
				t.Fatalf("ExtractAgentId(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}
