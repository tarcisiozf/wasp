build-disassembler:
	go build -o bin/wasp-asm cmd/disassembler/main.go

build-wasp:
	go build -o bin/wasp cmd/wasp/main.go

build: build-disassembler build-wasp

test:
	go test -v ./...