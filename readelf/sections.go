package readelf

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Section is a section of a binary extracted via readelf
type Section struct {
	Name    string
	Type    string
	Address int64
	Offset  int64
	Size    int64
	EntSize int64
}

func (s Section) IsEmpty() bool {
	return len(s.Name) == 0 && s.Size == 0
}

// ListSections parses the output of "readelf -s" to get section
// information.
func ListSections(filename string) ([]Section, error) {
	args := []string{"-S", filename}
	cmd := exec.Command("readelf", args...)
	p, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("running readelf %s: %s", args, err)
	}
	defer p.Close()
	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("error running readelf %s: %s", args, err)
	}
	return parseListSections(p)
}

func parseListSections(r io.Reader) ([]Section, error) {
	scanner := bufio.NewScanner(r)
	started := false
	// [Nr] Name Type Address Offset
	re1 := regexp.MustCompile(`\[\s*(\d+)\]\s+(\S+)?\s+(\S+)\s+([[:xdigit:]]+)\s+([[:xdigit:]]+)`)
	//	 Size              EntSize          Flags  Link  Info  Align
	re2 := regexp.MustCompile(`([[:xdigit:]]+)\s+([[:xdigit:]]+)\s+(\w+)?\s+(\d+)\s+(\d+)\s+(\d+)`)

	ret := []Section{}
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "Section Headers:") {
			started = true
			continue
		}
		if strings.HasPrefix(scanner.Text(), "Key to Flags:") {
			started = false
			continue
		}
		if !started {
			continue
		}
		line1Str := scanner.Text()
		if !scanner.Scan() {
			return nil, fmt.Errorf("expected matching line")
		}
		line2Str := scanner.Text()
		line1 := re1.FindStringSubmatch(line1Str)
		line2 := re2.FindStringSubmatch(line2Str)
		// skip the header
		if len(line1) == 0 {
			continue
		}
		if len(line1) != 6 {
			log.Printf("skipping bad readelf parse on line 1: %s", line1Str)
			continue
		}
		if len(line2) != 7 {
			log.Printf("skipping bad readelf parse on line 2: %s", line2Str)
			continue
		}

		// [Nr] Name Type Address Offset
		//	 Size              EntSize          Flags  Link  Info  Align
		s := Section{Name: line1[2],
			Type:    line1[3],
			Address: parseHex(line1[4]),
			Offset:  parseHex(line1[5]),
			Size:    parseHex(line2[1]),
			EntSize: parseHex(line2[2])}

		if s.Name == "" {
			continue
		}
		ret = append(ret, s)
	}
	return ret, nil
}

func parseHex(s string) int64 {
	i, err := strconv.ParseInt(s, 16, 64)
	if err != nil {
		log.Println("error parsing hex %s: %s", s, err)
		return 0
	}
	return i
}
