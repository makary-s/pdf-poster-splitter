package main

import (
	_ "embed"
	"strings"
)

//go:embed build_version
var buildVersionRaw string

func appVersion() string {
	return strings.TrimSpace(buildVersionRaw)
}
