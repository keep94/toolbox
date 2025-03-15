// Package build extracts version information from current binary.
package build

import (
	"runtime/debug"
	"strings"
)

// MainVersion returns the module version of the binary's main package
func MainVersion() (version string, ok bool) {
	bi, biok := debug.ReadBuildInfo()
	if !biok {
		return
	}
	return bi.Main.Version, true
}

// BuildId tries to extract the first 7 digits of the git commit checksum
// from the version string. If it cannot do this, it just returns the
// version string. If version string ends with "+dirty" then the return
// value will also end with "+dirty"
func BuildId(version string) string {
	versionParts := strings.Split(version, "-")
	if len(versionParts) != 3 {
		return version
	}
	version = versionParts[2]
	var dirty bool
	if strings.HasSuffix(version, "+dirty") {
		dirty = true
		version = version[:len(version)-6]
	}
	if len(version) > 7 {
		version = version[:7]
	}
	if dirty {
		return version + "+dirty"
	}
	return version
}
