# OAPI Builder

A library for generating OpenAPI specifications and validating request and response in Go.

## Features

- Generating OpenAPI 3.0 specifications
- Schema validation via libopenapi
- Flexible schema construction via builder pattern
- Support for core OpenAPI components:
    - Info
    - Paths
    - Components/Schemas
- Automatic YAML formatting with proper indentation

## Install

```bash
go get github.com/IlKhr/docovalidate
```

## How to use

```go
package main

import (
    "log/slog"
    "github.com/IlKhr/docovalidate/pkg/oapi_builder"
)

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
    builder := oapi_builder.New(logger)

    // Create schemas
    mainInfo := &MainInfoSchema{...}
    paths := []oapi_builder.Schemer{...}
    components := &ComponentsSchema{...}

    // Generate specification
    builder.MustGenerateSchemas(
        oapi_builder.HandlersWithSchemas{
            MainInfoSchemas:   mainInfo,
            PathSchemas:       paths,
            ComponentsSchemas: components,
        },
        "openapi.yaml",
    )
}
```

## Licence

MIT
