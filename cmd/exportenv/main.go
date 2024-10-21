package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"

	"github.com/alexflint/go-arg"
)

type Args struct {
	File  string `arg:"-f,--file" help:"Path to the .env file"`
	Eval  bool   `arg:"-e,--eval" help:"Evaluate the env file without exporting"`
	Unset bool   `arg:"--unset" help:"Unset the environment variables defined in the file"`
}

var envLinePattern = regexp.MustCompile(`^\s*([^#=]+)\s*=\s*(.*?)(\s*#.*)?$`)

func main() {
	// Configure slog with JSON handler
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	slog.SetDefault(logger)

	var args Args
	arg.MustParse(&args)

	filePath := ".env"
	if args.File != "" {
		filePath = args.File
	}

	envVars, err := parseEnvFile(filePath)
	if err != nil {
		slog.Error("Error parsing env file", slog.String("file", filePath), slog.Any("error", err))
		return
	}

	if args.Eval {
		displayEnvVars(envVars)
		return
	}

	if args.Unset {
		for key := range envVars {
			fmt.Printf("unset %s\n", key)
		}
		return
	}

	for key, value := range envVars {
		fmt.Printf("export %s=%s\n", key, value)
	}
}

// parseEnvFile reads and parses the env file, ignoring comments and empty lines.
func parseEnvFile(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			slog.Error("Error closing file", slog.String("file", filePath), slog.Any("error", err))
		}
	}()

	envVars := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Ignore comments and empty lines.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Match valid env variable patterns.
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

// cleanValue handles quotes and removes comments from value.
func cleanValue(value string) string {
	value = strings.TrimSpace(value) // Trim leading and trailing spaces.

	if len(value) == 0 {
		return value
	}

	// Remove quotes if value is wrapped in single, double, or backticks.
	if (value[0] == '"' && value[len(value)-1] == '"') ||
		(value[0] == '\'' && value[len(value)-1] == '\'') ||
		(value[0] == '`' && value[len(value)-1] == '`') {
		value = value[1 : len(value)-1]
	}

	value = strings.TrimSpace(value) // Trim spaces after removing quotes.

	// Replace escaped quotes and backticks.
	value = strings.ReplaceAll(value, `\"`, `"`)
	value = strings.ReplaceAll(value, `\'`, `'`)
	value = strings.ReplaceAll(value, "\\`", "`")

	return value
}

// displayEnvVars prints the env vars for evaluation purposes.
func displayEnvVars(envVars map[string]string) {
	fmt.Println("Evaluated environment variables:")
	for key, value := range envVars {
		fmt.Printf("%s=%s\n", key, value)
	}
}
