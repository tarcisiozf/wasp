0000000: 0061 736d                                 ; WASM_BINARY_MAGIC
0000004: 0100 0000                                 ; WASM_BINARY_VERSION
; section "Type" (1)
0000008: 01                                        ; section code
0000009: 00                                        ; section size (guess)
000000a: 02                                        ; num types
; func type 0
000000b: 60                                        ; func
000000c: 00                                        ; num params
000000d: 01                                        ; num results
000000e: 7f                                        ; i32
; func type 1
000000f: 60                                        ; func
0000010: 02                                        ; num params
0000011: 6f                                        ; externref
0000012: 7f                                        ; i32
0000013: 01                                        ; num results
0000014: 6f                                        ; externref
0000009: 0b                                        ; FIXUP section size
; section "Function" (3)
0000015: 03                                        ; section code
0000016: 00                                        ; section size (guess)
0000017: 02                                        ; num functions
0000018: 00                                        ; function 0 signature index
0000019: 01                                        ; function 1 signature index
0000016: 03                                        ; FIXUP section size
; section "Export" (7)
000001a: 07                                        ; section code
000001b: 00                                        ; section size (guess)
000001c: 02                                        ; num exports
000001d: 0d                                        ; string length
000001e: 7365 6c65 6374 5f73 696d 706c 65         select_simple  ; export name
000002b: 00                                        ; export kind
000002c: 00                                        ; export func index
000002d: 10                                        ; string length
000002e: 7365 6c65 6374 5f65 7874 6572 6e72 6566  select_externref  ; export name
000003e: 00                                        ; export kind
000003f: 01                                        ; export func index
000001b: 24                                        ; FIXUP section size
; section "Code" (10)
0000040: 0a                                        ; section code
0000041: 00                                        ; section size (guess)
0000042: 02                                        ; num functions
; function body 0
0000043: 00                                        ; func body size (guess)
0000044: 00                                        ; local decl count
0000045: 41                                        ; i32.const
0000046: 0a                                        ; i32 literal
0000047: 41                                        ; i32.const
0000048: 14                                        ; i32 literal
0000049: 41                                        ; i32.const
000004a: 00                                        ; i32 literal
000004b: 1b                                        ; select
000004c: 0b                                        ; end
0000043: 09                                        ; FIXUP func body size
; function body 1
000004d: 00                                        ; func body size (guess)
000004e: 00                                        ; local decl count
000004f: d0                                        ; ref.null
0000050: 6f                                        ; ref.null type
0000051: 20                                        ; local.get
0000052: 00                                        ; local index
0000053: 20                                        ; local.get
0000054: 01                                        ; local index
0000055: 1c                                        ; select
0000056: 01                                        ; num result types
0000057: 6f                                        ; result type
0000058: 0b                                        ; end
000004d: 0b                                        ; FIXUP func body size
0000041: 17                                        ; FIXUP section size
