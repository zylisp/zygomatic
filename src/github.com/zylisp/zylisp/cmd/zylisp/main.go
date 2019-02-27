/*
The ZYLISP command line REPL is compiled to `bin/zylisp`.
*/
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/zylisp/zcore"
	"github.com/zylisp/zylisp"
)

func usage(myflags *flag.FlagSet) {
	fmt.Printf("zylisp command line help:\n")
	myflags.PrintDefaults()
	os.Exit(1)
}

func main() {
	cfg := zcore.NewZlispConfig("zylisp")
	cfg.DefineFlags()
	err := cfg.Flags.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		usage(cfg.Flags)
	}

	if err != nil {
		panic(err)
	}
	err = cfg.ValidateConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "zylisp command line error: '%v'\n", err)
		usage(cfg.Flags)
	}

	// the library does all the heavy lifting.
	zylisp.ReplMain(cfg)
}
