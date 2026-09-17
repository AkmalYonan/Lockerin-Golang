package docs

import (
	_ "embed"
)

// OpenAPISpec contains the embedded OpenAPI 3.0 specification for Lockerin API.
//
//go:embed openapi.yaml
var OpenAPISpec []byte
