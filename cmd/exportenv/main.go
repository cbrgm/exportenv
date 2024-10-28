package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/alexflint/go-arg"
)

type Args struct {
	EnvFiles []string `arg:"--env-file,separate" help:"Paths to the .env files, processed in the order given"`
	NoExpand bool     `arg:"--no-expand" help:"Disable variable expansion"`
	Override bool     `arg:"-o,--override" help:"Override variables from previous files if they already exist"`
	Vars     []string `arg:"-v,--var,separate" help:"Set variables from command line in the form KEY=VALUE"`
	Cmd      []string `arg:"positional" help:"Command to execute with the environment variables"`
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	slog.SetDefault(logger)

	var args Args
	arg.MustParse(&args)

	// Load env files with the specified override behavior
	envVars, err := loadEnvFiles(args.EnvFiles, args.Override)
	if err != nil {
		slog.Error("Error loading env files", slog.Any("error", err))
		return
	}

	cmdVars := parseCommandLineVars(args.Vars)
	mergeEnvVars(envVars, cmdVars)

	if !args.NoExpand {
		expandEnvVars(envVars)
	}

	sortedEnvVars := sortEnvVars(envVars)

	if len(args.Cmd) == 0 {
		printExportableEnvVars(sortedEnvVars)
		return
	}

	handleExecution(args.Cmd, sortedEnvVars)
}

// loadEnvFiles loads variables from multiple env files in order, using .env as a default if no files are provided.
// If override is true, succeeding files will overwrite variables from previous files.
func loadEnvFiles(files []string, override bool) (map[string]string, error) {
	// Use .env as default if no files are specified
	if len(files) == 0 {
		files = []string{".env"}
	}

	envVars := make(map[string]string)
	for _, file := range files {
		fileVars, err := parseEnvFile(file)
		if err != nil {
			return nil, err
		}
		for k, v := range fileVars {
			// Set variable only if it doesn't exist or override is true
			if override || !existsInMap(envVars, k) {
				envVars[k] = v
			}
		}
	}
	return envVars, nil
}

// existsInMap checks if a key exists in the map.
func existsInMap(m map[string]string, key string) bool {
	_, exists := m[key]
	return exists
}

// parseCommandLineVars parses command-line variables from -v flags.
func parseCommandLineVars(vars []string) map[string]string {
	cmdVars := make(map[string]string)
	for _, v := range vars {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) == 2 {
			cmdVars[parts[0]] = parts[1]
		} else if len(parts) == 1 {
			// If no value is specified, treat it as an empty value
			cmdVars[parts[0]] = ""
		}
	}
	return cmdVars
}

// mergeEnvVars merges .env variables with command-line variables (command-line takes precedence).
func mergeEnvVars(envVars, cmdVars map[string]string) {
	for k, v := range cmdVars {
		envVars[k] = v
	}
}

// expandEnvVars performs variable expansion (e.g., ${VAR} syntax) in .env values.
func expandEnvVars(envVars map[string]string) {
	for key, value := range envVars {
		envVars[key] = os.Expand(value, func(varName string) string {
			if val, ok := envVars[varName]; ok {
				return val
			}
			return ""
		})
	}
}

// printExportableEnvVars prints environment variables in an exportable format.
func printExportableEnvVars(sortedEnvVars []string) {
	for _, v := range sortedEnvVars {
		parts := strings.SplitN(v, "=", 2)
		key := parts[0]
		value := ""
		if len(parts) > 1 {
			value = parts[1]
		}

		// Always enclose the value in double quotes to ensure compatibility with spaces and special characters.
		// If the value is empty, it will be output as export key="".
		quotedValue := `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`

		// Print the export statement
		fmt.Printf("export %s=%s\n", key, quotedValue)
	}
}

// handleExecution executes the given command within the modified environment.
func handleExecution(cmdArgs, envVars []string) {
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Env = append(os.Environ(), envVars...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		slog.Error("Error executing command", slog.Any("error", err))
	}
}

// sortEnvVars sorts environment variables by key.
func sortEnvVars(envVars map[string]string) []string {
	keys := make([]string, 0, len(envVars))
	for k := range envVars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sortedEnv := make([]string, len(envVars))
	for i, k := range keys {
		sortedEnv[i] = k + "=" + envVars[k]
	}
	return sortedEnv
}

// parseEnvFile reads an env file into a map with support for comments, multiline values, and interpolation.
func parseEnvFile(filePath string) (map[string]string, error) {
	envVars := make(map[string]string)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	// nolint: errcheck
	defer file.Close()

	var (
		key       string
		value     string
		multiline bool
		quoteType rune
	)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Ignore comment or empty lines
		if isCommentOrEmpty(line) {
			continue
		}

		// Handle multiline values continuation
		if multiline {
			// Check if the multiline value ends on this line
			if strings.HasSuffix(line, string(quoteType)) {
				// Remove trailing quote and add the line to the multiline value
				value += "\n" + strings.TrimSuffix(line, string(quoteType))
				// Remove any inline comment after the closing quote
				value = removeInlineComment(value)
				envVars[key] = value
				multiline = false
			} else {
				// Continue adding to the multiline value
				value += "\n" + line
			}
			continue
		}

		// Parse line to get key, value, and multiline start
		var val string
		key, val, multiline, quoteType = parseLine(line)
		if multiline {
			value = val
			continue
		}

		// Expand variables for double-quoted values
		if quoteType == '"' {
			val = expandVariables(val, envVars)
			val = strings.ReplaceAll(val, `\n`, "\n") // Handle \n as newlines
		}

		// Store key-value pair
		envVars[key] = val
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return envVars, nil
}

// parseLine parses a line and returns the key, value, and whether it is a multiline start.
func parseLine(line string) (string, string, bool, rune) {
	keyValueLine := regexp.MustCompile(`^\s*([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)$`)
	matches := keyValueLine.FindStringSubmatch(line)
	if matches == nil {
		return "", "", false, 0
	}

	key, val := matches[1], matches[2]

	// Remove inline comments if outside quotes
	val = removeInlineComment(val)

	// Check for quoted values (single or double)
	if strings.HasPrefix(val, "\"") || strings.HasPrefix(val, "'") {
		quoteType := rune(val[0])
		val = strings.TrimPrefix(val, string(quoteType))

		// Check if it's a single-line quoted value by verifying it ends with the same quote
		if strings.HasSuffix(val, string(quoteType)) {
			val = strings.TrimSuffix(val, string(quoteType))
			return key, val, false, 0 // Single-line quoted value
		}

		// Start of a multiline quoted value
		return key, val, true, quoteType
	}

	// Unquoted single-line value
	return key, val, false, 0
}

// isCommentOrEmpty checks if a line is a comment or empty.
func isCommentOrEmpty(line string) bool {
	return line == "" || strings.HasPrefix(line, "#")
}

// removeInlineComment removes inline comments if not inside quotes.
func removeInlineComment(val string) string {
	var result strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, char := range val {
		if (char == '"' || char == '\'') && !inQuote {
			// Starting a quoted section
			inQuote = true
			quoteChar = char
		} else if char == quoteChar && inQuote {
			// Ending a quoted section
			inQuote = false
		} else if char == '#' && !inQuote {
			// Found a comment outside quotes; ignore the rest of the line
			break
		}
		result.WriteRune(char)
	}

	return strings.TrimSpace(result.String())
}

// expandVariables expands ${VAR} syntax for double-quoted values.
func expandVariables(val string, envVars map[string]string) string {
	return os.Expand(val, func(varName string) string {
		if v, exists := envVars[varName]; exists {
			return v
		}
		return ""
	})
}
