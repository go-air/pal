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

import "github.com/go-air/pal/mem"

// PkgFact represents information to pass from
// a depended upon package to its importer.
type PkgFact struct {
	PackageName string
	MemModel    *mem.Model // provides mem.T operations
	SrcInfo     []SrcInfo  // indexed by mem.T
}

func (p *PkgFact) AFact() {}