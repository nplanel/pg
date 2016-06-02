package main

import (
	_ "encoding/binary"
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"io"
	"os"
)

//var resX = 528
//var resY = 440
var resX = 720

// 420 480 540 600 700 1000 TVL (TV Lines)

var resY = 600
var resBytesPerPixel = 2
var resLineSz = resX * resBytesPerPixel
var resPageSz = resLineSz * resY

var stepBytes = 4

var window *sdl.Window
var renderer *sdl.Renderer

var diskfile *os.File
var data = make([]byte, resPageSz*6)
var data2 = make([]byte, resX*resY*3)

var once = true
var SaveBMP = false

func pixelToRGBAMono(data []byte, idx *int, x, y int) (r, g, b, a uint8) {
	//	v := binary.LittleEndian.Uint16(data[*idx : *idx+2])
	*idx = (y * resX * resBytesPerPixel) + (x * resBytesPerPixel)
	v := uint16(data[*idx+1])
	if once {
		once = false
		fmt.Printf("%x %x %x %x %x\n", data[*idx], data[*idx+1], data[*idx+2], data[*idx+3], v)
	}
	if y < 2 && x == 0 {
		fmt.Printf("%x %x %x %x\n", data[*idx], data[*idx+1], v, *idx)
	}

	r = uint8(v)
	g = uint8(v)
	b = uint8(v)
	a = 0xff

	*idx += 2
	return
}

func pixelToRGBAYUV420p(yuv []byte, idx *int, cx, cy int) (r, g, b, a uint8) {
	//	v := binary.LittleEndian.Uint16(yuv[*idx : *idx+2])
	*idx = (cy * resX * resBytesPerPixel) + (cx * resBytesPerPixel)
	va := uint16(yuv[*idx+1])
	if once {
		once = false
		fmt.Printf("%x %x %x %x %x\n", yuv[*idx], yuv[*idx+1], yuv[*idx+2], yuv[*idx+3], va)
	}

	/*
		u := yuv[*idx+0]
		y1 := yuv[*idx+1]
		v := yuv[*idx+2]
		y2 := yuv[*idx+3]
		y2 = y2
	*/

	sizetotal := resX * resY
	y := yuv[(cy*resX)+cx]
	u := yuv[(cy/2)*(resX/2)+(cx/2)+sizetotal]
	v := yuv[(cy/2)*(resX/2)+(cx/2)+sizetotal+(sizetotal/4)]

	if cy < 2 && cx == 0 {
		fmt.Printf("%x %x %x %x   %d %d %d\n", yuv[*idx], yuv[*idx+1], va, *idx, y, u, v)
	}

	r = uint8(y)
	g = uint8(u)
	b = uint8(v)
	a = 0xff

	*idx += 2
	return
}

func pixelToRGBAYUV411(yuv []byte, idx *int, cx, cy int) (r, g, b, a uint8) {
	*idx = (cy * resX * resBytesPerPixel) + (cx * resBytesPerPixel)
	va := uint16(yuv[*idx+1])
	if once {
		once = false
		fmt.Printf("%x %x %x %x %x\n", yuv[*idx], yuv[*idx+1], yuv[*idx+2], yuv[*idx+3], va)
	}

	u := yuv[*idx+0]
	y1 := yuv[*idx+1]
	y2 := yuv[*idx+2]
	v := yuv[*idx+3]
	y3 := yuv[*idx+4]
	y4 := yuv[*idx+5]

	if cy < 2 && cx == 0 {
		fmt.Printf("%x %x %x %x   %d %d %d\n", yuv[*idx], yuv[*idx+1], va, *idx, y1, u, v)
	}

	r = uint8(y1)
	g = uint8(u)
	b = uint8(v)
	a = 0xff

	renderer.SetDrawColor(y1, u, v, a)
	renderer.DrawPoint(cx, cy)
	renderer.SetDrawColor(y2, u, v, a)
	renderer.DrawPoint(cx+1, cy)
	renderer.SetDrawColor(y3, u, v, a)
	renderer.DrawPoint(cx+2, cy)
	renderer.SetDrawColor(y4, u, v, a)
	renderer.DrawPoint(cx+3, cy)

	*idx += 2
	return
}

func pixelToRGBAYUV422(yuv []byte, idx *int, cx, cy int) (r, g, b, a uint8) {
	*idx = (cy * resX * resBytesPerPixel) + (cx * resBytesPerPixel)
	va := uint16(yuv[*idx+1])
	if once {
		once = false
		fmt.Printf("%x %x %x %x %x\n", yuv[*idx], yuv[*idx+1], yuv[*idx+2], yuv[*idx+3], va)
	}

	u := yuv[*idx+0]
	y1 := yuv[*idx+1]
	v := yuv[*idx+2]
	y2 := yuv[*idx+3]

	if cy < 2 && cx == 0 {
		fmt.Printf("%x %x %x %x   %d %d %d\n", yuv[*idx], yuv[*idx+1], va, *idx, y1, u, v)
	}

	r = uint8(y1)
	g = uint8(u)
	b = uint8(v)
	a = 0xff

	renderer.SetDrawColor(y1, u, v, a)
	renderer.DrawPoint(cx, cy)
	renderer.SetDrawColor(y2, u, v, a)
	renderer.DrawPoint(cx+1, cy)

	*idx += 2
	return
}

func clamp(v float64, min, max uint8) (r uint8) {
	if v < float64(min) {
		r = min
	}
	if v > float64(max) {
		r = max
	}
	return uint8(v)
}

func yuv2rgb(_yValue, _uValue, _vValue uint8) (r, g, b uint8) {
	yValue := float64(_yValue)
	vValue := float64(_vValue)
	uValue := float64(_uValue)

	rTmp := yValue + (1.370705 * (vValue - 128))
	gTmp := yValue - (0.698001 * (vValue - 128)) - (0.337633 * (uValue - 128))
	bTmp := yValue + (1.732446 * (uValue - 128))
	r = clamp(rTmp, 0, 255)
	g = clamp(gTmp, 0, 255)
	b = clamp(bTmp, 0, 255)
	return
}

func pixelToRGBAYUV420android(yuv []byte, idx *int, cx, cy int) (r, g, b, a uint8) {
	*idx = (cy * resX * resBytesPerPixel) + (cx * resBytesPerPixel)
	va := uint16(yuv[*idx+1])
	if once {
		once = false
		fmt.Printf("%x %x %x %x %x\n", yuv[*idx], yuv[*idx+1], yuv[*idx+2], yuv[*idx+3], va)
	}

	y1 := yuv[*idx]
	u := yuv[*idx+1]
	v := yuv[*idx+1]

	if cy < 2 && cx == 0 {
		fmt.Printf("%x %x %x %x   %d %d %d\n", yuv[*idx], yuv[*idx+1], va, *idx, y1, u, v)
	}
	r, g, b = yuv2rgb(y1, u, v)
	a = 0xff

	renderer.SetDrawColor(r, g, b, a)
	renderer.DrawPoint(cx, cy)

	*idx += 2
	return
}

func pixelToRGBAYUVcolorntsc(yuv []byte, idx *int, cx, cy int) (r, g, b, a uint8) {
	*idx = (cy * resX * resBytesPerPixel) + (cx * resBytesPerPixel)
	va := uint16(yuv[*idx+1])
	if once {
		once = false
		fmt.Printf("%x %x %x %x %x %x  %x\n", yuv[*idx], yuv[*idx+1], yuv[*idx+2], yuv[*idx+3], yuv[*idx+4], yuv[*idx+5], va)
	}

	u := yuv[*idx+0]
	y1 := yuv[*idx+1]
	v := yuv[*idx+2]
	y2 := yuv[*idx+3]

	if cy < 2 && cx == 0 {
		fmt.Printf("%x %x %x %x   %d %d %d\n", yuv[*idx], yuv[*idx+1], va, *idx, y1, u, v)
	}
	a = 0xff

	c := int(y1) - 16
	d := int(u) - 128
	e := int(v) - 128
	r = clamp(float64((298*c+409*e+128)>>8), 0, 255)
	g = clamp(float64((298*c-100*d-208*e+128)>>8), 0, 255)
	b = clamp(float64((298*c+516*d+128)>>8), 0, 255)
	renderer.SetDrawColor(r, b, g, a)
	renderer.DrawPoint(cx, cy)
	c = int(y2) - 16
	r = clamp(float64((298*c+409*e+128)>>8), 0, 255)
	g = clamp(float64((298*c-100*d-208*e+128)>>8), 0, 255)
	b = clamp(float64((298*c+516*d+128)>>8), 0, 255)
	renderer.SetDrawColor(r, b, g, a)
	renderer.DrawPoint(cx+1, cy)

	if cy < 2 && cx == 0 {
		fmt.Printf("%x %x %x %x   %d %d %d\n", yuv[*idx], yuv[*idx+1], va, *idx, y1, u, v)
	}
	a = 0xff

	*idx += 2
	return
}

func yuv2rgb_itur(_yValue, _uValue, _vValue uint8) (r, g, b uint8) {
	y := float64(_yValue)
	v := float64(_vValue)
	u := float64(_uValue)

	r = clamp(float64(y+1.402*(v-128)), 0, 255)
	g = clamp(float64(y-0.344*(u-128)-0.714*(v-128)), 0, 255)
	b = clamp(float64(y+1.772*(u-128)), 0, 255)
	return
}

func yuv2rgb_generic(_yValue, _uValue, _vValue uint8) (r, g, b uint8) {
	Y := float64(_yValue)
	Cb := float64(int(_uValue)-128) * 1.772 /* b-y signal +0.886 -0.886 */
	Cr := float64(int(_vValue)-128) * 1.402 /* r-y signal +0.701 -0.701 */

	/* http://discoverybiz.net/enu0/faq/faq_YUV_YCbCr_YPbPr.html */
	/* ITU601 */
	Kry := 0.299
	Kby := 0.114
	/* ITU709 */
	//	Kry = 0.2126
	//	Kby = 0.0722
	/* SMPTE 240M */
	//	Kry = 0.212
	//	Kby = 0.087

	Kgy := 1.0 - Kry - Kby

	/*
		Kru := -Kry
		Kgu := -Kgy
		Kbu := 1 - Kby

		Krv := 1 - Kry
		Kgv := -Kgy
		Kbv := -Kby
	*/

	r = clamp(Y+Cr, 0, 255)
	g = clamp(Y-(Kby/Kgy)*Cb-(Kry/Kgy)*Cr, 0, 255)
	b = clamp(Y+Cb, 0, 255)
	return
}

func pixelToRGBAYUV_ITUR(yuv []byte, idx *int, cx, cy int) (r, g, b, a uint8) {
	*idx = (cy * resX * resBytesPerPixel) + (cx * resBytesPerPixel)
	va := uint16(yuv[*idx+1])
	if once {
		once = false
		fmt.Printf("%x %x %x %x %x %x  %x\n", yuv[*idx], yuv[*idx+1], yuv[*idx+2], yuv[*idx+3], yuv[*idx+4], yuv[*idx+5], va)
	}

	u := yuv[*idx+0]
	y1 := yuv[*idx+1]
	v := yuv[*idx+2]
	y2 := yuv[*idx+3]

	if cy < 2 && cx == 0 {
		fmt.Printf("%x %x %x %x   %d %d %d\n", yuv[*idx], yuv[*idx+1], va, *idx, y1, u, v)
	}
	a = 0xff

	r, g, b = yuv2rgb_generic(y1, u, v)
	renderer.SetDrawColor(r, g, b, a)
	renderer.DrawPoint(cx, cy)
	r, g, b = yuv2rgb_generic(y2, u, v)
	renderer.SetDrawColor(r, g, b, a)
	renderer.DrawPoint(cx+1, cy)

	if cy < 2 && cx == 0 {
		fmt.Printf("%x %x %x %x   %d %d %d\n", yuv[*idx], yuv[*idx+1], va, *idx, y1, u, v)
	}
	a = 0xff

	*idx += 2
	return
}

func pixelToRGBAYUVnp(yuv []byte, idx *int, cx, cy int) (r, g, b, a uint8) {
	*idx = (cy * resX * resBytesPerPixel) + (cx * resBytesPerPixel)
	va := uint16(yuv[*idx+1])
	if once {
		once = false
		fmt.Printf("%x %x %x %x %x %x  %x\n", yuv[*idx], yuv[*idx+1], yuv[*idx+2], yuv[*idx+3], yuv[*idx+4], yuv[*idx+5], va)
	}

	u := yuv[*idx+0]
	y1 := yuv[*idx+1]
	y2 := yuv[*idx+2]
	v := yuv[*idx+3]
	y3 := yuv[*idx+4]
	y4 := yuv[*idx+5]

	if cy < 2 && cx == 0 {
		fmt.Printf("%x %x %x %x   %d %d %d\n", yuv[*idx], yuv[*idx+1], va, *idx, y1, u, v)
	}
	a = 0xff

	r, g, b = yuv2rgb(y1, u, v)
	renderer.SetDrawColor(r, g, b, a)
	renderer.DrawPoint(cx, cy)
	r, g, b = yuv2rgb(y2, u, v)
	renderer.SetDrawColor(r, g, b, a)
	renderer.DrawPoint(cx+1, cy)
	r, g, b = yuv2rgb(y3, u, v)
	renderer.SetDrawColor(r, g, b, a)
	renderer.DrawPoint(cx+2, cy)
	r, g, b = yuv2rgb(y4, u, v)
	renderer.SetDrawColor(r, g, b, a)
	renderer.DrawPoint(cx+3, cy)

	if cy < 2 && cx == 0 {
		fmt.Printf("%x %x %x %x   %d %d %d\n", yuv[*idx], yuv[*idx+1], va, *idx, y1, u, v)
	}
	a = 0xff

	*idx += 2
	return
}

func pixelToRGBA512(data []byte, idx *int) (r, g, b, a uint8) {
	//	v := binary.LittleEndian.Uint16(data[*idx : *idx+2])
	v := uint16(data[*idx+1])
	if once {
		once = false
		fmt.Printf("%x %x %x", data[*idx], data[*idx+1], v)
	}

	r = uint8((v & 0x1C0) >> 6)
	g = uint8((v & 0x038) >> 3)
	b = uint8(v & 0x007)
	a = 0xff

	*idx += 2
	return
}

func pixelToRGBA24b(data []byte, idx *int) (r, g, b, a uint8) {
	//	v := binary.LittleEndian.Uint16(data[*idx : *idx+2])
	v := uint16(data[*idx+1])
	if once {
		once = false
		fmt.Printf("%x %x %x", data[*idx], data[*idx+1], v)
	}

	r = uint8(data[*idx+1])
	g = uint8(data[*idx+3])
	b = uint8(data[*idx+5])
	a = 0xff

	*idx += stepBytes
	return
}

func pixelToRGBA(data []byte, x, y int) (r, g, b, a uint8) {
	idx := (resX * y) + (x * resBytesPerPixel)
	idx++
	r = uint8(data[idx+0])
	g = uint8(data[idx+2])
	b = uint8(data[idx+4])
	a = 0xff
	return
}

func displayat(offset int64) {
	fmt.Println("display at ", offset, resBytesPerPixel, stepBytes, resX)

	_, err := diskfile.Seek(offset, os.SEEK_SET)
	if err != nil {
		panic(err)
	}
	_, err = diskfile.Read(data)
	if err == io.EOF {
		return
	}
	if err != nil {
		panic(err)
	}

	renderer.Clear()
	for y := 0; y < resY; y++ {
		for x := 0; x < resX; x += 2 {
			n := 0
			/*r, g, b, a := */ pixelToRGBAYUV_ITUR(data, &n, x, y)

		}
	}
	renderer.Present()
	window.UpdateSurface()
}

func main() {
	fmt.Println("display Quantel Paintbox Image")

	sdl.Init(sdl.INIT_EVERYTHING)

	var err error
	window, err = sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		resX, resY, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}
	surface = surface

	renderer, err = sdl.CreateSoftwareRenderer(surface) //sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", sdl.GetError())
		os.Exit(2)
	}
	defer renderer.Destroy()

	diskfile, err = os.Open("disk.img")
	if err != nil {
		panic(err)
	}
	defer diskfile.Close()

	addr := int64(0x114BC80)
	//addr = int64(152881792)
	addr = int64(0x90f7740)
	addr = int64(152930112)
	addr = int64(165371712)
	addr = int64(165789312)
	addr = int64(85562496) // manoir
	// end of line 90EAFFF

	addr = int64(158653616)

	displayat(addr)
	running := true
	for running {
		event := sdl.WaitEvent()
		switch t := event.(type) {
		case *sdl.KeyDownEvent:
			fmt.Printf("[%d ms] Keyboard\ttype:%d\tsym:%c %x\tmodifiers:%d\tstate:%d\trepeat:%d\n",
				t.Timestamp, t.Type, t.Keysym.Sym, t.Keysym.Sym, t.Keysym.Mod, t.State, t.Repeat)
			switch t.Keysym.Sym {
			case 'q':
				running = false
			case 'o':
				addr = addr - int64(resPageSz)
			case 'l':
				addr = addr + int64(resPageSz)
			case 'p':
				addr -= int64(resLineSz * 10)
			case 'm':
				addr += int64(resLineSz * 10)
			case 'u':
				resBytesPerPixel--
			case 'i':
				resBytesPerPixel++
			case 'k':
				stepBytes--
			case 'j':
				stepBytes++

			case 'e':
				resX -= 10
			case 'r':
				resX += 10
			case 'd':
				resX--
			case 'f':
				resX++

			case 's':
				surface.SaveBMP(fmt.Sprintf("imgQuantel-%08x.bmp", addr))

			case sdl.K_UP:
				addr -= int64(resLineSz)
			case sdl.K_DOWN:
				addr += int64(resLineSz)
			case sdl.K_RIGHT:
				addr -= int64(resBytesPerPixel * 2)
			case sdl.K_LEFT:
				addr += int64(resBytesPerPixel * 2)

			case 'b':
				addr -= int64(resBytesPerPixel * 20)
			case 'n':
				addr += int64(resBytesPerPixel * 20)
			}

			if stepBytes > 8 {
				stepBytes = 8
			}
			if stepBytes < 1 {
				stepBytes = 1
			}
			if resBytesPerPixel > 8 {
				resBytesPerPixel = 8
			}
			if resBytesPerPixel < 1 {
				resBytesPerPixel = 1
			}

			if addr < 0 {
				addr = 0
			}

			resLineSz = resX * resBytesPerPixel
			resPageSz = resX * resY * resBytesPerPixel

			displayat(addr)
		case *sdl.QuitEvent:
			running = false
		}

	}

	sdl.Quit()
}
