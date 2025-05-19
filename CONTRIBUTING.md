# Contributing to AgentRuntime

Thank you for considering contributing to AgentRuntime! This document outlines the process for contributing to the project and provides guidelines to follow.

## Code of Conduct

By participating in this project, you are expected to uphold our Code of Conduct: be respectful, considerate, and constructive in all interactions.

## Getting Started

1. Fork the repository
2. Clone your fork locally
   ```bash
   git clone https://github.com/yourusername/agentruntime.git
   cd agentruntime
   ```
3. Add the upstream repository as a remote
   ```bash
   git remote add upstream https://github.com/habiliai/agentruntime.git
   ```
4. Create a new branch for your work
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Development Environment

### Prerequisites

- Go 1.24 or higher
- Make
- Git
- Docker-compose (for running tests)

### Setting Up the Environment
1. Install Go and set up your Go workspace
2. Install dependencies
   ```bash
   go mod tidy
   ``` 
3. Set up your local environment variables. You can copy the `.env.example` file to `.env` and modify it as needed.
   ```bash
    cp .env.example .env
    ```
4. Before you test, you should run the following command to set up the database:
   ```bash
   docker compose up postgres # or docker-compose up postgres
   ```
5. Run all tests
   ```bash
    make test
    ```

### Building and Testing

We use Make to automate common development tasks:

- Build: `make build`
- Run tests: `make test`
- Run single test: `go test -v ./path/to/package -run TestName`
- Lint: `make lint`
- Clean: `make clean`

## Code Style Guidelines

Please follow these style guidelines when writing code:

- **Imports**: Follow standard Go import grouping:
  1. Standard library imports
  2. Third-party imports
  3. Internal/project imports

- **Error Handling**: Return errors using package `err` variables when possible

- **Naming Conventions**:
  - Use camelCase for variables
  - Use PascalCase for exported identifiers

- **Dependency Injection**: Use dependency injection via internal/di package

- **Testing**: Use testify/suite for test organization

- **Comments**: Document public interfaces and complex logic

- **Error Wrapping**: Wrap errors with context when propagating up the call stack

- **Struct Tags**: Use consistent field tags for GORM and JSON serialization

## Pull Request Process

1. Ensure your code follows our style guidelines and passes all tests
2. Update the documentation if necessary
3. Make sure your commits are clean, logical, and have clear messages
4. Open a pull request against the `main` branch
5. Fill out the pull request template with details about your changes
6. Request review from a maintainer

### Commit Message Guidelines

Use a clear and descriptive title for your commits with the following prefixes:

- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation changes
- `style:` for code style changes (formatting, etc)
- `refactor:` for code refactoring
- `test:` for adding tests
- `chore:` for routine tasks, maintenance, etc

Example: `feat: add support for X`

## Feature Requests and Bug Reports

Please use GitHub Issues to submit feature requests and bug reports. For bug reports, please include:

1. A clear, descriptive title
2. Steps to reproduce the issue
3. Expected behavior
4. Actual behavior
5. Environment details (OS, Go version, etc)

For feature requests, describe the feature you'd like to see and the problem it solves.

## License

By contributing to AgentRuntime, you agree that your contributions will be licensed under the project's [MIT License](LICENSE).

## Questions?

If you have questions or need help, feel free to:

- Open an issue with your question
- Reach out to the maintainers

Thank you for contributing to AgentRuntime!