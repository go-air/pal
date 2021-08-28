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
	"github.com/go-air/pal/internal/plain"
	"sync"
)

type T[Index plain.Coder] struct {
	mu   sync.Mutex
	d    map[string]*PkgRes[Index]
	perm []int
}

// New generates a new results.T object for managing pointer analysis
// results.
func New[Index plain.Coder]() (*T[Index], error) {
	return &T[Index]{d: make(map[string]*PkgRes[Index])}, nil
}

// AFact satisfying golang.org/x/tools/go/analysis's Facts.
//
// Using this in that framework makes it analyse package dependencies before
// analyzing the respective package.
func (t *T[Index]) AFact() {}

func (t *T[Index]) Lookup(pkgPath string) *PkgRes[Index] {
	t.mu.Lock()
	defer t.mu.Unlock()
	res, _ := t.d[pkgPath]
	return res
}

func (t *T[Index]) Put(pkgName string, pkgR *PkgRes[Index]) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.d[pkgName] = pkgR
	return nil
}
