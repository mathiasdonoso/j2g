# j2g
Generate Go structs from JSON.

## Warning :warning:
This project is still in WIP. :construction:
Expect rough edges and occasional bugs.

## The problem
When working with APIs or JSON payloads, you often need to manually write structs that match JSON responses.
This is time-consuming and error-prone, especially when dealing with large or deeply nested objects.

## The solution
**j2g** automates that process.
It takes JSON input and generates properly formatted Go structs.

## Limitations :warning:
Currently, **j2g** only supports JSON objects ({ ... }) as input.
JSON documents that start with an array ([ ... ]) **are not yet supported**, but support for this format is planned for a future release.

## Installation
```bash
go install github.com/mathiasdonoso/j2g@latest
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

## Contributing
Feel free to open issues or submit pull requests to improve the tool. Contributions are welcome.
