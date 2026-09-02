module github.com/tarcisiozf/wasp

go 1.25

require (
	github.com/bytecodealliance/wasmtime-go v1.0.0
	github.com/jairad26/go-simd v0.0.0-20260306190313-fc879d565e63
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/text v0.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/jairad26/go-simd => github.com/tarcisiozf/go-simd v0.0.0-20260306190313-fc879d565e63
