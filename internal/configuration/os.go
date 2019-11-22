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
