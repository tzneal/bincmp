# symSizeComp

Usage: symSizeComb bin1 bin2

I'm currently using this to determine code size differences when modifying rewrite
rules in go's dev.ssa branch.


	todd@tz-lab$ symSizeComp ~/Projects/go/bin/go ~/Projects/goclean/bin/go  | head
	# delta name sz1 sz2
	-16 fmt.getField 272 288
	-16 syscall.WaitStatus.ExitStatus 64 80
	-16 encoding/json.(*decodeState).indirect 1584 1600
	-16 math/big.(*Int).binaryGCD 800 816
	-16 sync.(*WaitGroup).Wait 272 288
	...
	-16 reflect.Value.Complex 224 240
	-16 runtime.mallocinit 1136 1152
	/home/todd/Projects/go/bin/go is smaller than /home/todd/Projects/goclean/bin/go [-4914]
