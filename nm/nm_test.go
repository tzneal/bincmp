package nm

import (
	"strings"
	"testing"
)

func TestListSymbols(t *testing.T) {
	inp := `0000000000a95240 0000000000000008 B encoding/xml.HTMLEntity
0000000000a877d8 0000000000000008 D encoding/xml.second
0000000000a95258 0000000000000008 B encoding/xml.tinfoMap
00000000008d2f18 0000000000000008 r $f64.0010000000000000
00000000008d2f20 0000000000000008 r $f64.3cb0000000000000
0000000000456730 0000000000000009 T runtime.prefetchnta
0000000000456700 0000000000000009 T runtime.prefetcht0
`
	exp := []Symbol{
		{"encoding/xml.HTMLEntity", SymbolTypeGlobalBSS, 8, 0xa95240},
		{"encoding/xml.second", SymbolTypeGlobalData, 8, 0xa877d8},
		{"encoding/xml.tinfoMap", SymbolTypeGlobalBSS, 8, 0xa95258},
		{"$f64.0010000000000000", SymbolTypeReadOnlyData, 8, 0x8d2f18},
		{"$f64.3cb0000000000000", SymbolTypeReadOnlyData, 8, 0x8d2f20},
		{"runtime.prefetchnta", SymbolTypeGlobalText, 9, 0x456730},
		{"runtime.prefetcht0", SymbolTypeGlobalText, 9, 0x456700}}

	syms, err := parseListSymbols(strings.NewReader(inp))
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	if len(syms) != len(exp) {
		t.Fatalf("expected %d syms, got %d", len(exp), len(syms))
	}
	for i := range syms {
		if syms[i] != exp[i] {
			t.Errorf("expected %s, got %s", exp[i], syms[i])
		}
	}
}
