package configuration

import (
	"runtime"
)

// Windows boolean if running on windows
var Windows bool = runtime.GOOS == "windows"

// Macos boolean if running on windows
var Macos bool = runtime.GOOS == "darwin"

// Linux boolean if running on windows
var Linux bool = runtime.GOOS == "linux"

// AMD64 boolean if running on amd64 architecture
var AMD64 bool = runtime.GOARCH == "amd64"

// ARM64 boolean if running on arm64 architecture
var ARM64 bool = runtime.GOARCH == "arm64"
