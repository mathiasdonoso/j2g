# j2g
Generate Go structs from JSON — fast and easy.

## The problem
When working with APIs or JSON payloads, you often need to manually write structs that match JSON responses.
This is time-consuming and error-prone, especially when dealing with large or deeply nested objects.

## The solution
**j2g** automates that process.
It takes JSON input (from a file, stdin, or even a curl command) and generates properly formatted Go structs — ready to paste or save to a .go file.

## Installation
```bash
go install github.com/mathiasdonoso/j2g
```

## Usage
From a json file
```bash
# Create a struct from a JSON file and save it to file.go
j2g request.json -o file.go
```

Print to stdout (default)
```bash
j2g request.json
```

From a curl command
```bash
# You can also create a struct from a curl request
curl https://api.restful-api.dev/objects | j2g request.json
```

Customize the struct name
```bash
j2g request.json -name Response
```

## Contributing
Feel free to open issues or submit pull requests to improve the tool. Contributions are welcome!
