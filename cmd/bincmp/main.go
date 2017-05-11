package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tzneal/bincmp/cmp"
)

func main() {
	pattern := flag.String("pattern", "", "regular expression to match against symbols")
	disassemble := flag.Bool("disassemble", false, "dump objdump disassembly")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "%s [options] bin1 bin2\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		return
	}

	opts := cmp.Options{
		Pattern:     *pattern,
		Writer:      cmp.DefaultWriter,
		Disassemble: *disassemble,
	}
	cmp := cmp.NewComparer(flag.Arg(0), flag.Arg(1), opts)
	cmp.CompareFiles()
	fmt.Println()
	cmp.CompareSymbols()
	fmt.Println()
	cmp.CompareSections()
}
