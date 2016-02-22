# symSizeComp

## Installation

	go get -u github.com/tzneal/symSizeComp

## Usage

	$ symSizeComp 
	Usage of symSizeComp:
	symSizeComp [options] newBinary oldBinary
	  -difference
		sort output by the symbol size difference (default true)
	  -disassemble
		display disassembly of non-matching functions
	  -larger
		only display larger symbols
	  -pattern string
		regular expression to match against symbols
	  -relative
		sort output by the relative symbol size difference
	  -size
		sort output by the new symbol size

## Example Output
### Normal
	$ symSizeComp ~/Projects/go/bin/go ~/Projects/goclean/bin/go
	# symbol differences
	-4415 runtime.pclntab 1474439 1478854 -0.298542%
	-192 type..eq.net/url.URL 784 976 -19.672131%
	-176 type..eq.net.netFD 464 640 -27.500000%
	...
	-16 type..eq.net.nssCriterion 272 288 -5.555556%
	-16 type..eq.net/http.connectMethod 272 288 -5.555556%

	# section differences
	global text (code) = -15904 bytes (-0.365848%)
	read-only data = -4511 bytes (-0.284264%)
	Total difference 20415 bytes (-0.326807%)

# With Disassembly
	$ symSizeComp --size --disassemble --pattern="type.*eq.*11.*float32" ~/Projects/go/bin/go ~/Projects/goclean/bin/go
	# symbol differences
	-16 type..eq.[11]float32 80 96 -16.666667%
	  571950:    mov    0x8(%rsp),%rdi                      5734b0:    mov    0x8(%rsp),%rdi
	  571955:    mov    0x10(%rsp),%rsi                     5734b5:    mov    0x10(%rsp),%rsi
	  57195a:    xor    %eax,%eax                           5734ba:    xor    %eax,%eax
	  57195c:    mov    $0xb,%rdx                           5734bc:    mov    $0xb,%rdx
	  571963:    cmp    %rdx,%rax                           5734c3:    cmp    %rdx,%rax
	  571966:    jge    571987 <type..eq.[11]float32+0x37>  5734c6:    jge    5734f3 <type..eq.[11]float32+0x43>
	  571968:    lea    (%rdi,%rax,4),%rbx                  5734c8:    cmp    $0x0,%rdi
	  57196c:    movss  (%rbx),%xmm0                        5734cc:    je     573503 <type..eq.[11]float32+0x53>
	  571970:    lea    (%rsi,%rax,4),%rbx                  5734ce:    lea    (%rdi,%rax,4),%rbx
	  571974:    movss  (%rbx),%xmm1                        5734d2:    movss  (%rbx),%xmm0
	  571978:    ucomiss %xmm0,%xmm1                        5734d6:    cmp    $0x0,%rsi
	  57197b:    jne    57198d <type..eq.[11]float32+0x3d>  5734da:    je     5734ff <type..eq.[11]float32+0x4f>
	  57197d:    jp     57198d <type..eq.[11]float32+0x3d>  5734dc:    lea    (%rsi,%rax,4),%rbx
	  57197f:    inc    %rax                                5734e0:    movss  (%rbx),%xmm1
	  571982:    cmp    %rdx,%rax                           5734e4:    ucomiss %xmm0,%xmm1
	  571985:    jl     571968 <type..eq.[11]float32+0x18>  5734e7:    jne    5734f9 <type..eq.[11]float32+0x49>
	  571987:    movb   $0x1,0x18(%rsp)                     5734e9:    jp     5734f9 <type..eq.[11]float32+0x49>
	  57198c:    retq                                       5734eb:    inc    %rax
	  57198d:    movb   $0x0,0x18(%rsp)                     5734ee:    cmp    %rdx,%rax
	  571992:    retq                                       5734f1:    jl     5734c8 <type..eq.[11]float32+0x18>
	  571993:    int3                                       5734f3:    movb   $0x1,0x18(%rsp)
	  571994:    int3                                       5734f8:    retq   
	  571995:    int3                                       5734f9:    movb   $0x0,0x18(%rsp)
	  571996:    int3                                       5734fe:    retq   
	  571997:    int3                                       5734ff:    mov    %eax,(%rsi)
	  571998:    int3                                       573501:    jmp    5734dc <type..eq.[11]float32+0x2c>
	  571999:    int3                                       573503:    mov    %eax,(%rdi)
	  57199a:    int3                                       573505:    jmp    5734ce <type..eq.[11]float32+0x1e>
	  57199b:    int3                                       573507:    int3   
	  57199c:    int3                                       573508:    int3   
	  57199d:    int3                                       573509:    int3   
	  57199e:    int3                                       57350a:    int3   
	  57199f:    int3                                       57350b:    int3   
								57350c:    int3   


