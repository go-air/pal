package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	"github.com/go-air/pal/internal/plain"
)

var outputHex = flag.Bool("x", false, "hex output")

func main() {
	flag.Parse()
	u := plain.Uint(0)
	for _, a := range flag.Args() {
		err := (&u).PlainDecode(bytes.NewBuffer([]byte(a)))
		if err != nil {
			fmt.Fprintf(os.Stderr, "err: %s\n", err)
			continue
		}
		if *outputHex {
			fmt.Printf("%s %x\n", a, u)
		} else {
			fmt.Printf("%s %d\n", a, u)
		}
	}
}
