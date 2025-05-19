# AgentRuntime Development Guidelines

Note: For comprehensive contribution guidelines, please also see [CONTRIBUTING.md](./CONTRIBUTING.md).

## Build Commands
- Build: `make build`
- Run tests: `make test`
- Run single test: `go test -v ./path/to/package -run TestName`
- Lint: `make lint`
- Clean: `make clean`

## Code Style
- Imports: Standard Go import grouping (stdlib, 3rd party, internal)
- Error handling: Return errors with package `err` variables when possible
- Naming: Use camelCase for variables, PascalCase for exported identifiers
- DI: Use dependency injection via internal/di package
- Testing: Use testify/suite for test organization
- Comments: Document public interfaces and complex logic
- Errors: Wrap errors with context when propagating up the call stack
- Struct tags: Use consistent field tags for GORM and JSON serialization
- Context: Use context.Context for cancellation and timeouts