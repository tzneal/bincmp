package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/tzneal/bincmp/cmp"
)

func main() {
	pattern := flag.String("pattern", "", "regular expression to match against symbols")
	disassemble := flag.Bool("disassemble", false, "dump objdump disassembly")
	noColor := flag.Bool("no-color", false, "force disable of color output")
	forceColor := flag.Bool("color", false, "force color output, regardless of terminal")
	noSymTab := flag.Bool("no-symtab", false, "only show section size difs")

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

	switch {
	case *forceColor && *noColor:
		fmt.Fprint(os.Stderr, "--color and --no-color are incompatible")
		os.Exit(1)
	case *forceColor:
		color.NoColor = false
	case *noColor:
		color.NoColor = true
	}

	opts := cmp.Options{
		Pattern:     *pattern,
		Writer:      cmp.DefaultWriter,
		Disassemble: *disassemble,
	}
	cmp := cmp.NewComparer(flag.Arg(0), flag.Arg(1), opts)
	cmp.CompareFiles()
	fmt.Println()
	if !*noSymTab {
		cmp.CompareSymbols()
		fmt.Println()
	}
	cmp.CompareSections()
}
