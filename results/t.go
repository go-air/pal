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
	"sync"
)

type T struct {
	mu   sync.Mutex
	d    map[string]*PkgRes
	perm []int
}

// New generates a new results.T object for managing pointer analysis
// results.
func New() (*T, error) {
	return &T{d: make(map[string]*PkgRes)}, nil
}

// AFact satisfying golang.org/x/tools/go/analysis's Facts.
//
// Using this in that framework makes it analyse package dependencies before
// analyzing the respective package.
func (t *T) AFact() {}

func (t *T) Lookup(pkgPath string) *PkgRes {
	t.mu.Lock()
	defer t.mu.Unlock()
	res, _ := t.d[pkgPath]
	return res
}

func (t *T) Put(pkgName string, pkgR *PkgRes) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.d[pkgName] = pkgR
	return nil
}
