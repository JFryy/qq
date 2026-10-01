package cli

import "runtime/debug"

// Version is injected at build time via:
//
//	-ldflags "-X github.com/JFryy/qq/cli.Version=v1.2.3"
var Version = ""

// resolveVersion returns the injected version, falling back to the module
// version recorded by `go install module@version`, then to "dev".
func resolveVersion() string {
	if Version != "" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}
