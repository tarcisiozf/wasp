build-asm:
	go build -o bin/wasp-asm cmd/asm/main.go

build-run:
	go build -o bin/wasp cmd/run/main.go

build-goom:
	go build -o bin/goom cmd/goom/main.go

build: build-asm build-run build-goom