Here's the full list of WebAssembly value types and their codes from the spec:

## Number Types
| Type | Code |
|------|------|
| i32 | `0x7F` |
| i64 | `0x7E` |
| f32 | `0x7D` |
| f64 | `0x7C` |

## Vector Types
| Type | Code |
|------|------|
| v128 | `0x7B` |

## Reference Types
| Type | Code |
|------|------|
| funcref | `0x70` |
| externref | `0x6F` |

## Block/Result Types
| Type | Code |
|------|------|
| void (empty result) | `0x40` |

Your current implementation covers `i32`, `i64`, and `void`. To be complete, you'd add:

```go
var (
    Int32     = addType(0x7F, &typeInt32{})
    Int64     = addType(0x7E, &typeInt64{})
    Float32   = addType(0x7D, &typeFloat32{})
    Float64   = addType(0x7C, &typeFloat64{})
    Vec128    = addType(0x7B, &typeVec128{})
    FuncRef   = addType(0x70, &typeFuncRef{})
    ExternRef = addType(0x6F, &typeExternRef{})
    Void      = addType(0x40, &typeVoid{})
)
```

**Note:** The codes use negative LEB128 encoding in the binary format, which is why they appear as high values like `0x7F` (-1 as signed byte) for i32.