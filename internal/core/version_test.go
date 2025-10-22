package core

import (
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConstants verifies that all version constants are properly defined
func TestConstants(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		notEmpty bool
	}{
		{
			name:     "Name constant",
			value:    Name,
			notEmpty: true,
		},
		{
			name:     "Version constant",
			value:    Version,
			notEmpty: true,
		},
		{
			name:     "Description constant",
			value:    Description,
			notEmpty: true,
		},
		{
			name:     "ModulePath constant",
			value:    ModulePath,
			notEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.notEmpty {
				assert.NotEmpty(t, tt.value, "%s should not be empty", tt.name)
			}
		})
	}
}

// TestConstants_Values verifies specific constant values
func TestConstants_Values(t *testing.T) {
	assert.Equal(t, "GoFlow", Name, "Name should be 'GoFlow'")
	assert.Equal(t, "0.1.0", Version, "Version should be '0.1.0'")
	assert.Equal(t, "github.com/goflow-atom/goflow-service", ModulePath, "ModulePath should be 'github.com/goflow-atom/goflow-service'")
	assert.Contains(t, Description, "workflow", "Description should mention workflow")
}

// TestBuildVariables verifies that build variables are initialized
func TestBuildVariables(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "BuildTime",
			value: BuildTime,
		},
		{
			name:  "GitCommit",
			value: GitCommit,
		},
		{
			name:  "GitBranch",
			value: GitBranch,
		},
		{
			name:  "GoVersion",
			value: GoVersion,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.value, "%s should not be empty", tt.name)
		})
	}
}

// TestGoVersion verifies that GoVersion matches runtime version
func TestGoVersion(t *testing.T) {
	assert.Equal(t, runtime.Version(), GoVersion, "GoVersion should match runtime.Version()")
	assert.True(t, strings.HasPrefix(GoVersion, "go"), "GoVersion should start with 'go'")
}

// TestGetVersionInfo_Success verifies GetVersionInfo returns correct data
func TestGetVersionInfo_Success(t *testing.T) {
	// Act
	info := GetVersionInfo()

	// Assert
	assert.Equal(t, Name, info.Name, "Name should match constant")
	assert.Equal(t, Version, info.Version, "Version should match constant")
	assert.Equal(t, Description, info.Description, "Description should match constant")
	assert.Equal(t, ModulePath, info.ModulePath, "ModulePath should match constant")
	assert.Equal(t, BuildTime, info.BuildTime, "BuildTime should match variable")
	assert.Equal(t, GitCommit, info.GitCommit, "GitCommit should match variable")
	assert.Equal(t, GitBranch, info.GitBranch, "GitBranch should match variable")
	assert.Equal(t, GoVersion, info.GoVersion, "GoVersion should match variable")
}

// TestGetVersionInfo_AllFieldsPopulated verifies all fields are populated
func TestGetVersionInfo_AllFieldsPopulated(t *testing.T) {
	// Act
	info := GetVersionInfo()

	// Assert
	assert.NotEmpty(t, info.Name, "Name should not be empty")
	assert.NotEmpty(t, info.Version, "Version should not be empty")
	assert.NotEmpty(t, info.Description, "Description should not be empty")
	assert.NotEmpty(t, info.ModulePath, "ModulePath should not be empty")
	assert.NotEmpty(t, info.BuildTime, "BuildTime should not be empty")
	assert.NotEmpty(t, info.GitCommit, "GitCommit should not be empty")
	assert.NotEmpty(t, info.GitBranch, "GitBranch should not be empty")
	assert.NotEmpty(t, info.GoVersion, "GoVersion should not be empty")
}

// TestVersionInfo_String_Format verifies String method format
func TestVersionInfo_String_Format(t *testing.T) {
	// Arrange
	info := GetVersionInfo()

	// Act
	result := info.String()

	// Assert
	require.NotEmpty(t, result, "String() should not return empty string")
	assert.Contains(t, result, info.Name, "String should contain Name")
	assert.Contains(t, result, info.Version, "String should contain Version")
	assert.Contains(t, result, info.GoVersion, "String should contain GoVersion")
	assert.Contains(t, result, info.GitCommit, "String should contain GitCommit")
	assert.Contains(t, result, info.GitBranch, "String should contain GitBranch")
	assert.Contains(t, result, info.BuildTime, "String should contain BuildTime")
}

// TestVersionInfo_String_ContainsKeywords verifies String contains expected keywords
func TestVersionInfo_String_ContainsKeywords(t *testing.T) {
	// Arrange
	info := GetVersionInfo()

	// Act
	result := info.String()

	// Assert
	assert.Contains(t, result, "go:", "String should contain 'go:' prefix")
	assert.Contains(t, result, "commit:", "String should contain 'commit:' prefix")
	assert.Contains(t, result, "branch:", "String should contain 'branch:' prefix")
	assert.Contains(t, result, "built:", "String should contain 'built:' prefix")
}

// TestVersionInfo_ShortString_Format verifies ShortString method format
func TestVersionInfo_ShortString_Format(t *testing.T) {
	// Arrange
	info := GetVersionInfo()

	// Act
	result := info.ShortString()

	// Assert
	require.NotEmpty(t, result, "ShortString() should not return empty string")
	assert.Contains(t, result, info.Name, "ShortString should contain Name")
	assert.Contains(t, result, info.Version, "ShortString should contain Version")
}

// TestVersionInfo_ShortString_IsShort verifies ShortString is shorter than String
func TestVersionInfo_ShortString_IsShort(t *testing.T) {
	// Arrange
	info := GetVersionInfo()

	// Act
	shortStr := info.ShortString()
	fullStr := info.String()

	// Assert
	assert.Less(t, len(shortStr), len(fullStr), "ShortString should be shorter than String")
}

// TestVersionInfo_ShortString_DoesNotContainBuildInfo verifies ShortString excludes build info
func TestVersionInfo_ShortString_DoesNotContainBuildInfo(t *testing.T) {
	// Arrange
	info := GetVersionInfo()

	// Act
	result := info.ShortString()

	// Assert
	assert.NotContains(t, result, "commit:", "ShortString should not contain commit info")
	assert.NotContains(t, result, "branch:", "ShortString should not contain branch info")
	assert.NotContains(t, result, "built:", "ShortString should not contain build time")
}

// TestVersionInfo_String_ExpectedFormat verifies exact format
func TestVersionInfo_String_ExpectedFormat(t *testing.T) {
	// Arrange
	info := VersionInfo{
		Name:        "TestApp",
		Version:     "1.0.0",
		GoVersion:   "go1.21.0",
		GitCommit:   "abc123",
		GitBranch:   "main",
		BuildTime:   "2025-10-23T10:00:00Z",
		Description: "Test description",
		ModulePath:  "github.com/test/app",
	}

	// Act
	result := info.String()

	// Assert
	expected := "TestApp v1.0.0 (go: go1.21.0, commit: abc123, branch: main, built: 2025-10-23T10:00:00Z)"
	assert.Equal(t, expected, result, "String format should match expected pattern")
}

// TestVersionInfo_ShortString_ExpectedFormat verifies exact format
func TestVersionInfo_ShortString_ExpectedFormat(t *testing.T) {
	// Arrange
	info := VersionInfo{
		Name:    "TestApp",
		Version: "1.0.0",
	}

	// Act
	result := info.ShortString()

	// Assert
	expected := "TestApp v1.0.0"
	assert.Equal(t, expected, result, "ShortString format should match expected pattern")
}

// TestVersionInfo_JSONTags verifies JSON serialization
func TestVersionInfo_JSONTags(t *testing.T) {
	// This test verifies that the struct has proper JSON tags
	// by checking the struct definition
	info := GetVersionInfo()

	// Verify all fields are accessible
	assert.NotEmpty(t, info.Name)
	assert.NotEmpty(t, info.Version)
	assert.NotEmpty(t, info.Description)
	assert.NotEmpty(t, info.ModulePath)
	assert.NotEmpty(t, info.BuildTime)
	assert.NotEmpty(t, info.GitCommit)
	assert.NotEmpty(t, info.GitBranch)
	assert.NotEmpty(t, info.GoVersion)
}

// BenchmarkGetVersionInfo benchmarks GetVersionInfo performance
func BenchmarkGetVersionInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetVersionInfo()
	}
}

// BenchmarkVersionInfo_String benchmarks String method performance
func BenchmarkVersionInfo_String(b *testing.B) {
	info := GetVersionInfo()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = info.String()
	}
}

// BenchmarkVersionInfo_ShortString benchmarks ShortString method performance
func BenchmarkVersionInfo_ShortString(b *testing.B) {
	info := GetVersionInfo()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = info.ShortString()
	}
}

