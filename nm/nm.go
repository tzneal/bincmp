package nm

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
)

// SymbolType is the type of symbol parsed from the nm output
//go:generate stringer -type=SymbolType
type SymbolType byte

const (
	SymbolTypeUnknown SymbolType = iota
	SymbolTypeBSS
	SymbolTypeGlobalBSS
	SymbolTypeData
	SymbolTypeGlobalData
	SymbolTypeText
	SymbolTypeGlobalText
	SymbolTypeReadOnlyData
	SymbolTypeGlobalReadOnlyData
)

type Symbol struct {
	Name  string
	Type  SymbolType
	Size  int64
	Value int64
}

func (s Symbol) IsEmpty() bool {
	return len(s.Name) == 0 && s.Size == 0
}

func (s Symbol) String() string {
	return fmt.Sprintf("<%s %s %d>", s.Name, s.Type, s.Size)
}

func ListSymbols(filename string) ([]Symbol, error) {
	args := []string{"-S", filename}
	cmd := exec.Command("nm", args...)
	p, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("running nm %s: %s", args, err)
	}
	defer p.Close()
	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("error running nm %s: %s", args, err)
	}
	return parseListSymbols(p)
}

func parseListSymbols(r io.Reader) ([]Symbol, error) {
	scanner := bufio.NewScanner(r)
	ret := []Symbol{}
	for scanner.Scan() {
		line := strings.Fields(scanner.Text())
		// format is "address size type name" with
		// - type being 1 character
		// - symbol possibly having spaces
		if len(line) < 4 || len(line[2]) != 1 {
			continue
		}
		sym, err := parseSymbol(line)
		if err != nil {
			return nil, fmt.Errorf("error parsing symbol: %s", err)
		}
		ret = append(ret, sym)
	}
	return ret, nil
}

// parseSymbol parses a line of output from nm and returns a symbol.
func parseSymbol(line []string) (Symbol, error) {
	// format is "address size type name"
	if len(line) < 4 || len(line[2]) != 1 {
		return Symbol{}, errors.New(fmt.Sprintf("unexpected format %v", line))
	}
	value, err := strconv.ParseInt(line[0], 16, 64)
	if err != nil {
		return Symbol{}, errors.New(fmt.Sprintf("couldn't parse value %s", line[0]))
	}
	size, err := strconv.ParseInt(line[1], 16, 64)
	if err != nil {
		return Symbol{}, errors.New(fmt.Sprintf("couldn't parse size %s", line[1]))
	}
	name := strings.Join(line[3:], " ")
	return Symbol{name, decodeType(line[2]), size, value}, nil
}

// decodeType maps section type characters to a more readable section name.
func decodeType(t string) SymbolType {
	switch t {
	case "b":
		return SymbolTypeBSS
	case "B":
		return SymbolTypeGlobalBSS
	case "d":
		return SymbolTypeData
	case "D":
		return SymbolTypeGlobalData
	case "t":
		return SymbolTypeText
	case "T":
		return SymbolTypeGlobalText
	case "r":
		return SymbolTypeReadOnlyData
	case "R":
		return SymbolTypeGlobalReadOnlyData
	default:
		return SymbolTypeUnknown
	}
}
