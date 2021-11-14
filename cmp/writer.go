package cmp

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/tzneal/bincmp/nm"
	"github.com/tzneal/bincmp/objdump"
	"github.com/tzneal/bincmp/readelf"
)

// Writer is used to allow customizing the difference output
type Writer interface {
	StartFiles(a, b os.FileInfo) error

	StartSymbols()
	WriteSymbol(symA, symB nm.Symbol) error
	WriteDisassembly(fnA, fnB objdump.Function) error
	EndSymbols()

	StartSections()
	WriteSection(sectA, sectB readelf.Section) error
	EndSections()
}

var DefaultWriter Writer = &stdoutWriter{}

const MaxSymLen = 60

type stdoutWriter struct {
	w      *tabwriter.Writer
	totals [3]int64
}

func (s *stdoutWriter) StartFiles(a, b os.FileInfo) error {
	w := tabwriter.NewWriter(os.Stdout, 2, 2, 2, ' ', 0)
	defer w.Flush()
	fmt.Fprintf(w, "binary\tdelta\told\tnew\n")
	delta := b.Size() - a.Size()
	pct := (float64(b.Size())/float64(a.Size()) - 1) * 100
	fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%10.2f%%\n", b.Name(), delta, a.Size(), b.Size(), pct)
	return nil
}

func (s *stdoutWriter) StartSymbols() {
	s.w = tabwriter.NewWriter(os.Stdout, 2, 2, 2, ' ', 0)
	fmt.Fprintf(s.w, "symbol name\tdelta\told\tnew\n")
	s.totals = [3]int64{}
}

func (s *stdoutWriter) WriteDisassembly(fnA, fnB objdump.Function) error {
	// prepare for next call to write symbol
	s.w.Flush()
	s.w = tabwriter.NewWriter(os.Stdout, 2, 2, 2, ' ', 0)

	tw := tabwriter.NewWriter(os.Stdout, 2, 2, 2, ' ', 0)
	defer tw.Flush()
	n := len(fnA.Asm)
	if n < len(fnB.Asm) {
		n = len(fnB.Asm)
	}
	for i := 0; i < n; i++ {
		aAsm := ""
		aOff := ""
		if i < len(fnA.Asm) {
			aAsm = fnA.Asm[i].Asm
			aOff = fmt.Sprintf("0x%x", fnA.Asm[i].Offset)
		}
		bAsm := ""
		bOff := ""
		if i < len(fnB.Asm) {
			bAsm = fnB.Asm[i].Asm
			bOff = fmt.Sprintf("0x%x", fnB.Asm[i].Offset)
		}

		diff := ""

		// have to do it this way instead of a no-op color function, or
		// else tabwriter gets confused on the character count
		mark := color.New(color.FgHiWhite).SprintFunc()
		hl := color.New(color.FgHiWhite).SprintFunc()
		if aAsm != bAsm {
			diff = "!"
			mark = color.New(color.FgYellow).SprintFunc()
			hl = color.New(color.FgHiGreen).SprintFunc()
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", aOff, aAsm, mark(diff), bOff, hl(bAsm))
	}
	fmt.Fprintf(tw, "\n")
	return nil
}

func (s *stdoutWriter) WriteSymbol(symA, symB nm.Symbol) error {
	// If it's a symbol that is only present in A or B, we need to
	// pick a non-empty name here (otherwise we would see an empty name
	// in a report, which is not helpful).
	var symName string
	if symA.Name != "" {
		symName = symA.Name
	} else if symB.Name != "" {
		symName = symB.Name
	} else {
		symName = "<?>" // Should never happen
	}
	if len(symName) > MaxSymLen {
		symName = symName[0:MaxSymLen/2] + "..." + symName[len(symName)-MaxSymLen/2-3:]
	}
	if !symA.IsEmpty() && !symB.IsEmpty() {
		delta := symB.Size - symA.Size
		pct := (float64(symB.Size)/float64(symA.Size) - 1) * 100
		fmt.Fprintf(s.w, "%s\t%d\t%d\t%d\t%10.2f%%\n", symName, delta, symA.Size, symB.Size, pct)
		s.totals[0] += delta
		s.totals[1] += symA.Size
		s.totals[2] += symB.Size
		return nil
	} else if !symA.IsEmpty() {
		delta := -symA.Size
		fmt.Fprintf(s.w, "%s\t%d\t%d\t\n", symName, delta, symA.Size)
		s.totals[0] += delta
		s.totals[1] += symA.Size
	} else if !symB.IsEmpty() {
		delta := symB.Size
		fmt.Fprintf(s.w, "%s\t%d\t\t%d\n", symName, delta, symB.Size)
		s.totals[0] += delta
		s.totals[2] += symB.Size
	}
	return nil
}

func (s *stdoutWriter) EndSymbols() {
	pct := (float64(s.totals[2])/float64(s.totals[1]) - 1) * 100
	fmt.Fprintf(s.w, "total\t%d\t%d\t%d\t%10.2f%%\n", s.totals[0], s.totals[1], s.totals[2], pct)
	s.w.Flush()
	s.w = nil
}

func (s *stdoutWriter) StartSections() {
	s.w = tabwriter.NewWriter(os.Stdout, 2, 2, 2, ' ', 0)
	fmt.Fprintf(s.w, "name\tdelta\told\tnew\n")
	s.totals = [3]int64{}
}

func (s *stdoutWriter) WriteSection(sectA, sectB readelf.Section) error {
	if !sectA.IsEmpty() && !sectB.IsEmpty() {
		delta := sectB.Size - sectA.Size
		pct := (float64(sectB.Size)/float64(sectA.Size) - 1) * 100
		fmt.Fprintf(s.w, "%s\t%d\t%d\t%d\t%10.2f%%\n", sectA.Name, delta, sectA.Size, sectB.Size, pct)
		s.totals[0] += delta
		s.totals[1] += sectA.Size
		s.totals[2] += sectB.Size
	} else if !sectA.IsEmpty() {
		delta := -sectA.Size
		fmt.Fprintf(s.w, "%s\t%d\t%d\t\n", sectA.Name, delta, sectA.Size)
		s.totals[0] += delta
		s.totals[1] += sectA.Size
	} else if !sectB.IsEmpty() {
		delta := sectB.Size
		fmt.Fprintf(s.w, "%s\t%d\t\t%d\n", sectB.Name, delta, sectB.Size)
		s.totals[0] += delta
		s.totals[2] += sectB.Size
	}
	return nil
}

func (s *stdoutWriter) EndSections() {
	pct := (float64(s.totals[2])/float64(s.totals[1]) - 1) * 100
	fmt.Fprintf(s.w, "total\t%d\t%d\t%d\t%10.2f%%\n", s.totals[0], s.totals[1], s.totals[2], pct)
	s.w.Flush()
	s.w = nil
}
