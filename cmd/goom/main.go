package main

import (
	"fmt"
	"image/color"
	"os"
	"sync"
	"time"
	"wasp/wasi"
	"wasp/wasp"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var gameCh = make(chan ebiten.Game, 1)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		println("Usage: run <wasm file>")
		os.Exit(1)
	}

	wasm, err := os.ReadFile(args[0])
	if err != nil {
		println("Error reading file:", err.Error())
		os.Exit(1)
	}

	start := time.Now()
	module, err := wasp.NewModule(wasm)
	if err != nil {
		println("Error loading module:", err.Error())
		os.Exit(1)
	}
	fmt.Println("WASM loaded in ", time.Since(start))

	store := wasp.NewStore(module)

	linker := wasp.NewLinker()

	dg(linker, store)

	sp := wasi.NewWasiSnapshotPreview1()
	sp.SetArgs([]string{"doom1.wad"}) // Pass remaining args to WASI
	sp.AddPreopen(3, ".")             // Preopen current directory as fd 3
	//sp.AddPreopen(4, "doom1.wad")
	sp.SetMemory(store.Memories[0]) // Set the memory for WASI to use
	if err := sp.Register(linker); err != nil {
		println("Error defining function:", err.Error())
		os.Exit(1)
	}

	options := []wasp.InstanceOption{
		wasp.WithLinker(linker),
		wasp.IgnoreUnreachable(), // Allow DOOM to continue despite UBSan panics
	}
	for _, arg := range args {
		switch arg {
		case "-v", "--verbose":
			options = append(options, wasp.Verbose())
		}
	}

	instance, err := wasp.NewInstance(
		module,
		store,
		options...,
	)
	if err != nil {
		println("Error creating runtime:", err.Error())
		os.Exit(1)
	}

	fn, err := module.GetExportedFunction("_start")
	if err != nil {
		println("Error getting function:", err.Error())
		os.Exit(1)
	}

	go (func() {
		_, err := instance.Call(fn)
		if err != nil {
			println("Error calling function:", err.Error())
			os.Exit(1)
		}
		if err := instance.Run(); err != nil {
			println("Error during execution:", err.Error())
			os.Exit(1)
		}
	})()

	if err := ebiten.RunGame(<-gameCh); err != nil {
		panic(err)
	}
}

type keyEvent struct {
	key     byte
	pressed bool
}

type Game struct {
	width  int
	height int

	img       *ebiten.Image
	pixels    []byte
	keyEvents chan keyEvent
}

var _ ebiten.Game = (*Game)(nil)

const (
	keyRightArrow = 0xae
	keyLeftArrow  = 0xac
	keyUpArrow    = 0xad
	keyDownArrow  = 0xaf
	keyStrafeL    = 0xa0
	keyStrafeR    = 0xa1
	keyUse        = 0xa2
	keyFire       = 0xa3
	keyEscape     = 27
	keyEnter      = 13
	keyTab        = 9
)

var keyMap = map[byte]ebiten.Key{
	keyRightArrow: ebiten.KeyRight,
	keyLeftArrow:  ebiten.KeyLeft,
	keyUpArrow:    ebiten.KeyUp,
	keyDownArrow:  ebiten.KeyDown,
	keyStrafeL:    ebiten.KeyA,
	keyStrafeR:    ebiten.KeyD,
	keyUse:        ebiten.KeyE,
	keyFire:       ebiten.KeySpace,
	keyEscape:     ebiten.KeyEscape,
	keyEnter:      ebiten.KeyEnter,
	keyTab:        ebiten.KeyTab,
}

func (g *Game) Update() error {
	for code, key := range keyMap {
		var pressed bool
		if inpututil.IsKeyJustPressed(key) {
			pressed = true
		} else if inpututil.IsKeyJustReleased(key) {
			pressed = false
		} else {
			continue
		}
		g.keyEvents <- keyEvent{code, pressed}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.img.WritePixels(g.pixels)
	screen.Fill(color.Black)
	screen.DrawImage(g.img, nil)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.width, g.height
}

func newGame(width, height int) *Game {
	return &Game{
		width:  width,
		height: height,

		pixels: make([]byte, width*height*4), // RGBA format
		img:    ebiten.NewImage(width, height),

		keyEvents: make(chan keyEvent, 32),
	}
}

func dg(linker *wasp.Linker, store *wasp.Store) {
	var started time.Time
	var init sync.Once
	var game *Game
	var width, height int

	defaultMemory := store.Memories[0]

	linker.Define("dg", "init", func(resx, resy int32) {
		init.Do(func() {
			started = time.Now()

			width = int(resx)
			height = int(resy)
			ebiten.SetWindowSize(width, height)
			ebiten.SetWindowTitle("GOOM")

			game = newGame(width, height)
			gameCh <- game
			close(gameCh)
		})
	})

	frameCount := 0
	startTime := time.Now()
	linker.Define("dg", "draw_frame", func(ptr int32) {
		if game == nil {
			return
		}

		numPixels := width * height
		size := numPixels * 4
		data := defaultMemory.Load(int(ptr), size)

		pixels := make([]byte, size)
		for i := 0; i < numPixels; i++ {
			offset := i * 4
			pixels[offset] = data[offset+2]   // R
			pixels[offset+1] = data[offset+1] // G
			pixels[offset+2] = data[offset]   // B
			pixels[offset+3] = 0xFF           // A
		}
		copy(game.pixels, pixels)

		frameCount++
		elapsedTime := time.Since(startTime).Seconds()
		if elapsedTime > 0 {
			fps := float64(frameCount) / elapsedTime
			fmt.Printf("Frame Count: %d, FPS: %.2f\n", frameCount, fps)
		}
	})

	linker.Define("dg", "sleep_ms", func(ms int32) {
		time.Sleep(time.Duration(ms) * time.Millisecond)
	})

	linker.Define("dg", "get_ticks_ms", func() int32 {
		elapsed := time.Since(started)
		ms := int32(elapsed.Milliseconds())
		return ms
	})

	linker.Define("dg", "get_key", func() int32 {
		if game == nil {
			return -1
		}

		var event keyEvent
		select {
		case event = <-game.keyEvents:
		default:
			return -1
		}

		key := int32(event.key)
		if event.pressed {
			key |= 1 << 9
		}

		return key
	})

	linker.Define("dg", "set_window_title", func(ptr, len int32) {
		fmt.Printf("DB set_window_title called with ptr=%d, len=%d\n", ptr, len)
	})

	// UBSan stubs - allow undefined behavior to continue without panicking
	linker.Define("env", "__ubsan_handle_shift_out_of_bounds", func(dataPtr, lhs, rhs int32) {
		// Log but don't panic - this allows DOOM to continue despite UB
		fmt.Printf("[UBSAN] shift out of bounds: lhs=%d, rhs=%d (continuing)\n", lhs, rhs)
	})
}
