package objdump

import (
	"strings"
	"testing"
)

func TestParseDisassembly(t *testing.T) {
	inp := `
TEXT vendor/golang_org/x/net/http2/hpack.pair(SB) /home/todd/Projects/go/src/vendor/golang_org/x/net/http2/hpack/tables.go
        tables.go:128   0x5997a0        4883ec30                SUBQ $0x30, SP          
        tables.go:128   0x5997a4        48896c2428              MOVQ BP, 0x28(SP)       
        tables.go:128   0x5997a9        488d6c2428              LEAQ 0x28(SP), BP       
        tables.go:128   0x5997ae        48c744245800000000      MOVQ $0x0, 0x58(SP)     
        tables.go:128   0x5997b7        48c744246000000000      MOVQ $0x0, 0x60(SP)     
        tables.go:128   0x5997c0        48c744246800000000      MOVQ $0x0, 0x68(SP)     
        tables.go:128   0x5997c9        48c744247000000000      MOVQ $0x0, 0x70(SP)     
        tables.go:128   0x5997d2        48c744247800000000      MOVQ $0x0, 0x78(SP)     
        tables.go:129   0x5997db        48c7042400000000        MOVQ $0x0, 0(SP)        
        tables.go:129   0x5997e3        48c744240800000000      MOVQ $0x0, 0x8(SP)      
        tables.go:129   0x5997ec        48c744241000000000      MOVQ $0x0, 0x10(SP)     
        tables.go:129   0x5997f5        48c744241800000000      MOVQ $0x0, 0x18(SP)     
        tables.go:129   0x5997fe        48c744242000000000      MOVQ $0x0, 0x20(SP)     
        tables.go:129   0x599807        488b442438              MOVQ 0x38(SP), AX       
        tables.go:129   0x59980c        48890424                MOVQ AX, 0(SP)          
        tables.go:129   0x599810        488b442440              MOVQ 0x40(SP), AX       
        tables.go:129   0x599815        4889442408              MOVQ AX, 0x8(SP)        
        tables.go:129   0x59981a        488b442448              MOVQ 0x48(SP), AX       
        tables.go:129   0x59981f        4889442410              MOVQ AX, 0x10(SP)       
        tables.go:129   0x599824        488b442450              MOVQ 0x50(SP), AX       
        tables.go:129   0x599829        4889442418              MOVQ AX, 0x18(SP)       
        tables.go:129   0x59982e        488b0424                MOVQ 0(SP), AX          
        tables.go:129   0x599832        4889442458              MOVQ AX, 0x58(SP)       
        tables.go:128   0x599837        488d7c2460              LEAQ 0x60(SP), DI       
        tables.go:129   0x59983c        488d742408              LEAQ 0x8(SP), SI        
        tables.go:129   0x599841        48896c24f0              MOVQ BP, -0x10(SP)      
        tables.go:129   0x599846        488d6c24f0              LEAQ -0x10(SP), BP      
        tables.go:129   0x59984b        e884f9ebff              CALL 0x4591d4           
        tables.go:129   0x599850        488b6d00                MOVQ 0(BP), BP          
        tables.go:129   0x599854        488b6c2428              MOVQ 0x28(SP), BP       
        tables.go:129   0x599859        4883c430                ADDQ $0x30, SP          
        tables.go:129   0x59985d        c3                      RET      
`
	fns, err := parseDisassembly(strings.NewReader(inp))
	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}
	if len(fns) != 1 {
		t.Fatalf("expected one function, got %d", len(fns))
	}
	fn := fns[0]
	if exp := "vendor/golang_org/x/net/http2/hpack.pair"; fn.Name != exp {
		t.Errorf("expected %s, got %s", exp, fn.Name)
	}
	if exp := "/home/todd/Projects/go/src/vendor/golang_org/x/net/http2/hpack/tables.go"; fn.File != exp {
		t.Errorf("expected %s, got %s", exp, fn.File)
	}
	if exp := 32; exp != len(fn.Asm) {
		t.Errorf("expected %d instructions, got %d", exp, len(fn.Asm))
	}
	lastInsn := fn.Asm[len(fn.Asm)-1]
	if exp := (Disasm{File: "tables.go", Line: 129, Offset: 0x59985d,
		Bin: "c3", Asm: "RET"}); exp != lastInsn {
		t.Errorf("expected %v, got %v", exp, lastInsn)
	}
}
