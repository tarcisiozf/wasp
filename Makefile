build-asm:
	go build -o bin/wasp-asm cmd/asm/main.go

build-run:
	go build -o bin/wasp cmd/run/main.go

build: build-asm build-run