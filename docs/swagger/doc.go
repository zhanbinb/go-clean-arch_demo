// Package swagger hosts the GENERATED OpenAPI documentation.
//
// Until `make swagger` runs, this file is a STUB. The stub exists so that
// `go build ./cmd/rest` succeeds without first running code generation.
//
// When `make swagger` is executed, this file is REPLACED with the real
// generated code that defines SwaggerInfo, registers it via swag.Register,
// and wires up the ginSwagger handler. After replacement the /swagger UI
// endpoint becomes functional.
//
// If you started the server and the /swagger/index.html page is empty or
// missing, you forgot to run \`make swagger\`.
package swagger

import "fmt"

func init() {
	const hint = "[swagger] stub package loaded — Swagger UI is DISABLED. " +
		"Run `make swagger` to generate real docs/swagger/*.go files."
	fmt.Println(hint)
}
