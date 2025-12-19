package utils

import "testing"

func TestExtractConfigName(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"Configurations(ConfigurationName='TestConfig')", "TestConfig"},
		{"Configurations(ConfigurationName='MyConfig')", "MyConfig"},
		{"Configurations(ConfigurationName='')", ""},
		{"Other(ConfigurationName='Nope')", "Other(ConfigurationName='Nope')"},
		{"", ""},
	}
	for _, c := range cases {
		got := ExtractConfigName(c.in)
		if got != c.out {
			t.Errorf("ExtractConfigName(%q) = %q, want %q", c.in, got, c.out)
		}
	}
}
