package main

import (
	"fmt"
	"runtime"
)

// Version information set by build flags
var (
	Version   = "dev"          // Set by -ldflags "-X main.Version=..."
	BuildTime = "unknown"      // Set by -ldflags "-X main.BuildTime=..."
	GitCommit = "unknown"      // Set by -ldflags "-X main.GitCommit=..."
)

// VersionInfo returns formatted version information
func VersionInfo() string {
	return fmt.Sprintf(`MCP MindPort %s
Build Time: %s
Git Commit: %s
Go Version: %s
OS/Arch:    %s/%s`, 
		Version, 
		BuildTime, 
		GitCommit, 
		runtime.Version(), 
		runtime.GOOS, 
		runtime.GOARCH,
	)
}

// ShortVersion returns just the version string
func ShortVersion() string {
	return Version
}

// IsDevBuild returns true if this is a development build
func IsDevBuild() bool {
	return Version == "dev"
}