package objdump

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type Function struct {
	Name string
	File string
	Asm  []Disasm
}
type Disasm struct {
	File   string
	Line   int64
	Offset int64
	Bin    string
	Asm    string
}

func (f Function) IsEmpty() bool {
	return len(f.Name) == 0
}

var errNotFound = errors.New("function not found")

func IsNotFound(e error) bool {
	return e == errNotFound
}

func DisassembleFunction(filename string, fn string) (Function, error) {
	args := []string{"tool", "objdump", "-s", "^" + regexp.QuoteMeta(fn) + "$", filename}
	cmd := exec.Command("go", args...)
	p, err := cmd.StdoutPipe()
	if err != nil {
		return Function{}, fmt.Errorf("running go %s: %s", args, err)
	}
	defer p.Close()
	if err = cmd.Start(); err != nil {
		return Function{}, fmt.Errorf("error running go %s: %s", args, err)
	}
	fns, err := parseDisassembly(p)
	if err != nil {
		return Function{}, err
	}
	if len(fns) == 1 {
		return fns[0], nil
	}

	if len(fns) == 0 {
		return Function{}, errNotFound
	}

	return Function{}, fmt.Errorf("found too many functions, got %d", len(fns))
}

func Disassemble(filename string) ([]Function, error) {
	args := []string{"tool", "objdump", filename}
	cmd := exec.Command("go", args...)
	p, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("running go %s: %s", args, err)
	}
	defer p.Close()
	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("error running go %s: %s", args, err)
	}
	return parseDisassembly(p)
}

func parseDisassembly(r io.Reader) ([]Function, error) {
	scanner := bufio.NewScanner(r)

	//TEXT strings.EqualFold(SB) /home/todd/Projects/go/src/strings/strings.go
	fnRe := regexp.MustCompile(`^TEXT ([^(]+)\(SB\) (.*)$`)
	//		tables.go:128   0x5997a0        4883ec30                SUBQ $0x30, SP
	asmRe := regexp.MustCompile(`\s+([^:]*):(\d*)\s+(0x[[:xdigit:]]+)\s+([[:xdigit:]]+)\s+(.*)$`)
	curFn := Function{}
	ret := []Function{}
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			if !curFn.IsEmpty() {
				ret = append(ret, curFn)
			}
			curFn = Function{}
			continue
		}

		// start of a function
		if strings.HasPrefix(line, "TEXT") {
			if !curFn.IsEmpty() {
				return nil, fmt.Errorf("new function without finishing last")
			}

			fields := fnRe.FindStringSubmatch(line)
			if len(fields) == 0 {
				return nil, fmt.Errorf("unable to parse function from %s", line)
			}
			curFn.Name = fields[1]
			curFn.File = fields[2]
			continue
		}

		if curFn.IsEmpty() {
			return nil, fmt.Errorf("got input %s with no function", line)
		}
		// disassembly of an existing function
		fields := asmRe.FindStringSubmatch(line)
		file := fields[1]
		lineNo := parseInt(fields[2], 10)
		off := parseInt(fields[3][2:], 16)
		bin := fields[4]
		asm := strings.TrimSpace(fields[5])

		curFn.Asm = append(curFn.Asm, Disasm{
			File:   file,
			Line:   lineNo,
			Offset: off,
			Bin:    bin,
			Asm:    asm})
	}
	if !curFn.IsEmpty() {
		ret = append(ret, curFn)
	}
	return ret, nil
}

func parseInt(x string, base int) int64 {
	v, _ := strconv.ParseInt(x, base, 64)
	return v
}
