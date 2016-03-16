package main

import "testing"

func TestPaddingCount(t *testing.T) {
	testData := []struct {
		code  string
		count int
	}{
		{"6ae32e:    int3", 1},
		{"6ae32e:    ret1", 0},
		{"6ae32e:    callq  4c7d00", 0}}
	for _, tc := range testData {
		if want, got := tc.count, paddingCnt(tc.code); want != got {
			t.Errorf("expected padding count for %s = %d, got %d", tc.code, tc.count, got)
		}
	}
}

func TestParseDisassemblySymbol(t *testing.T) {
	testData := []struct {
		line    string
		parsed  bool
		symName string
	}{
		{"0000000000401000 <main.init.1>:", true, "main.init.1"},
		{"401013:       48 83 ec 48             sub    $0x48,%rsp", false, ""},
		{"401496:       e8 e5 82 08 00          callq  489780 <runtime.writebarrierptr>", false, ""},
		{"0000000000401680 <main.addBuildFlags>:", true, "main.addBuildFlags"}}
	for _, tc := range testData {
		if gotSym, gotParsed := findDisSymbolName(tc.line); gotParsed != tc.parsed || gotSym != tc.symName {
			t.Errorf("expected symbol = %s for '%s', got %s", tc.symName, tc.line, gotSym)
		}
	}
}

func TestCleanDisassembly(t *testing.T) {
	testData := []struct {
		code    string
		cleaned string
	}{
		{"6ae32e:    int3  ", "6ae32e:    int3"},
		{"  401017:       48 8b 0d 12 a2 71 00    mov    0x71a212(%rip),%rcx        # b1b230 <main.cmdBuild>",
			"401017:       48 8b 0d 12 a2 71 00    mov    0x71a212(%rip),%rcx"},
		{"  401049:       0f 85 5b 01 00 00       jne    4011aa <main.init.1+0x1aa>",
			"401049:       0f 85 5b 01 00 00       jne    4011aa"}}
	for _, tc := range testData {
		if want, got := tc.cleaned, cleanDis(tc.code); want != got {
			t.Errorf("expected cleanDis(%s) to be `%s`,  got `%s`", tc.code, tc.cleaned, got)
		}
	}
}
