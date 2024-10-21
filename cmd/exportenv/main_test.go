package main

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
)

func TestCleanValue(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"\"hello\"", "hello"},
		{"'world'", "world"},
		{"`test`", "test"},
		{"no_quotes", "no_quotes"},
		{"\"escaped\\\"quote\"", "escaped\"quote"},
		{"'escaped\\'quote'", "escaped'quote"},
		{"`escaped\\`tick`", "escaped`tick"},
		{" ", ""},                               // Expecting empty string for input with only spaces
		{"\"\"", ""},                            // Empty quoted string should result in an empty string
		{"\"leading space \"", "leading space"}, // Leading and trailing spaces within quotes should be trimmed
	}

	for _, tt := range tests {
		result := cleanValue(tt.input)
		if result != tt.expected {
			t.Errorf("cleanValue(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestParseEnvFile(t *testing.T) {
	tests := []struct {
		input    string
		expected map[string]string
	}{
		{
			"KEY=value\n# Comment\n\n",
			map[string]string{"KEY": "value"},
		},
		{
			"KEY1=value1\nKEY2=value2\n",
			map[string]string{"KEY1": "value1", "KEY2": "value2"},
		},
		{
			"KEY=value # inline comment\n",
			map[string]string{"KEY": "value"},
		},
		{
			"KEY='value with spaces'\n",
			map[string]string{"KEY": "value with spaces"},
		},
		{
			"KEY=`value with backticks`\n",
			map[string]string{"KEY": "value with backticks"},
		},
		{
			"KEY=value\\ with\\ backslashes\n",
			map[string]string{"KEY": "value\\ with\\ backslashes"},
		},
		{
			"KEY=\n",
			map[string]string{"KEY": ""},
		},
	}

	for _, tt := range tests {
		envVars, err := parseEnvFileFromString(tt.input)
		if err != nil {
			t.Fatalf("parseEnvFileFromString(%q) failed with error: %v", tt.input, err)
		}
		if !reflect.DeepEqual(envVars, tt.expected) {
			t.Errorf("parseEnvFileFromString(%q) = %v, expected %v", tt.input, envVars, tt.expected)
		}
	}
}

// Helper function to simulate reading from a file
func parseEnvFileFromString(input string) (map[string]string, error) {
	scanner := bufio.NewScanner(strings.NewReader(input))
	envVars := make(map[string]string)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Match valid env variable patterns
		if matches := envLinePattern.FindStringSubmatch(line); matches != nil {
			key := strings.TrimSpace(matches[1])
			value := cleanValue(strings.TrimSpace(matches[2]))
			envVars[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return envVars, nil
}

func TestEnvLinePattern(t *testing.T) {
	tests := []struct {
		line     string
		expected bool
	}{
		{"KEY=value", true},
		{"KEY='value with spaces'", true},
		{"KEY=`value with backticks`", true},
		{"# This is a comment", false},
		{" ", false},
		{"invalid line", false},
	}

	for _, tt := range tests {
		matches := envLinePattern.MatchString(tt.line)
		if matches != tt.expected {
			t.Errorf("envLinePattern.MatchString(%q) = %v, expected %v", tt.line, matches, tt.expected)
		}
	}
}
