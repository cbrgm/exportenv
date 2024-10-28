# exportenv

`exportenv` is a simple, easy-to-use utility for reading environment variables from `.env` files and exporting them to the current shell session. It supports multiple `.env` files, variable expansion, overriding variables, and evaluating environment variables without exporting them.

## Usage

Run the tool to read environment variables from `.env` files and export them into your session. By default, `exportenv` uses `.env` in the current directory if no file is specified.

### Flags

- `--env-file, -f <path>`: Specify one or more paths to `.env` files, processed in order. If no files are provided, `exportenv` defaults to using `.env` in the current directory.
- `--override, -o`: Allow variables in succeeding `.env` files to overwrite variables from earlier ones.
- `--no-expand`: Disable variable expansion for `${VAR}` syntax in `.env` values.
- `-v <KEY=VALUE>`: Set variables directly from the command line, which take precedence over `.env` files.
- `--`: Use `--` before a command to execute it with the loaded environment variables.

### Examples

#### Basic Usage

Load variables from `.env` in the current directory and export them:
```
eval $(./exportenv)
```

#### Using Custom `.env` Files

Load variables from specified `.env` files:
```
eval $(./exportenv --env-file /path/to/.env1 --env-file /path/to/.env2)
```

#### Override Variables

Use multiple `.env` files where later files override variables from earlier ones:
```
eval $(./exportenv --env-file .env --env-file .env.local --override)
```

#### Set Command-Line Variables

Define variables directly via `-v`, which take precedence over `.env` files:
```
eval $(./exportenv -v NAME=JohnDoe -v ENV=production)
```

#### Running Commands with Environment Variables

Run a command with the variables loaded from `.env`. Use `--` before the command to pass it to `exportenv`:
```
./exportenv --env-file .env -- my_command --option=value
```

### .env File Format

A valid `.env` file should follow these guidelines:

- Each line should be in the `KEY=VALUE` format.
- Comments start with `#` and are ignored.
- Empty lines are skipped.
- Values can be:
  - **Unquoted**: `FOO=bar baz`
  - **Double-quoted**: `FOO="bar baz"`, supporting escape sequences like `\n`, `\t`, and `\"`.
  - **Single-quoted**: `FOO='bar baz'`, which takes the value literally, including special characters.

Example `.env` file:

```
# Application settings
APP_NAME="My Application"
APP_ENV=production

# Database settings
DB_HOST=localhost
DB_PORT=5432
DB_USER="my_user"
DB_PASS='my$ecret'

# Multiline value for a private key
PRIVATE_KEY="```--BEGIN PRIVATE KEY```--\nMIIEvQIBADANB ... \n```--END PRIVATE KEY```--"
```

### Example of Running Commands with Environment Variables

1. **Run a command** with variables from multiple `.env` files (e.g., `.env`, `.env.local`) and command-line overrides:
```
./exportenv --env-file .env --env-file .env.local -v MODE=debug -- my_command --option=value
```

2. **Load and export** variables from `.env` files with overrides and run a server:
```
eval $(./exportenv --env-file .env --env-file .env.production --override) && npm start
```

### Notes

- `.env` files are processed in the order they’re specified, unless `--override` is set.
- Variable expansion (`${VAR}` syntax) is enabled by default but can be disabled with `--no-expand`.
- Always use `eval` when loading variables to ensure they are exported into the current session.

