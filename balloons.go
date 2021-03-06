package main

import (
	"fmt"
	"image/png"
	"os"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

const winWidth, winHeight int = 800, 600

type texture struct {
	pos
	pixels      []byte
	w, h, pitch int
}

type rgba struct {
	r, g, b byte
}

type pos struct {
	x, y float32
}

func clear(pixels []byte) {
	for i := range pixels {
		pixels[i] = 0
	}
}

func setPixel(x, y int, c rgba, pixels []byte) {
	index := (y*winWidth + x) * 4
	if index < len(pixels)-4 && index >= 0 {
		pixels[index] = c.r
		pixels[index+1] = c.g
		pixels[index+2] = c.b
	}
}

func (tex *texture) draw(pixels []byte) {
	for y := 0; y < tex.h; y++ {
		for x := 0; x < tex.w; x++ {
			screenY := y + int(tex.y)
			screenX := x + int(tex.x)
			if screenX >= 0 && screenX < winWidth && screenY >= 0 && screenY < winHeight {
				texIndex := y*tex.pitch + x*4
				screenIndex := screenY*winWidth*4 + screenX*4

				pixels[screenIndex] = tex.pixels[texIndex]
				pixels[screenIndex+1] = tex.pixels[texIndex+1]
				pixels[screenIndex+2] = tex.pixels[texIndex+2]
				pixels[screenIndex+3] = tex.pixels[texIndex+3]
			}
		}
	}
}

func (tex *texture) drawAlpha(pixels []byte) {
	for y := 0; y < tex.h; y++ {
		for x := 0; x < tex.w; x++ {
			screenY := y + int(tex.y)
			screenX := x + int(tex.x)
			if screenX >= 0 && screenX < winWidth && screenY >= 0 && screenY < winHeight {
				texIndex := y*tex.pitch + x*4
				screenIndex := screenY*winWidth*4 + screenX*4

				sourceR := int(tex.pixels[texIndex])
				sourceG := int(tex.pixels[texIndex+1])
				sourceB := int(tex.pixels[texIndex+2])
				sourceA := int(tex.pixels[texIndex+3])

				destinationR := int(pixels[screenIndex])
				destinationG := int(pixels[screenIndex+1])
				destinationB := int(pixels[screenIndex+2])

				resultR := (sourceR*255 + destinationR*(255-sourceA)) / 255
				resultG := (sourceG*255 + destinationG*(255-sourceA)) / 255
				resultB := (sourceB*255 + destinationB*(255-sourceA)) / 255

				pixels[screenIndex] = byte(resultR)
				pixels[screenIndex+1] = byte(resultG)
				pixels[screenIndex+2] = byte(resultB)
				// pixels[screenIndex+3] = tex.pixels[texIndex+3]
			}
		}
	}
}

func loadBalloons() []texture {

	balloonStrings := []string{"balloon_red.png", "balloon_green.png", "balloon_blue.png"}
	balloonTextures := make([]texture, len(balloonStrings))

	for i, bstr := range balloonStrings {
		infile, err := os.Open(bstr)
		if err != nil {
			panic(err)
		}
		defer infile.Close()

		img, err := png.Decode(infile)
		if err != nil {
			panic(err)
		}

		w := img.Bounds().Max.X
		h := img.Bounds().Max.Y

		balloonPixels := make([]byte, w*h*4)
		bIndex := 0
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				r, g, b, a := img.At(x, y).RGBA()
				balloonPixels[bIndex] = byte(r / 256)
				bIndex++
				balloonPixels[bIndex] = byte(g / 256)
				bIndex++
				balloonPixels[bIndex] = byte(b / 256)
				bIndex++
				balloonPixels[bIndex] = byte(a / 256)
				bIndex++
			}
		}
		balloonTextures[i] = texture{pos{0, 0}, balloonPixels, w, h, w * 4}
	}
	return balloonTextures
}

func main() {

	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Testing SDL2", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, int32(winWidth), int32(winHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer renderer.Destroy()

	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(winWidth), int32(winHeight))
	if err != nil {
		fmt.Println(err)
	}
	defer tex.Destroy()

	pixels := make([]byte, winWidth*winHeight*4)
	balloonTextures := loadBalloons()
	direction := 1

	for {
		frameStart := time.Now()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}

		clear(pixels)

		for _, tex := range balloonTextures {
			tex.drawAlpha(pixels)
		}

		balloonTextures[1].x += float32(1 * direction)
		if balloonTextures[1].x > 400 || balloonTextures[1].x < 0 {
			direction = direction * -1
		}

		tex.Update(nil, pixels, winWidth*4)
		renderer.Copy(tex, nil, nil)
		renderer.Present()
		elapsedTime := float32(time.Since(frameStart).Seconds() * 1000)
		fmt.Println("ms per frame:", elapsedTime)
		if elapsedTime < 5 {
			sdl.Delay(5 - uint32(elapsedTime))
			elapsedTime = float32(time.Since(frameStart).Seconds())
		}
		sdl.Delay(16)
	}
}
