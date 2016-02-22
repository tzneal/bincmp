package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

type symbol struct {
	name string
	typ  string
	size int
}

type binaryInfo struct {
	filename    string
	symbols     map[string]*symbol // symbol informatoin, indexed by symbol name
	disassembly map[string]*dsym   // disassembly of functions, indexed by symbol name
	secSizes    map[string]int     // section sizes, indexed by section symbol as reported by nm
}
type dsym struct {
	code   []string
	maxLen int
}

var (
	disasmFunctions = flag.Bool("disassemble", false, "display disassembly of non-matching functions")
	onlyLarger      = flag.Bool("larger", false, "only display larger symbols")
	sortSize        = flag.Bool("size", false, "sort output by the new symbol size")
	pattern         = flag.String("pattern", "", "regular expression to match against symbols")
	sortDifference  = flag.Bool("difference", true, "sort output by the symbol size difference")
	sortRelative    = flag.Bool("relative", false, "sort output by the relative symbol size difference")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "%s [options] newBinary oldBinary\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		return
	}

	f1Bin := parseBinary(flag.Arg(0))
	f2Bin := parseBinary(flag.Arg(1))
	f1Bin.printDiff(f2Bin)
}

// run executes a process and returns a scanner that allows parsing stdout line
// by line.
func run(args ...string) *bufio.Scanner {
	cmd := exec.Command(args[0], args[1:]...)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error executing command: %v\n", args)
		os.Exit(1)
	}
	err = cmd.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error starting command: %v\n", args)
		os.Exit(1)
	}

	return bufio.NewScanner(cmdReader)
}

// parseSymbol parses a line of output from nm and returns a symbol.
func parseSymbol(line []string) (*symbol, error) {
	// format is "address size type name"
	if len(line) < 4 || len(line[2]) != 1 {
		return nil, errors.New(fmt.Sprintf("unexpected format %v", line))
	}
	size, err := strconv.ParseInt(line[1], 16, 64)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("couldn't parse size %s", line[1]))
	}
	name := strings.Join(line[3:], " ")
	return &symbol{name, line[2], int(size)}, nil
}

// printDisasm prints out a side by side listing of the disassembly of a function.
func (sym1 *dsym) printDisasm(sym2 *dsym) {
	if sym1 == nil || sym2 == nil {
		return
	}
	for i, j := 0, 0; i < len(sym1.code) && j < len(sym2.code); {
		if i < len(sym1.code) {
			fmt.Printf("%s", sym1.code[i])
		}
		// pad to the same length so the rhs listing will be aligned
		printSpaces(sym1.maxLen - len(sym1.code[i]))
		if j < len(sym2.code) {
			fmt.Printf("%s", sym2.code[j])
		}
		fmt.Printf("\n")

		i++
		j++
	}
}

func printSpaces(n int) {
	for i := 0; i < n; i++ {
		fmt.Print(" ")
	}
}

func parseBinary(fn string) *binaryInfo {
	if _, err := os.Stat(fn); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "%s doesn't exist\n", fn)
		os.Exit(1)
	}

	binfo := &binaryInfo{}
	binfo.parse(fn)
	return binfo
}

type symDiff struct {
	sym1, sym2 *symbol
}

// sizeDifference returns the difference in sizes between two symbols.
func (sd *symDiff) sizeDifference() int {
	return sd.sym1.size - sd.sym2.size
}

// pctDifference returns the relative size difference between two symbols.
func (sd *symDiff) pctDifference() float64 {
	return 100 * (float64(sd.sym1.size)/float64(sd.sym2.size) - 1)
}

func (bi *binaryInfo) symbolDiff(bi2 *binaryInfo) []*symDiff {
	ret := make([]*symDiff, 0)
	for name, sym1 := range bi.symbols {
		if sym2, ok := bi2.symbols[name]; ok {
			if sym1.size == sym2.size {
				continue
			}
			if *onlyLarger && sym1.size < sym2.size {
				continue
			}
			ret = append(ret, &symDiff{sym1, sym2})
		}
	}
	return ret
}

type symbolSort struct {
	syms []*symDiff
	by   func(sym1, sym2 *symDiff) bool
}

func (s symbolSort) Len() int {
	return len(s.syms)
}

func (s symbolSort) Swap(i, j int) {
	s.syms[i], s.syms[j] = s.syms[j], s.syms[i]
}

func (s symbolSort) Less(i, j int) bool {
	return s.by(s.syms[i], s.syms[j])
}

// bySize is used to sort symbol differences by the size of the new symbol.
func bySize(s1, s2 *symDiff) bool {
	return s1.sym1.size < s2.sym1.size
}

// bySize is used to sort symbol differences by the absolute size difference.
// If the absolute difference is equal, symbols are sorted by relative size
// difference.
func bySizeDiff(s1, s2 *symDiff) bool {
	s1Sz := s1.sizeDifference()
	s2Sz := s2.sizeDifference()
	if s1Sz != s2Sz {
		return s1Sz < s2Sz
	}
	return s1.pctDifference() < s2.pctDifference()
}

// byRelSize is used to sort symbol differences by the relative size
// difference.  If the relative difference is equal, symbols are sorted by
// absolute size difference.
func byRelSizeDiff(s1, s2 *symDiff) bool {
	s1Pct := s1.pctDifference()
	s2Pct := s2.pctDifference()
	if s1Pct != s2Pct {
		return s1Pct < s2Pct
	}
	return s1.sizeDifference() < s2.sizeDifference()
}

// byName is used to sort symbol differences by the symbol name.
func byName(s1, s2 *symDiff) bool {
	return s1.sym1.name < s2.sym1.name
}

// printDiff prints out differing symbols according to user flags, followed by a
// summary of the size differences of the different symbol sections.
func (bi *binaryInfo) printDiff(bi2 *binaryInfo) {
	symDiffs := bi.symbolDiff(bi2)
	// by default, sort by name
	sort.Sort(symbolSort{symDiffs, byName})

	// then use stable sorts by the user parameter to keep
	// equivalent symbols sorted by name
	if *sortSize {
		sort.Stable(symbolSort{symDiffs, bySize})
		*sortDifference = false
		*sortRelative = false
	}
	if *sortDifference {
		sort.Stable(symbolSort{symDiffs, bySizeDiff})
		*sortRelative = false
	}
	if *sortRelative {
		sort.Stable(symbolSort{symDiffs, byRelSizeDiff})
	}

	fmt.Printf("# symbol differences\n")
	for _, sym := range symDiffs {
		fmt.Printf("%d %s %d %d %f%%\n", sym.sizeDifference(), sym.sym1.name, sym.sym1.size, sym.sym2.size, sym.pctDifference())
		if *disasmFunctions {
			s1Dis := bi.disassembly[sym.sym1.name]
			s2Dis := bi2.disassembly[sym.sym2.name]
			s1Dis.printDisasm(s2Dis)
		}
	}

	// print a summary of the section size differences from bi to bi2
	fmt.Printf("\n# section differences\n")
	var biTotSz, bi2TotSz int
	for k := range bi.secSizes {
		szDiff := bi.secSizes[k] - bi2.secSizes[k]
		biTotSz += bi.secSizes[k]
		bi2TotSz += bi2.secSizes[k]
		if szDiff == 0 {
			continue
		}
		pct := 100 * (float64(bi.secSizes[k])/float64(bi2.secSizes[k]) - 1)
		fmt.Printf("%s = %d bytes (%f%%)\n", decodeType(k), szDiff, pct)
	}
	pct := 100 * (float64(biTotSz)/float64(bi2TotSz) - 1)
	fmt.Printf("Total difference %d bytes (%f%%)\n", biTotSz-bi2TotSz, pct)

}

// parse fills out the binaryInfo structure.
func (bi *binaryInfo) parse(fn string) {
	bi.filename = fn
	bi.parseNm()
	if *disasmFunctions {
		bi.parseObjdump()
	}
}

// parseNm is used to parse the output of the nm command, determining the name,
// size and section type of symbols.
func (bi *binaryInfo) parseNm() {
	bi.symbols = make(map[string]*symbol)
	bi.secSizes = make(map[string]int)
	var scanner *bufio.Scanner
	if runtime.GOOS == "darwin" {
		scanner = run("gnm", "-S", "--size-sort", bi.filename)
	} else {
		scanner = run("nm", "-S", "--size-sort", bi.filename)
	}
	var re *regexp.Regexp
	if len(*pattern) > 0 {
		re = regexp.MustCompile(*pattern)
	}
	for scanner.Scan() {
		line := strings.Fields(scanner.Text())
		// format is "address size type name" with
		// - type being 1 character
		// - symbol possibly having spaces
		if len(line) < 4 || len(line[2]) != 1 {
			continue
		}
		sym, err := parseSymbol(line)
		// match against the user pattern
		if re != nil && !re.MatchString(sym.name) {
			continue
		}
		if err == nil {
			bi.symbols[sym.name] = sym
			bi.secSizes[sym.typ] += sym.size
		} else {
			fmt.Fprintf(os.Stderr, "error parsing symbol: %s\n", err)
		}
	}
}

// parseObjDump is used to parse the output of the objdump -d command and find
// the disassembly of function symbols.
func (bi *binaryInfo) parseObjdump() {
	var scanner *bufio.Scanner
	if runtime.GOOS == "darwin" {
		scanner = run("gobjdump", "-d", "--no-show-raw-insn", bi.filename)
	} else {
		scanner = run("objdump", "-d", "--no-show-raw-insn", bi.filename)
	}
	bi.disassembly = make(map[string]*dsym)

	// regexp for matching the start of disassembly for a symbol
	startDis, err := regexp.Compile("^[0-9a-f]+ <(.*?)>:$")
	if err != nil {
		fmt.Fprintf(os.Stderr, "bad regexp\n")
		os.Exit(1)
	}

	var lastSym string
	for scanner.Scan() {
		match := startDis.FindStringSubmatch(scanner.Text())
		if len(match) > 0 {
			lastSym = match[1]
			continue
		}
		if len(lastSym) > 0 {
			if _, ok := bi.disassembly[lastSym]; !ok {
				bi.disassembly[lastSym] = &dsym{}
			}
			sym := bi.disassembly[lastSym]

			// TODO: Parse the output of objdump
			code := strings.Replace(scanner.Text(), "\t", "    ", -1)
			sym.code = append(sym.code, code)
			if len(code) > sym.maxLen {
				sym.maxLen = len(code)
			}
		}
	}
}

// decodeType maps section type characters to a more readable section name.
func decodeType(t string) string {
	switch t {
	case "b":
		return "bss"
	case "B":
		return "global bss"
	case "d":
		return "data"
	case "D":
		return "global data"
	case "t":
		return "text (code)"
	case "T":
		return "global text (code)"
	case "r":
		return "read-only data"
	case "R":
		return "global read-only data"
	default:
		fmt.Fprintf(os.Stderr, "unknown section symbol %s", t)
		return "unknown"
	}
}
