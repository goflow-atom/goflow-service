//go:build wireinject
// +build wireinject

// Package main provides the Wire injector for the GoFlow server.
//
// This file uses Wire's compile-time dependency injection to generate
// the initialization code. The actual implementation is generated in
// wire_gen.go by running the `wire` command.
//
// To regenerate the wire_gen.go file, run:
//
//	go run github.com/google/wire/cmd/wire
//
// Or use the Makefile:
//
//	make wire
package main

import (
	"github.com/goflow-atom/goflow-service/internal/app"
	"github.com/goflow-atom/goflow-service/internal/server"
	"github.com/google/wire"
)

// InitializeApplication initializes the entire application with all dependencies.
//
// This function is a Wire injector that uses the app.ProviderSet to create
// all required dependencies and return a fully configured server instance.
//
// Wire will generate the actual implementation in wire_gen.go.
//
// Returns:
//   - *server.Server: Fully configured HTTP server with all dependencies
//   - func(): Cleanup function to be called on shutdown
//   - error: Error if initialization fails
//
// Example:
//
//	srv, cleanup, err := InitializeApplication()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer cleanup()
func InitializeApplication() (*server.Server, func(), error) {
	wire.Build(app.ProviderSet)
	return nil, nil, nil
}

