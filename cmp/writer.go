package cmp

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/tzneal/bincmp/nm"
	"github.com/tzneal/bincmp/readelf"
)

// Writer is used to allow customizing the difference output
type Writer interface {
	StartSymbols()
	WriteSymbol(symA, symB nm.Symbol) error
	EndSymbols()

	StartSections()
	WriteSection(sectA, sectB readelf.Section) error
	EndSections()
}

var DefaultWriter Writer = &stdoutWriter{}

type stdoutWriter struct {
	w *tabwriter.Writer
}

func (s *stdoutWriter) StartSymbols() {
	s.w = tabwriter.NewWriter(os.Stdout, 2, 2, 2, ' ', 0)
	fmt.Fprintf(s.w, "delta\tname\told\tnew\n")
}

func (s *stdoutWriter) WriteSymbol(symA, symB nm.Symbol) error {
	if !symA.IsEmpty() && !symB.IsEmpty() {
		delta := symB.Size - symA.Size
		pct := (float64(symB.Size)/float64(symA.Size) - 1) * 100
		fmt.Fprintf(s.w, "%d\t%s\t%d\t%d\t%10.2f%%\n", delta, symA.Name, symA.Size, symB.Size, pct)
	}
	return nil
}

func (s *stdoutWriter) EndSymbols() {
	s.w.Flush()
	s.w = nil
}

func (s *stdoutWriter) StartSections() {
	s.w = tabwriter.NewWriter(os.Stdout, 2, 2, 2, ' ', 0)
	fmt.Fprintf(s.w, "delta\tname\told\tnew\n")
}

func (s *stdoutWriter) WriteSection(sectA, sectB readelf.Section) error {
	if !sectA.IsEmpty() && !sectB.IsEmpty() {
		delta := sectB.Size - sectA.Size
		pct := (float64(sectB.Size)/float64(sectA.Size) - 1) * 100
		fmt.Fprintf(s.w, "%d\t%s\t%d\t%d\t%10.2f%%\n", delta, sectA.Name, sectA.Size, sectB.Size, pct)
	}
	return nil
}

func (s *stdoutWriter) EndSections() {
	s.w.Flush()
	s.w = nil
}
