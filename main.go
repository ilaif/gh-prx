package main

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/ilaif/gh-prx/pkg/cmd"
)

var (
	version = "dev"
)

func main() {
	cmd.Execute(buildVersion(version))
}

func buildVersion(version string) string {
	result := version
	result = fmt.Sprintf("%s\ngoos: %s\ngoarch: %s", result, runtime.GOOS, runtime.GOARCH)
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Sum != "" {
		result = fmt.Sprintf("%s\nmodule version: %s, checksum: %s", result, info.Main.Version, info.Main.Sum)
	}

	return result
}
