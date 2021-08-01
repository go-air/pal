package results

import (
	"fmt"
	"runtime/debug"
	"sync"
)

type T struct {
	mu   sync.Mutex
	d    map[string]*Pkg
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

func NewT() (*T, error) {
	return &T{d: make(map[string]*Pkg)}, nil
}

func (t *T) AFact() {}

func (t *T) Lookup(pkgPath string) *Pkg {
	t.mu.Lock()
	defer t.mu.Unlock()
	res, _ := t.d[pkgPath]
	return res
}

func (t *T) Put(pkgName string, pkgR *Pkg) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.d[pkgName] = pkgR
	return nil
}
