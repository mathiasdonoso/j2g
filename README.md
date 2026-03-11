# j2g
Generate Go structs from JSON.

> **Work in progress.** Expect rough edges and occasional bugs.

## The problem
When working with APIs or JSON payloads, you often need to manually write structs that match JSON responses.
This is time-consuming and error-prone, especially when dealing with large or deeply nested objects.

## The solution
**j2g** automates that process.
It takes JSON input and generates properly formatted Go structs.

## Installation
```bash
go install github.com/mathiasdonoso/j2g/cmd/j2g@latest
```

## Usage
**j2g** works exclusively with standard input and output.
It reads JSON from stdin and prints the generated Go code to stdout.
This makes it easy to use with tools like curl, cat, or pipes in general.

```bash
# Create a struct from a JSON file and print the result to stdout.
cat request.json | j2g
```

```bash
# Create a struct from a JSON file and save the result to file.go.
cat request.json | j2g > file.go
```

```bash
# From a curl command
curl https://api.restful-api.dev/objects/1 | j2g > structs.go
```

```bash
# Override the root struct name with --name
echo '{"id": 1, "name": "Alice"}' | j2g --name Response
```

## Flags

| Flag | Description | Default |
|---|---|---|
| `-n`, `--name <StructName>` | Name of the root Go struct | `Result` |
| `-h`, `--help` | Show help message | — |

## Use with AI agents (experimental idea)

> **Note:** This is an untested hypothesis — I'm not sure if this is a legitimate use case in practice, but it seems worth trying.

AI coding agents (like Claude Code, Copilot, etc.) typically generate Go structs from JSON by reasoning about the payload inline — spending tokens on type inference, field name normalization, and nesting. Since `j2g` handles all of that deterministically, an agent with shell access could delegate the work instead:

```bash
# Agent fetches an API response and pipes it directly to j2g
curl https://api.example.com/data | j2g
```

The potential benefit: the agent skips the reasoning step and gets back a correct struct immediately, saving output tokens and avoiding hallucinated types or missed fields. This matters most for large, deeply nested payloads.

If you try this and it works well (or doesn't), feel free to open an issue — feedback on this use case is welcome.

## Contributing
Feel free to open issues or submit pull requests to improve the tool. Contributions are welcome.
