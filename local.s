0000000: 0061 736d                                 ; WASM_BINARY_MAGIC
0000004: 0100 0000                                 ; WASM_BINARY_VERSION
; section "Type" (1)
0000008: 01                                        ; section code
0000009: 00                                        ; section size (guess)
000000a: 02                                        ; num types
; func type 0
000000b: 60                                        ; func
000000c: 01                                        ; num params
000000d: 7f                                        ; i32
000000e: 00                                        ; num results
; func type 1
000000f: 60                                        ; func
0000010: 00                                        ; num params
0000011: 00                                        ; num results
0000009: 08                                        ; FIXUP section size
; section "Import" (2)
0000012: 02                                        ; section code
0000013: 00                                        ; section size (guess)
0000014: 01                                        ; num imports
; import header 0
0000015: 07                                        ; string length
0000016: 636f 6e73 6f6c 65                        console  ; import module name
000001d: 03                                        ; string length
000001e: 6c6f 67                                  log  ; import field name
0000021: 00                                        ; import kind
0000022: 00                                        ; import signature index
0000013: 0f                                        ; FIXUP section size
; section "Function" (3)
0000023: 03                                        ; section code
0000024: 00                                        ; section size (guess)
0000025: 01                                        ; num functions
0000026: 01                                        ; function 0 signature index
0000024: 02                                        ; FIXUP section size
; section "Start" (8)
0000027: 08                                        ; section code
0000028: 00                                        ; section size (guess)
0000029: 01                                        ; start func index
0000028: 01                                        ; FIXUP section size
; section "Code" (10)
000002a: 0a                                        ; section code
000002b: 00                                        ; section size (guess)
000002c: 01                                        ; num functions
; function body 0
000002d: 00                                        ; func body size (guess)
000002e: 01                                        ; local decl count
000002f: 01                                        ; local type count
0000030: 7f                                        ; i32
0000031: 41                                        ; i32.const
0000032: 0a                                        ; i32 literal
0000033: 21                                        ; local.set
0000034: 00                                        ; local index
0000035: 20                                        ; local.get
0000036: 00                                        ; local index
0000037: 10                                        ; call
0000038: 00                                        ; function index
0000039: 0b                                        ; end
000002d: 0c                                        ; FIXUP func body size
000002b: 0e                                        ; FIXUP section size
