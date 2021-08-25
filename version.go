// Copyright 2021 The pal authors (see AUTHORS)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pal

import (
	"fmt"
	"runtime/debug"
	"strings"
)

const modPath = "github.com/go-air/pal"

func tryMod(mod *debug.Module) string {
	if strings.HasPrefix(mod.Path, modPath) {
		return fmt.Sprintf("%s %s %s",
			mod.Path, mod.Version, mod.Sum)
	}
	return ""
}

func Version() (string, error) {
	// something around this will be needed once we put in
	// place per-package caching.
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "", fmt.Errorf("couldn't read build info (are you running go test?)")
	}
	s := tryMod(&bi.Main)
	if s != "" {
		return s, nil
	}
	for _, dep := range bi.Deps {
		s = tryMod(dep)
		if s != "" {
			return s, nil
		}
	}
	return "", fmt.Errorf("could not find %s in debug modules", modPath)
}
