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
	"math"

	"github.com/go-air/pal/mem"
	"github.com/go-air/pal/values"
)

// Pkg represents information to pass from
// a depended upon package to its importer.
type Pkg struct {
	Values   values.T
	Start    mem.T
	MemModel *mem.Model // provides mem.T operations
	SrcInfo  []SrcInfo  // indexed by mem.T
}

func NewPkg(pkgName string, vs values.T) *Pkg {
	mdl := mem.NewModel(vs)
	return &Pkg{
		Values:   vs,
		Start:    mem.T(1),
		MemModel: mdl,
		SrcInfo:  make([]SrcInfo, mdl.Len())}
}

func (pkg *Pkg) set(m mem.T, info *SrcInfo) {
	n := mem.T(uint32(cap(pkg.SrcInfo)))
	if m < n {
		pkg.SrcInfo[m] = *info
		return
	}
	if m > math.MaxUint32/2 {
		n = math.MaxUint32
	} else {
		for n <= m {
			n *= 2
		}
	}
	infos := make([]SrcInfo, n)
	copy(infos, pkg.SrcInfo)
	infos[m] = *info
	pkg.SrcInfo = infos
}
