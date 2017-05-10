package cmp

import (
	"sort"

	"github.com/tzneal/bincmp/nm"
	"github.com/tzneal/bincmp/readelf"
)

type symMap map[string]nm.Symbol

func uniqSymNames(a, b []nm.Symbol) (symMap, symMap, []string) {
	names := make(map[string]struct{}, len(a))
	aKnown := make(map[string]nm.Symbol, len(a))
	bKnown := make(map[string]nm.Symbol, len(b))
	for _, an := range a {
		aKnown[an.Name] = an
		names[an.Name] = struct{}{}
	}
	for _, bn := range b {
		bKnown[bn.Name] = bn
		names[bn.Name] = struct{}{}
	}
	ret := make([]string, 0, len(names))
	for n := range names {
		ret = append(ret, n)
	}
	sort.Strings(ret)
	return aKnown, bKnown, ret
}

type sectMap map[string]readelf.Section

func uniqSectNames(a, b []readelf.Section) (sectMap, sectMap, []string) {
	names := make(map[string]struct{}, len(a))
	aKnown := make(map[string]readelf.Section, len(a))
	bKnown := make(map[string]readelf.Section, len(b))
	for _, an := range a {
		aKnown[an.Name] = an
		names[an.Name] = struct{}{}
	}
	for _, bn := range b {
		bKnown[bn.Name] = bn
		names[bn.Name] = struct{}{}
	}
	ret := make([]string, 0, len(names))
	for n := range names {
		ret = append(ret, n)
	}
	sort.Strings(ret)
	return aKnown, bKnown, ret
}
