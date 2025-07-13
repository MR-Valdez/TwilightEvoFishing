package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"log"
	"runtime"
	"time"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/gltext"
	"github.com/go-vgo/robotgo"
	"github.com/kbinani/screenshot"
	hook "github.com/robotn/gohook"
	"golang.org/x/image/font/gofont/gobold"
)

type Tier struct {
	WaitSec int
}

type Resolution struct {
	Width, Height  int
	MoveCheck      [2]int         //Hero Move action location
	HeroCheck      [2]int         //Check if hero is alive - Attack picture
	YellowCheck    [2]int         //Yellow coords
	GreenCheck     [2]int         //Green coors
	InventorySlots [][2]int       //locations of items
	DropPoint      [2]int         //Where items are dropped - changed to hero pic
	FishingSpots   map[int][2]int // key = tier, value = {x,y} coords for that tier's fishing spot
}

var (
	enabled      = false
	currentTier  = 0
	lastTier     = 0
	statusText   = "PAUSED"
	fontRenderer *gltext.Font
	fontSize     float32
	resolution   Resolution
)

var resolutions = map[string]Resolution{
	"3840x2160": {
		Width: 3840, Height: 2160,
		MoveCheck:   [2]int{2790, 1717},
		HeroCheck:   [2]int{1670, 1900},
		YellowCheck: [2]int{921, 1254},
		GreenCheck:  [2]int{921, 1254},
		InventorySlots: [][2]int{
			{2390, 1800},
			{2390, 1940},
			{2535, 1940},
		},
		DropPoint: [2]int{1396, 1896},
		FishingSpots: map[int][2]int{
			1: {920, 2112},
			2: {797, 1972},
			3: {818, 1933},
			4: {700, 1762},
			5: {589, 2091},
		},
	},
	"2560x1440": {
		Width: 2560, Height: 1440,
		MoveCheck:   [2]int{1856, 1145},
		HeroCheck:   [2]int{1114, 1266},
		YellowCheck: [2]int{609, 834},
		GreenCheck:  [2]int{610, 836},
		InventorySlots: [][2]int{
			{1597, 1200},
			{1597, 1293},
			{1691, 1293},
		},
		DropPoint: [2]int{927, 1268},
		FishingSpots: map[int][2]int{
			1: {603, 1408},
			2: {531, 1314},
			3: {545, 1288},
			4: {466, 1174},
			5: {392, 1394},
		},
	},
	"1920x1080": {
		Width: 1920, Height: 1080,
		MoveCheck:   [2]int{1392, 860},
		HeroCheck:   [2]int{835, 951},
		YellowCheck: [2]int{453, 626},
		GreenCheck:  [2]int{454, 626},
		InventorySlots: [][2]int{
			{1197, 899},
			{1197, 967},
			{1270, 967},
		},
		DropPoint: [2]int{695, 950},
		FishingSpots: map[int][2]int{
			1: {460, 1056},
			2: {396, 986},
			3: {408, 968},
			4: {350, 881},
			5: {294, 1046},
		},
	},
}

func initGL() {
	runtime.LockOSThread()
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCompatProfile)
	glfw.WindowHint(glfw.TransparentFramebuffer, glfw.True)
	glfw.WindowHint(glfw.Decorated, glfw.False)
	glfw.WindowHint(glfw.Floating, glfw.True)
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.AlphaBits, 8)
}

func createOverlayWindow() *glfw.Window {
	screenBounds := screenshot.GetDisplayBounds(0)
	screenWidth := screenBounds.Dx()
	screenHeight := screenBounds.Dy()

	winWidth := int(float64(screenWidth) * 0.091)
	winHeight := int(float64(screenHeight) * 0.06)

	window, err := glfw.CreateWindow(winWidth, winHeight, "FishingBot Overlay", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	window.SetPos(int(float64(screenWidth)*0.05), int(float64(screenHeight)*0.05))

	return window
}

func setupFont(screenHeight int) {
	if err := gl.Init(); err != nil {
		panic(err)
	}

	fontData := bytes.NewReader(gobold.TTF)
	fontSize = float32(float64(screenHeight) * 0.01)
	font, err := gltext.LoadTruetype(fontData, int32(fontSize), 32, 126, gltext.LeftToRight)
	if err != nil {
		panic("Failed to load font: " + err.Error())
	}
	fontRenderer = font
}

func renderOverlayLoop(window *glfw.Window) {
	prevBounds := screenshot.GetDisplayBounds(0)
	prevWidth := prevBounds.Dx()
	prevHeight := prevBounds.Dy()

	for !window.ShouldClose() {
		screenBounds := screenshot.GetDisplayBounds(0)
		screenWidth := screenBounds.Dx()
		screenHeight := screenBounds.Dy()

		if screenWidth != prevWidth || screenHeight != prevHeight {
			winWidth := int(float64(screenWidth) * 0.091)
			winHeight := int(float64(screenHeight) * 0.06)
			window.SetSize(winWidth, winHeight)
			window.SetPos(int(float64(screenWidth)*0.05), int(float64(screenHeight)*0.05))
			setupFont(screenHeight)
			prevWidth, prevHeight = screenWidth, screenHeight
		}

		width, height := window.GetSize()

		gl.Viewport(0, 0, int32(width), int32(height))
		gl.ClearColor(0.0, 0.0, 0.0, 0.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		gl.Disable(gl.DEPTH_TEST)

		gl.MatrixMode(gl.PROJECTION)
		gl.LoadIdentity()
		gl.Ortho(0, float64(width), 0, float64(height), -1, 1)
		gl.MatrixMode(gl.MODELVIEW)
		gl.LoadIdentity()

		gl.Color3f(1, 1, 1)
		lineHeight := fontSize * 1.3
		y := float32(height) - lineHeight
		fontRenderer.Printf(10, y, "Status: %s", statusText)
		y -= lineHeight
		fontRenderer.Printf(10, y, "Tier: %d", currentTier)
		y -= lineHeight
		fontRenderer.Printf(10, y, "CTRL+F7 = Change Tier")
		y -= lineHeight
		fontRenderer.Printf(10, y, "CTRL+F8 = Toggle On/Off")

		window.SwapBuffers()
		glfw.PollEvents()
		time.Sleep(100 * time.Millisecond)
	}
}

func runBotLogic() {
	tiers := map[int]Tier{
		1: {WaitSec: 17},
		2: {WaitSec: 40},
		3: {WaitSec: 24},
		4: {WaitSec: 20},
		5: {WaitSec: 50},
	}

	robotgo.SetDelay(200)
	deaths := 0

	fmt.Println("🎣 Fishing App Started... Focus your game window before pressing F8")
	evChan := hook.Start()
	defer hook.End()

	go func() {
		for ev := range evChan {
			switch ev.Kind {
			case hook.KeyDown:
				if ev.Rawcode == 119 && (ev.Mask == 2 || ev.Mask == 32) { // Ctrl+F8
					enabled = !enabled
					if enabled {
						statusText = "ENABLED"
						fmt.Println("✅ Bot ENABLED")
					} else {
						statusText = "PAUSING"
						fmt.Println("⏸️ Bot PAUSED")
					}
				} else if ev.Rawcode == 118 && (ev.Mask == 2 || ev.Mask == 32) { // Ctrl+F7
					currentTier++
					if currentTier > 5 {
						currentTier = 1
					}
					fmt.Printf("🎣 Switched to Tier %d\n", currentTier)
				}

			case hook.MouseDown:
				if ev.Button == hook.MouseMap["center"] {
					x, y := robotgo.Location()
					fmt.Printf("🖱️ center-click at: x=%d, y=%d\n", x, y)
				}
			}
		}
	}()

	for {
		if !enabled {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if currentTier == 0 {
			fmt.Println("⚠️ No tier selected. Please press Ctrl+F7 to choose a tier.")
			time.Sleep(500 * time.Millisecond)
			continue
		}

		screenBounds := screenshot.GetDisplayBounds(0)
		screenWidth := screenBounds.Dx()
		screenHeight := screenBounds.Dy()

		key := fmt.Sprintf("%dx%d", screenWidth, screenHeight)
		res, ok := resolutions[key]
		if !ok {
			fmt.Printf("⚠️ Unsupported resolution: %s\n", key)
			enabled = false
			statusText = "PAUSED"
			continue
		}
		resolution = res

		if currentTier != lastTier {
			fmt.Printf("⬆️ Tier changed to %d — dropping fish...\n", currentTier)
			if lastTier != 0 {
				deaths = dropAllFish(4, screenWidth, screenHeight)
			}
			lastTier = currentTier
		}

		fmt.Println("Selecting hero...")
		robotgo.KeyTap("f1")
		time.Sleep(500 * time.Millisecond)

		heroCheckx := resolution.HeroCheck[0]
		heroChecky := resolution.HeroCheck[1]
		if isBlack(heroCheckx, heroChecky, screenBounds) {
			fmt.Println("⚠️ Hero is dead or not selected. Retrying...")
			if currentTier == 5 && enabled {
				enabled = false
				statusText = "PAUSED"
				fmt.Println("⏸️ Bot PAUSED for Tier 5 Death")
			}
			deaths++
			time.Sleep(15 * time.Second)
			continue
		}

		if !enabled {
			statusText = "PAUSED"
			fmt.Println("Fishing was paused - F8 to re-enable")
			continue
		}

		fmt.Println("Hero alive. Moving to fishing spot...")
		deaths = dropAllFish(deaths, screenWidth, screenHeight)

		// Move to fishing spot
		fishX, fishY := resolution.FishingSpots[currentTier][0], resolution.FishingSpots[currentTier][1]
		robotgo.Move(fishX, fishY, 0)
		robotgo.Click("right")

		if !waitInterruptible(time.Duration(tiers[currentTier].WaitSec)*time.Second, 250*time.Millisecond) {
			fmt.Println("⏹️ Wait interrupted due to pause")
			statusText = "PAUSED"
			continue
		}

		fmt.Println("Casting fishing skill...")
		robotgo.KeyTap("num8")

		lastRedDetected := time.Now()
		for {
			img, err := screenshot.CaptureRect(screenBounds)
			if err != nil {
				log.Printf("Screen capture failed: %v", err)
				continue
			}

			r1, g1, b1 := getRGB(img.At(resolution.YellowCheck[0], resolution.YellowCheck[1]))
			r2, g2, b2 := getRGB(img.At(resolution.GreenCheck[0], resolution.GreenCheck[1]))
			rm, gm, bm := getRGB(img.At(resolution.MoveCheck[0], resolution.MoveCheck[1]))

			if isYellowish(r1, g1, b1) {
				fmt.Println("🎯 Yellow detected → pressing UP")
				robotgo.KeyTap("up")
			} else if isGreenish(r2, g2, b2) {
				fmt.Println("🎯 Green detected → pressing DOWN")
				robotgo.KeyTap("down")
			} else if isRedish(rm, gm, bm) {
				if !enabled {
					statusText = "PAUSED"
					fmt.Println("Fishing was paused - F8 to re-enable")
					break
				}

				if time.Since(lastRedDetected) >= 5*time.Second {
					lastRedDetected = time.Now()
					fmt.Println("🐟 Red detected → Recasting...")
					robotgo.KeyTap("num8")
				}
			} else if isBlack(heroCheckx, heroChecky, screenBounds) {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func main() {
	initGL()
	window := createOverlayWindow()
	screenBounds := screenshot.GetDisplayBounds(0)
	screenHeight := screenBounds.Dy()
	setupFont(screenHeight)
	go runBotLogic()
	renderOverlayLoop(window)
}

func dropAllFish(deaths, screenWidth, screenHeight int) int {
	key := fmt.Sprintf("%dx%d", screenWidth, screenHeight)
	res, ok := resolutions[key]
	if !ok {
		fmt.Printf("Unsupported resolution: %s\n", key)
		return deaths
	}
	resolution = res

	if deaths >= 4 {
		fmt.Println("Dropping Fish")
		robotgo.KeyTap("f1")
		robotgo.KeyTap("f1")

		for _, slot := range resolution.InventorySlots {
			x, y := slot[0], slot[1]
			robotgo.Move(x, y)
			robotgo.Click("right")

			dx, dy := resolution.DropPoint[0], resolution.DropPoint[1]
			robotgo.Move(dx, dy)
			robotgo.Click()
		}

		time.Sleep(1 * time.Second)
		deaths = 0
	}
	return deaths
}

func getRGB(c color.Color) (uint32, uint32, uint32) {
	r, g, b, _ := color.RGBAModel.Convert(c).RGBA()
	return r >> 8, g >> 8, b >> 8
}

func isBlack(x, y int, bounds image.Rectangle) bool {
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		log.Printf("Screen capture failed: %v", err)
		return false
	}
	r, g, b := getRGB(img.At(x, y))
	return r == 0 && g == 0 && b == 0
}

func isYellowish(r, g, b uint32) bool {
	return r > 180 && g > 180 && b < 100
}

func isGreenish(r, g, b uint32) bool {
	return g > 180 && r < 150 && b < 150
}

func isRedish(r, g, b uint32) bool {
	return r > 180 && g < 120 && b < 120
}

func waitInterruptible(d time.Duration, checkInterval time.Duration) bool {
	elapsed := time.Duration(0)
	for elapsed < d {
		if !enabled {
			return false
		}
		time.Sleep(checkInterval)
		elapsed += checkInterval
	}
	return true
}
