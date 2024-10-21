# exportenv

`exportenv` is a dead simple, easy-to-use, and tiny utility for reading environment variables from a `.env` file and exporting them to the current shell session. It supports evaluating the environment variables, pointing to custom `.env` files, and unsetting variables.

## Usage

The basic usage involves running the tool, which reads the `.env` file in the current directory (or a specified file) and exports the environment variables into your session.

### Flags

- `-f, --file`: Specify the path to the `.env` file. If not provided, the tool defaults to a `.env` file in the current directory.
- `-e, --eval`: Evaluate and display the environment variables without exporting them.
- `--unset`: Unset the environment variables defined in the `.env` file, removing them from the current session.

### Examples

#### Basic Usage

```sh
eval $(./exportenv)
```

This command exports the environment variables from the `.env` file in the current directory.

#### Usage with Custom File

```sh
eval $(./exportenv --file /path/to/your/.env_file)
```

This command exports the environment variables from the specified `.env_file`.

#### Preview (Evaluate) Environment Variables

```sh
./exportenv --eval --file /path/to/your/.env_file
```

This command displays the environment variables from the specified file without exporting them.

#### Unsetting Environment Variables

```sh
eval $(./exportenv --unset --file /path/to/your/.env_file)
```

This command unsets the environment variables defined in the specified `.env_file` from the current session.

## Complex Usage Example

If you want to preview the variables from a custom `.env` file before exporting them and then proceed to export them, you can follow this two-step process:

1. **Preview (Evaluate) the Variables:**
```sh
./exportenv --eval --file ~/projects/config/.env
```

2. **Export the Variables:**
```sh
eval $(./exportenv --file ~/projects/config/.env)
```

This way, you ensure that the environment variables look correct before you actually export them.

## Notes

- The `.env` file should follow standard formatting with `KEY=VALUE` pairs. Any comments (lines starting with `#`) or malformed lines will be ignored.
- Values can be enclosed in quotes (`'`, `"`, or ``` ` ```), and the tool will handle removing these during export.

