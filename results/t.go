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

package results

import (
	"fmt"
	"runtime/debug"
	"sync"
)

type T struct {
	mu   sync.Mutex
	d    map[string]*ForPkg
	perm []int
}

const (
	myModuleName = "github.com/go-air/pal"
)

func init() {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		panic(fmt.Sprintf("couldn't read build info"))
	}
	fmt.Printf("bi.Main.Path: %s %s\n", bi.Main.Path, bi.Main.Version)
}

func New() (*T, error) {
	return &T{d: make(map[string]*ForPkg)}, nil
}

func (t *T) AFact() {}

func (t *T) Lookup(pkgPath string) *ForPkg {
	t.mu.Lock()
	defer t.mu.Unlock()
	res, _ := t.d[pkgPath]
	return res
}

func (t *T) Put(pkgName string, pkgR *ForPkg) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.d[pkgName] = pkgR
	return nil
}
