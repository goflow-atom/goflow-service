// Package core provides core functionality and constants for the GoFlow Workflow Engine.
//
// This package contains version information, build metadata, and other core
// constants used throughout the application.
package core

import (
	"fmt"
	"runtime"
)

const (
	// Name is the application name
	Name = "GoFlow"

	// Version is the current version of the GoFlow Workflow Engine
	Version = "0.1.0"

	// Description is a brief description of the application
	Description = "Production-grade workflow orchestration system"

	// ModulePath is the Go module path for the project
	ModulePath = "github.com/goflow-atom/goflow-service"
)

// Build information (set via ldflags during build)
var (
	// BuildTime is the time when the binary was built
	BuildTime = "unknown"

	// GitCommit is the git commit hash of the build
	GitCommit = "unknown"

	// GitBranch is the git branch of the build
	GitBranch = "unknown"

	// GoVersion is the Go version used to build the binary
	GoVersion = runtime.Version()
)

// VersionInfo contains detailed version information
type VersionInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	ModulePath  string `json:"module_path"`
	BuildTime   string `json:"build_time"`
	GitCommit   string `json:"git_commit"`
	GitBranch   string `json:"git_branch"`
	GoVersion   string `json:"go_version"`
}

// GetVersionInfo returns detailed version information
//
// Returns:
//   - VersionInfo: A struct containing all version and build information
//
// Example:
//
//	info := core.GetVersionInfo()
//	fmt.Printf("GoFlow %s (commit: %s)\n", info.Version, info.GitCommit)
func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Name:        Name,
		Version:     Version,
		Description: Description,
		ModulePath:  ModulePath,
		BuildTime:   BuildTime,
		GitCommit:   GitCommit,
		GitBranch:   GitBranch,
		GoVersion:   GoVersion,
	}
}

// String returns a formatted version string
//
// Returns:
//   - string: A human-readable version string
//
// Example:
//
//	fmt.Println(core.GetVersionInfo().String())
//	// Output: GoFlow v0.1.0 (go1.21.0, commit: abc123, branch: main, built: 2025-10-23T10:00:00Z)
func (v VersionInfo) String() string {
	return fmt.Sprintf("%s v%s (go: %s, commit: %s, branch: %s, built: %s)",
		v.Name,
		v.Version,
		v.GoVersion,
		v.GitCommit,
		v.GitBranch,
		v.BuildTime,
	)
}

// ShortString returns a short version string
//
// Returns:
//   - string: A short version string containing only name and version
//
// Example:
//
//	fmt.Println(core.GetVersionInfo().ShortString())
//	// Output: GoFlow v0.1.0
func (v VersionInfo) ShortString() string {
	return fmt.Sprintf("%s v%s", v.Name, v.Version)
}
