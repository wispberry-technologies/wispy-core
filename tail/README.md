# Tailwind CSS v4 Compiler for Go

This package provides a Go implementation of a Tailwind CSS v4 compatible compiler. It takes HTML content with Tailwind classes and outputs optimized CSS.

## Features

- Extract Tailwind classes from HTML content
- Sort classes according to Tailwind's ordering rules
- Generate optimized CSS output
- Command-line interface for processing files

## Usage

### As a Go Package

```go
package main

import (
    "fmt"
    "log"

    "github.com/theo/wispy-core/tail"
)

func main() {
    html := `<div class="flex p-4 bg-white">
        <h1 class="text-xl font-bold">Hello, world!</h1>
    </div>`

    css, err := tail.CompileHTML(html)
    if err != nil {
        log.Fatalf("Error compiling HTML: %v", err)
    }

    fmt.Println(css)
}
```

### Processing Files

```go
err := tail.CompileAndWriteFile("input.html", "output.css")
if err != nil {
    log.Fatalf("Error: %v", err)
}
```

### Command-Line Interface

```bash
# Process an HTML file
go run cmd/tailwindcli/main.go -input input.html -output output.css

# Process HTML from stdin
cat input.html | go run cmd/tailwindcli/main.go > output.css

# Process inline HTML
go run cmd/tailwindcli/main.go -html "<div class='flex p-4'>Hello</div>" > output.css
```

## Implementation Details

The compiler follows these steps:
1. Parse HTML and extract all class attributes
2. Deduplicate and normalize class names
3. Sort classes according to Tailwind's ordering system
4. Generate optimized CSS output

## Limitations

This is a simplified implementation and has the following limitations:
- Does not support custom Tailwind configurations
- Limited set of utility classes implemented
- No support for Just-In-Time mode
- Does not handle complex selectors or arbitrary values

## Testing

Run the tests with:

```bash
go test ./...
```
