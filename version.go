package pal

import (
	"fmt"
	"runtime/debug"
)

func Version() (string, error) {
	// something around this will be needed once we put in
	// place per-package caching.
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "", fmt.Errorf("couldn't read build info (are your running go test?)")
	}
	return fmt.Sprintf("%s %s %s\n", bi.Main.Path, bi.Main.Version, bi.Main.Sum), nil
}
