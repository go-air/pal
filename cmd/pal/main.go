package main

import (
	"github.com/go-air/pal"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(pal.Analyzer)
}
