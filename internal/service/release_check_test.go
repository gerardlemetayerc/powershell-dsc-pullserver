package service

import "testing"

func TestVersionsEqualIgnoringVPrefix(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{name: "same without prefix", a: "1.1.3", b: "1.1.3", want: true},
		{name: "latest has v prefix", a: "v1.1.3", b: "1.1.3", want: true},
		{name: "current has v prefix", a: "1.1.3", b: "v1.1.3", want: true},
		{name: "both have v prefix", a: "v1.1.3", b: "v1.1.3", want: true},
		{name: "uppercase V prefix", a: "V1.1.3", b: "1.1.3", want: true},
		{name: "different versions", a: "v1.1.4", b: "1.1.3", want: false},
		{name: "empty latest", a: "", b: "1.1.3", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := versionsEqualIgnoringVPrefix(tt.a, tt.b)
			if got != tt.want {
				t.Fatalf("versionsEqualIgnoringVPrefix(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
