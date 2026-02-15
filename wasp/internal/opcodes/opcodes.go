package opcodes

type Opcode = uint16

const (
	// Control flow instructions
	Unreachable        Opcode = 0x00
	Nop                Opcode = 0x01
	Block              Opcode = 0x02
	Loop               Opcode = 0x03
	If                 Opcode = 0x04
	Else               Opcode = 0x05
	End                Opcode = 0x0b
	Br                 Opcode = 0x0c
	BrIf               Opcode = 0x0d
	BrTable            Opcode = 0x0e
	Return             Opcode = 0x0f
	Call               Opcode = 0x10
	CallIndirect       Opcode = 0x11
	ReturnCall         Opcode = 0x12
	ReturnCallIndirect Opcode = 0x13
	Drop               Opcode = 0x1a
	Select             Opcode = 0x1b
	SelectT            Opcode = 0x1c

	// Variable instructions
	LocalGet  Opcode = 0x20
	LocalSet  Opcode = 0x21
	LocalTee  Opcode = 0x22
	GlobalGet Opcode = 0x23
	GlobalSet Opcode = 0x24

	// Memory load instructions
	I32Load    Opcode = 0x28
	I64Load    Opcode = 0x29
	F32Load    Opcode = 0x2a
	F64Load    Opcode = 0x2b
	I32Load8S  Opcode = 0x2c
	I32Load8U  Opcode = 0x2d
	I32Load16S Opcode = 0x2e
	I32Load16U Opcode = 0x2f
	I64Load8S  Opcode = 0x30
	I64Load8U  Opcode = 0x31
	I64Load16S Opcode = 0x32
	I64Load16U Opcode = 0x33
	I64Load32S Opcode = 0x34
	I64Load32U Opcode = 0x35

	// Memory store instructions
	I32Store   Opcode = 0x36
	I64Store   Opcode = 0x37
	F32Store   Opcode = 0x38
	F64Store   Opcode = 0x39
	I32Store8  Opcode = 0x3a
	I32Store16 Opcode = 0x3b
	I64Store8  Opcode = 0x3c
	I64Store16 Opcode = 0x3d
	I64Store32 Opcode = 0x3e

	// Memory instructions
	MemorySize Opcode = 0x3f
	MemoryGrow Opcode = 0x40

	// Const instructions
	I32Const Opcode = 0x41
	I64Const Opcode = 0x42
	F32Const Opcode = 0x43
	F64Const Opcode = 0x44

	// Comparison instructions
	I32Eqz Opcode = 0x45
	I32Eq  Opcode = 0x46
	I32Ne  Opcode = 0x47
	I32LtS Opcode = 0x48
	I32LtU Opcode = 0x49
	I32GtS Opcode = 0x4a
	I32GtU Opcode = 0x4b
	I32LeS Opcode = 0x4c
	I32LeU Opcode = 0x4d
	I32GeS Opcode = 0x4e
	I32GeU Opcode = 0x4f

	I64Eqz Opcode = 0x50
	I64Eq  Opcode = 0x51
	I64Ne  Opcode = 0x52
	I64LtS Opcode = 0x53
	I64LtU Opcode = 0x54
	I64GtS Opcode = 0x55
	I64GtU Opcode = 0x56
	I64LeS Opcode = 0x57
	I64LeU Opcode = 0x58
	I64GeS Opcode = 0x59
	I64GeU Opcode = 0x5a

	F32Eq Opcode = 0x5b
	F32Ne Opcode = 0x5c
	F32Lt Opcode = 0x5d
	F32Gt Opcode = 0x5e
	F32Le Opcode = 0x5f
	F32Ge Opcode = 0x60

	F64Eq Opcode = 0x61
	F64Ne Opcode = 0x62
	F64Lt Opcode = 0x63
	F64Gt Opcode = 0x64
	F64Le Opcode = 0x65
	F64Ge Opcode = 0x66

	// Integer arithmetic instructions
	I32Clz    Opcode = 0x67
	I32Ctz    Opcode = 0x68
	I32Popcnt Opcode = 0x69
	I32Add    Opcode = 0x6a
	I32Sub    Opcode = 0x6b
	I32Mul    Opcode = 0x6c
	I32DivS   Opcode = 0x6d
	I32DivU   Opcode = 0x6e
	I32RemS   Opcode = 0x6f
	I32RemU   Opcode = 0x70
	I32And    Opcode = 0x71
	I32Or     Opcode = 0x72
	I32Xor    Opcode = 0x73
	I32Shl    Opcode = 0x74
	I32ShrS   Opcode = 0x75
	I32ShrU   Opcode = 0x76
	I32Rotl   Opcode = 0x77
	I32Rotr   Opcode = 0x78

	I64Clz    Opcode = 0x79
	I64Ctz    Opcode = 0x7a
	I64Popcnt Opcode = 0x7b
	I64Add    Opcode = 0x7c
	I64Sub    Opcode = 0x7d
	I64Mul    Opcode = 0x7e
	I64DivS   Opcode = 0x7f
	I64DivU   Opcode = 0x80
	I64RemS   Opcode = 0x81
	I64RemU   Opcode = 0x82
	I64And    Opcode = 0x83
	I64Or     Opcode = 0x84
	I64Xor    Opcode = 0x85
	I64Shl    Opcode = 0x86
	I64ShrS   Opcode = 0x87
	I64ShrU   Opcode = 0x88
	I64Rotl   Opcode = 0x89
	I64Rotr   Opcode = 0x8a

	// Float instructions
	F32Abs      Opcode = 0x8b
	F32Neg      Opcode = 0x8c
	F32Ceil     Opcode = 0x8d
	F32Floor    Opcode = 0x8e
	F32Trunc    Opcode = 0x8f
	F32Nearest  Opcode = 0x90
	F32Sqrt     Opcode = 0x91
	F32Add      Opcode = 0x92
	F32Sub      Opcode = 0x93
	F32Mul      Opcode = 0x94
	F32Div      Opcode = 0x95
	F32Min      Opcode = 0x96
	F32Max      Opcode = 0x97
	F32Copysign Opcode = 0x98

	F64Abs      Opcode = 0x99
	F64Neg      Opcode = 0x9a
	F64Ceil     Opcode = 0x9b
	F64Floor    Opcode = 0x9c
	F64Trunc    Opcode = 0x9d
	F64Nearest  Opcode = 0x9e
	F64Sqrt     Opcode = 0x9f
	F64Add      Opcode = 0xa0
	F64Sub      Opcode = 0xa1
	F64Mul      Opcode = 0xa2
	F64Div      Opcode = 0xa3
	F64Min      Opcode = 0xa4
	F64Max      Opcode = 0xa5
	F64Copysign Opcode = 0xa6

	// Conversion instructions
	I32WrapI64     Opcode = 0xa7
	I32TruncF32S   Opcode = 0xa8
	I32TruncF32U   Opcode = 0xa9
	I32TruncF64S   Opcode = 0xaa
	I32TruncF64U   Opcode = 0xab
	I64ExtendI32S  Opcode = 0xac
	I64ExtendI32U  Opcode = 0xad
	I64TruncF32S   Opcode = 0xae
	I64TruncF32U   Opcode = 0xaf
	I64TruncF64S   Opcode = 0xb0
	I64TruncF64U   Opcode = 0xb1
	F32ConvertI32S Opcode = 0xb2
	F32ConvertI32U Opcode = 0xb3
	F32ConvertI64S Opcode = 0xb4
	F32ConvertI64U Opcode = 0xb5
	F32DemoteF64   Opcode = 0xb6
	F64ConvertI32S Opcode = 0xb7
	F64ConvertI32U Opcode = 0xb8
	F64ConvertI64S Opcode = 0xb9
	F64ConvertI64U Opcode = 0xba
	F64PromoteF32  Opcode = 0xbb

	// Reinterpret instructions
	I32ReinterpretF32 Opcode = 0xbc
	I64ReinterpretF64 Opcode = 0xbd
	F32ReinterpretI32 Opcode = 0xbe
	F64ReinterpretI64 Opcode = 0xbf

	I32Extend8S Opcode = 0xc0
)
