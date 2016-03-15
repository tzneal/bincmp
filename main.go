// symSizeComp compares symbol sizes between binaries
//
// Usage:
//
//	symSizeComp [options] oldBinary newBinary
//
// It runs nm to gather symbol sizes and report symbol by symbol differences
// as well as differences summarized by linker section.  It can also optionally
// use objdump to display a side-by-side disassembly of differing functions.
//
// The options are:
//
//	  -difference
//		sort output by the symbol size difference (default true)
//	  -disassemble
//		display disassembly of non-matching functions
//	  -exact
//		remove padding bytes from the symbol sizes of functions by examining disassembly
//	  -larger
//		only display larger symbols
//	  -pattern string
//		regular expression to match against symbols
//	  -relative
//		sort output by the relative symbol size difference
//	  -size
//		sort output by the new symbol size
//
//
// Example
//
// Defaut usage, print symbol size differences and summary
//
//	$ symSizeComp  ~/Projects/go/bin/go ~/Projects/goclean/bin/go
//	# symbol differences
//	-4415 runtime.pclntab 1474439 1478854 -0.298542%
//	-192 type..eq.net/url.URL 784 976 -19.672131%
//	-176 type..eq.net.netFD 464 640 -27.500000%
//	...
//	-16 type..eq.net.nssCriterion 272 288 -5.555556%
//	-16 type..eq.net/http.connectMethod 272 288 -5.555556%
//
//	# section differences
//	read-only data = -4511 bytes (-0.284264%)
//	global text (code) = -15904 bytes (-0.365848%)
//	Total difference 20415 bytes (-0.326807%)
//
//
// Filter on a single function and show the disassembly
//
//	$ symSizeComp --size --disassemble --pattern="type.*eq.*11.*float32" ~/Projects/go/bin/go ~/Projects/goclean/bin/go
//	# symbol differences
//	-16 type..eq.[11]float32 80 96 -16.666667%
//	  571950:    mov    0x8(%rsp),%rdi                      5734b0:    mov    0x8(%rsp),%rdi
//	  571955:    mov    0x10(%rsp),%rsi                     5734b5:    mov    0x10(%rsp),%rsi
//	  57195a:    xor    %eax,%eax                           5734ba:    xor    %eax,%eax
//	  57195c:    mov    $0xb,%rdx                           5734bc:    mov    $0xb,%rdx
//	  571963:    cmp    %rdx,%rax                           5734c3:    cmp    %rdx,%rax
//	  571966:    jge    571987 <type..eq.[11]float32+0x37>  5734c6:    jge    5734f3 <type..eq.[11]float32+0x43>
//	  571968:    lea    (%rdi,%rax,4),%rbx                  5734c8:    cmp    $0x0,%rdi
//	  57196c:    movss  (%rbx),%xmm0                        5734cc:    je     573503 <type..eq.[11]float32+0x53>
//	  571970:    lea    (%rsi,%rax,4),%rbx                  5734ce:    lea    (%rdi,%rax,4),%rbx
//	  571974:    movss  (%rbx),%xmm1                        5734d2:    movss  (%rbx),%xmm0
//	  571978:    ucomiss %xmm0,%xmm1                        5734d6:    cmp    $0x0,%rsi
//	  57197b:    jne    57198d <type..eq.[11]float32+0x3d>  5734da:    je     5734ff <type..eq.[11]float32+0x4f>
//	  57197d:    jp     57198d <type..eq.[11]float32+0x3d>  5734dc:    lea    (%rsi,%rax,4),%rbx
//	  57197f:    inc    %rax                                5734e0:    movss  (%rbx),%xmm1
//	...
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
	fileSize    int64
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
	exactSize       = flag.Bool("exact", false, "remove padding bytes from the symbol sizes of functions by examining disassembly")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "%s [options] oldBinary newBinary\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		return
	}

	oldBin := parseBinary(flag.Arg(0))
	newBin := parseBinary(flag.Arg(1))
	oldBin.printDiff(newBin)
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
	for i, j := 0, 0; i < len(sym1.code) || j < len(sym2.code); {
		if i < len(sym1.code) {
			fmt.Printf("%s", sym1.code[i])
			// pad to the same length so the rhs listing will be aligned
			printSpaces(sym1.maxLen - len(sym1.code[i]) + 1)
		} else {
			printSpaces(sym1.maxLen + 1)
		}
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
	binfo := &binaryInfo{}
	if st, err := os.Stat(fn); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "%s doesn't exist\n", fn)
		os.Exit(1)
	} else {
		binfo.fileSize = st.Size()
	}

	binfo.parse(fn)
	return binfo
}

type symDiff struct {
	old, new *symbol
}

// sizeDifference returns the difference in sizes between two symbols.
func (sd *symDiff) sizeDifference() int {
	return sd.new.size - sd.old.size
}

// pctDifference returns the relative size difference between two symbols.
func (sd *symDiff) pctDifference() float64 {
	return 100 * (float64(sd.new.size)/float64(sd.old.size) - 1)
}

func (bi *binaryInfo) symbolDiff(bi2 *binaryInfo) []*symDiff {
	ret := make([]*symDiff, 0)
	for name, old := range bi.symbols {
		if new, ok := bi2.symbols[name]; ok {
			if old.size == new.size {
				continue
			}
			if *onlyLarger && old.size > new.size {
				continue
			}
			ret = append(ret, &symDiff{old, new})
		}
	}
	return ret
}
func (bi *binaryInfo) hasDisassembly(sym string) bool {
	dis, ok := bi.disassembly[sym]
	if !ok {
		return false
	}
	return len(dis.code) > 0
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
	return s1.old.size < s2.old.size
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
	return s1.old.name < s2.old.name
}

// printDiff prints out differing symbols according to user flags, followed by a
// summary of the size differences of the different symbol sections.
func (bi *binaryInfo) printDiff(bi2 *binaryInfo) {
	symDiffs := bi.symbolDiff(bi2)
	// by default, sort by name
	sort.Sort(symbolSort{symDiffs, byName})

	// then use stable sorts by the user parameter to keep
	// equivalent symbols sorted by name
	if *sortDifference {
		sort.Stable(symbolSort{symDiffs, bySizeDiff})
	}
	if *sortSize {
		sort.Stable(symbolSort{symDiffs, bySize})
	}
	if *sortRelative {
		sort.Stable(symbolSort{symDiffs, byRelSizeDiff})
	}

	fmt.Printf("# symbol differences\n")
	for _, sym := range symDiffs {
		if *disasmFunctions && bi.hasDisassembly(sym.old.name) {
			fmt.Printf("%d %s %d %d %f%%\n", sym.sizeDifference(), sym.old.name, sym.old.size, sym.new.size, sym.pctDifference())
			s1Dis := bi.disassembly[sym.old.name]
			s2Dis := bi2.disassembly[sym.new.name]
			s1Dis.printDisasm(s2Dis)
			fmt.Println()
		} else if !*disasmFunctions {
			fmt.Printf("%d %s %d %d %f%%\n", sym.sizeDifference(), sym.old.name, sym.old.size, sym.new.size, sym.pctDifference())
		}
	}

	// print a summary of the section size differences from bi to bi2
	fmt.Printf("\n# file difference\n")
	fmt.Printf("%s %d\n", bi.filename, bi.fileSize)
	fmt.Printf("%s %d [%d bytes]\n", bi2.filename, bi2.fileSize, bi2.fileSize-bi.fileSize)
	fmt.Printf("\n# section differences\n")
	var biTotSz, bi2TotSz int
	for k := range bi.secSizes {
		szDiff := bi2.secSizes[k] - bi.secSizes[k]
		biTotSz += bi.secSizes[k]
		bi2TotSz += bi2.secSizes[k]
		if szDiff == 0 {
			continue
		}
		pct := 100 * (float64(bi2.secSizes[k])/float64(bi.secSizes[k]) - 1)
		fmt.Printf("%s = %d bytes (%f%%)\n", decodeType(k), szDiff, pct)
	}
	pct := 100 * (float64(bi2TotSz)/float64(biTotSz) - 1)
	fmt.Printf("Total difference %d bytes (%f%%)\n", bi2TotSz-biTotSz, pct)

}

// parse fills out the binaryInfo structure.
func (bi *binaryInfo) parse(fn string) {
	bi.filename = fn
	bi.parseNm()
	if *disasmFunctions || *exactSize {
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

var (
	// regexp for matching the start of disassembly for a symbol
	startDis = regexp.MustCompile("^[0-9a-f]+ <(.*?)>:$")
)

// findDisSymbolName matches a symbol name in the output of objdump.
func findDisSymbolName(c string) (string, bool) {
	match := startDis.FindStringSubmatch(c)
	if len(match) > 0 {
		return match[1], true
	}
	return "", false
}

func cleanDis(s string) string {
	code := strings.Replace(s, "\t", "    ", -1)
	// Remove comments
	if idx := strings.Index(code, "#"); idx != -1 {
		code = code[0:idx]
	}
	// Remove symbols (shortens the text)
	if idx := strings.Index(code, "<"); idx != -1 {
		code = code[0:idx]
	}
	return strings.TrimSpace(code)
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

	var lastSym string
	for scanner.Scan() {
		if pName, ok := findDisSymbolName(scanner.Text()); ok {
			lastSym = pName
			continue
		}

		if len(lastSym) > 0 {
			if _, ok := bi.disassembly[lastSym]; !ok {
				bi.disassembly[lastSym] = &dsym{}
			}
			sym := bi.disassembly[lastSym]

			// TODO: Parse the output of objdump
			code := cleanDis(scanner.Text())
			if len(code) == 0 {
				continue
			}
			sym.code = append(sym.code, code)
			if len(code) > sym.maxLen {
				sym.maxLen = len(code)
			}
		}
	}

	if *exactSize {
		for sn, sym := range bi.disassembly {
			for i := len(sym.code) - 1; i >= 0; i-- {
				if pb := paddingCnt(sym.code[i]); pb > 0 {
					bi.symbols[sn].size -= pb
				} else if i > 0 {
					code := bi.disassembly[sn].code
					bi.disassembly[sn].code = code[0 : i+1]
					break
				}
			}
		}
	}
}

func paddingCnt(c string) int {
	// TODO: Detect padding other than what the go compiler outputs on AMD64
	if strings.HasSuffix(c, "int3") {
		return 1
	}
	return 0
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
