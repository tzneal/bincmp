package main

import (
	"bufio"
	"flag"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type symbols map[string]int64

func main() {
	flag.Parse()
	f1 := flag.Arg(0)
	f2 := flag.Arg(1)

	f1Sym := parseSyms(f1)
	f2Sym := parseSyms(f2)
	delta := int64(0)
	fmt.Println("# delta name sz1 sz2")
	for name, sz := range f1Sym {
		if sz2, ok := f2Sym[name]; ok {
			delete(f1Sym, name)
			delete(f2Sym, name)
			if sz == sz2 {
				continue
			}
			fmt.Printf("%d %s %d %d\n", sz-sz2, name, sz, sz2)
			delta += sz - sz2
		}
	}
	for name := range f1Sym {
		fmt.Printf("-%s\n", name)
	}
	for name := range f2Sym {
		fmt.Printf("+%s\n", name)
	}
	if delta > 0 {
		fmt.Printf("%s is bigger than %s [%d]\n", f1, f2, delta)
	} else if delta < 0 {
		fmt.Printf("%s is smaller than %s [%d]\n", f1, f2, delta)
	}

}

func run(args ...string) *bufio.Scanner {
	cmd := exec.Command(args[0], args[1:]...)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	return bufio.NewScanner(cmdReader)
}
func parseSyms(fn string) symbols {
	syms := symbols{}
	scanner := run("nm", "-S", "--size-sort", fn)
	for scanner.Scan() {
		line := strings.Fields(scanner.Text())
		// address size type name
		if len(line) != 4 {
			continue
		}
		sz, err := strconv.ParseInt(line[1], 16, 64)
		if err == nil {
			syms[line[3]] = sz
		}
	}
	return syms
}
