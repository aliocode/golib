# Common Golang Libraries

Common libraries for my personal golang projects.

## Packages

### libcid

A package for managing context identifiers (CIDs). This helps with request tracking and logging by providing unique identifiers that can be passed through different layers of an application.

### libcloser

Utilities for safely managing resource closures. This package helps prevent resource leaks by providing helpers to ensure that resources are properly closed, even in error scenarios.

### libconfig

Configuration management utilities. This package provides a simple and consistent way to load, validate, and access application configuration from various sources like environment variables, configuration files, and command-line flags.

### liblog

A structured logging package that integrates with the CID system for consistent request tracking. Provides different log levels and formatting options suitable for both development and production environments.

### libmode

Utilities for determining and managing application runtime modes (development, testing, production). This helps with environment-specific behaviors and configurations.

### libpgx

PostgreSQL database utilities built on top of the pgx driver. Provides connection management, transaction helpers, and common query patterns.

## Usage

To use these libraries in your Go project, add the repository to your go.mod file:

```go
require github.com/aliocode/golib v0.1.0
```

Then import the specific packages you need:

```go
import (
    "github.com/aliocode/golib/libconfig"
    "github.com/aliocode/golib/liblog"
)
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
