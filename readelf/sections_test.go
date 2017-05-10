package readelf

import (
	"strings"
	"testing"
)

func TestListSections(t *testing.T) {
	inp := `There are 29 section headers, starting at offset 0x1e738:

Section Headers:
  [Nr] Name              Type             Address           Offset
       Size              EntSize          Flags  Link  Info  Align
  [ 0]                   NULL             0000000000000000  00000000
       0000000000000000  0000000000000000           0     0     0
  [ 1] .interp           PROGBITS         0000000000400238  00000238
       000000000000001c  0000000000000000   A       0     0     1
  [ 2] .note.ABI-tag     NOTE             0000000000400254  00000254
       0000000000000020  0000000000000000   A       0     0     4
  [ 3] .note.gnu.build-i NOTE             0000000000400274  00000274
       0000000000000024  0000000000000000   A       0     0     4
  [ 4] .gnu.hash         GNU_HASH         0000000000400298  00000298
       00000000000000c0  0000000000000000   A       5     0     8
  [ 5] .dynsym           DYNSYM           0000000000400358  00000358
       0000000000000cd8  0000000000000018   A       6     1     8
  [ 6] .dynstr           STRTAB           0000000000401030  00001030
       00000000000005dc  0000000000000000   A       0     0     1
  [ 7] .gnu.version      VERSYM           000000000040160c  0000160c
       0000000000000112  0000000000000002   A       5     0     2
  [ 8] .gnu.version_r    VERNEED          0000000000401720  00001720
       0000000000000070  0000000000000000   A       6     1     8
  [ 9] .rela.dyn         RELA             0000000000401790  00001790
       00000000000000a8  0000000000000018   A       5     0     8
  [10] .rela.plt         RELA             0000000000401838  00001838
       0000000000000a80  0000000000000018  AI       5    24     8
  [11] .init             PROGBITS         00000000004022b8  000022b8
       000000000000001a  0000000000000000  AX       0     0     4
  [12] .plt              PROGBITS         00000000004022e0  000022e0
       0000000000000710  0000000000000010  AX       0     0     16
  [13] .plt.got          PROGBITS         00000000004029f0  000029f0
       0000000000000008  0000000000000000  AX       0     0     8
  [14] .text             PROGBITS         0000000000402a00  00002a00
       0000000000011289  0000000000000000  AX       0     0     16
  [15] .fini             PROGBITS         0000000000413c8c  00013c8c
       0000000000000009  0000000000000000  AX       0     0     4
  [16] .rodata           PROGBITS         0000000000413ca0  00013ca0
       00000000000069b4  0000000000000000   A       0     0     32
  [17] .eh_frame_hdr     PROGBITS         000000000041a654  0001a654
       000000000000080c  0000000000000000   A       0     0     4
  [18] .eh_frame         PROGBITS         000000000041ae60  0001ae60
       0000000000002c84  0000000000000000   A       0     0     8
  [19] .init_array       INIT_ARRAY       000000000061de00  0001de00
       0000000000000008  0000000000000000  WA       0     0     8
  [20] .fini_array       FINI_ARRAY       000000000061de08  0001de08
       0000000000000008  0000000000000000  WA       0     0     8
  [21] .jcr              PROGBITS         000000000061de10  0001de10
       0000000000000008  0000000000000000  WA       0     0     8
  [22] .dynamic          DYNAMIC          000000000061de18  0001de18
       00000000000001e0  0000000000000010  WA       6     0     8
  [23] .got              PROGBITS         000000000061dff8  0001dff8
       0000000000000008  0000000000000008  WA       0     0     8
  [24] .got.plt          PROGBITS         000000000061e000  0001e000
       0000000000000398  0000000000000008  WA       0     0     8
  [25] .data             PROGBITS         000000000061e3a0  0001e3a0
       0000000000000260  0000000000000000  WA       0     0     32
  [26] .bss              NOBITS           000000000061e600  0001e600
       0000000000000d68  0000000000000000  WA       0     0     32
  [27] .gnu_debuglink    PROGBITS         0000000000000000  0001e600
       0000000000000034  0000000000000000           0     0     1
  [28] .shstrtab         STRTAB           0000000000000000  0001e634
       0000000000000102  0000000000000000           0     0     1
Key to Flags:
  W (write), A (alloc), X (execute), M (merge), S (strings), l (large)
  I (info), L (link order), G (group), T (TLS), E (exclude), x (unknown)
  O (extra OS processing required) o (OS specific), p (processor specific)
`
	r := strings.NewReader(inp)
	sects, err := parseListSections(r)
	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}
	if len(sects) != 28 {
		t.Errorf("expected 28 sections, got %d", len(sects))
	}
	//   [ 1] .interp           PROGBITS         0000000000400238  00000238
	//  000000000000001c  0000000000000000   A       0     0     1
	exp := Section{Name: ".interp",
		Type:    "PROGBITS",
		Address: 0x400238,
		Offset:  0x238,
		Size:    0x1c,
		EntSize: 0x0}
	if sects[0] != exp {
		t.Errorf("expected %v, got %v", exp, sects[0])
	}
}
