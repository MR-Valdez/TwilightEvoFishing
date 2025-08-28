package main

import (
	"bytes"
	"fmt"
	"log"
	"runtime"
	"slices"
	"strconv"
	"syscall"
	"time"
	"unsafe"

	"TwilightEvoApp/utils"

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
	MoveCheck      [2]int             //Hero Move action location
	HeroCheck      [2]int             //Check if hero is alive - Attack picture
	YellowCheck    [2]int             //Yellow coords
	GreenCheck     [2]int             //Green coors
	InventorySlots [][2]int           //locations of items
	DropPoint      [2]int             //Where items are dropped - changed to hero pic
	FishingSpots   map[int][2]int     // key = tier, value = {x,y} coords for that tier's fishing spot
	DodgeArrows    map[string][][]int //Coordinates for dodging arrows
}

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	setWindowLongPtr     = user32.NewProc("SetWindowLongPtrW")
	getWindowLongPtr     = user32.NewProc("GetWindowLongPtrW")
	setLayeredWindowAttr = user32.NewProc("SetLayeredWindowAttributes")

	constGWL_EXSTYLE       = -20
	constWS_EX_LAYERED     = 0x80000
	constWS_EX_TRANSPARENT = 0x20
	constLWA_ALPHA         = 0x2

	windowFocused    = false
	fishing          = false
	dodgeArrows      = false
	manaRefresh      = false //Item will be on num8
	pauseManaRefresh = false
	typing           = false //For those robotgo operations that would put stuff or impact things while typing
	// ToDo I don't know if it impacts arrow dodging or not need to confirm
	manaRefreshCurrent  = 0
	manaRefreshInterval = []int{12500, 15500}
	currentTier         = 0
	lastTier            = 0
	statusText          = "PAUSED"
	dodgeText           = "PAUSED"
	manaStatusText      = "PAUSED"
	fontRenderer        *gltext.Font
	fontSize            float32
	resolution          Resolution
	actions             = []uint16{16, 17, 18, 19, 30}
	lastActionTime      = time.Now()
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
		DodgeArrows: map[string][][]int{
			"red":   {{676, 1267}},
			"up":    {{957, 1267}, {1019, 1267}},
			"down":  {{957, 1267}, {990, 1267}},
			"left":  {{947, 1267}, {974, 1267}},
			"right": {{929, 1267}, {963, 1267}},
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
		DodgeArrows: map[string][][]int{
			"red":   {{450, 844}},
			"up":    {{1, 844}, {1, 844}},
			"down":  {{633, 844}, {671, 844}},
			"left":  {{609, 844}, {645, 844}},
			"right": {{615, 844}, {638, 844}},
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
		DodgeArrows: map[string][][]int{
			"red":   {{1, 2}, {}},
			"up":    {{1, 2}, {}},
			"down":  {{1, 2}, {}},
			"left":  {{1, 2}, {}},
			"right": {{1, 2}, {}},
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
	glfw.WindowHint(glfw.Focused, glfw.False)
	glfw.WindowHint(glfw.Floating, glfw.True)
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.AlphaBits, 8)
}

func createOverlayWindow() *glfw.Window {
	screenBounds := screenshot.GetDisplayBounds(0)
	screenWidth := screenBounds.Dx()
	screenHeight := screenBounds.Dy()

	winWidth := int(float64(screenWidth) * 0.095)
	winHeight := int(float64(screenHeight) * 0.06)

	window, err := glfw.CreateWindow(winWidth, winHeight, "FishingBot Overlay", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	window.SetPos(int(float64(screenWidth)*0.05), int(float64(screenHeight)*0.05))

	return window
}

func makeWindowClickThrough(win *glfw.Window) {
	hwnd := win.GetWin32Window()
	hwndPtr := uintptr(unsafe.Pointer(hwnd))

	style, _, _ := getWindowLongPtr.Call(hwndPtr, uintptr(constGWL_EXSTYLE))
	style |= uintptr(constWS_EX_LAYERED | constWS_EX_TRANSPARENT)

	setWindowLongPtr.Call(hwndPtr, uintptr(constGWL_EXSTYLE), style)

	// Optional: set window alpha (255 = fully opaque, 0 = invisible)
	setLayeredWindowAttr.Call(hwndPtr, 0, 255, uintptr(constLWA_ALPHA))
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
		windowFocused = robotgo.GetTitle() == "Warcraft III"
		screenBounds := screenshot.GetDisplayBounds(0)
		screenWidth := screenBounds.Dx()
		screenHeight := screenBounds.Dy()

		if screenWidth != prevWidth || screenHeight != prevHeight {
			winWidth := int(float64(screenWidth) * 0.095)
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
		fontRenderer.Printf(10, y, "Mana Item(ALT+F6): %s", manaStatusText)
		y -= lineHeight
		fontRenderer.Printf(10, y, "Dodge Arrows(CTRL+F6): %s", dodgeText)
		y -= lineHeight
		fontRenderer.Printf(10, y, "Fishing(CTRL+F8): %s", statusText)
		y -= lineHeight
		fontRenderer.Printf(10, y, "Fishing Tier(CTRL+F7): %d", currentTier)
		/* y -= lineHeight
		fontRenderer.Printf(10, y, "CTRL+F7 = Change Fishing Tier")
		y -= lineHeight
		fontRenderer.Printf(10, y, "CTRL+F8 = Toggle Fishing On/Off") */

		window.SwapBuffers()
		glfw.PollEvents()
		time.Sleep(100 * time.Millisecond)
	}
}

func hookLogic() {
	evChan := hook.Start()

	go func() {
		defer hook.End()
		for ev := range evChan {
			if windowFocused {
				switch ev.Kind {
				case hook.KeyDown:
					if ev.Rawcode == 119 && (ev.Mask == 2 || ev.Mask == 32) { // Ctrl+F8
						fishing = !fishing
						if fishing {
							statusText = "ENABLED"
							fmt.Println("✅ Bot ENABLED")
						} else {
							statusText = "PAUSING"
							fmt.Println("⏸️  Bot PAUSED")
						}
					} else if ev.Rawcode == 118 && (ev.Mask == 2 || ev.Mask == 32) { // Ctrl+F7
						currentTier++
						if currentTier > 5 {
							currentTier = 1
						}
						fmt.Printf("🎣 Switched to Tier %d\n", currentTier)
					} else if ev.Rawcode == 117 && (ev.Mask == 2 || ev.Mask == 32) { // Ctrl+F6
						dodgeArrows = !dodgeArrows
						if dodgeArrows {
							dodgeText = "ENABLED"
							fmt.Println("⬆️  Dodging ENABLED")
						} else {
							dodgeText = "PAUSED"
							fmt.Println("⬆️  Dodging PAUSED")
						}
					} else if ev.Rawcode == 117 && (ev.Mask == 8 || ev.Mask == 128) { //Alt+F6
						if !manaRefresh {
							lastActionTime = time.Now()
							manaStatusText = "ENABLED 12"
							manaRefresh = !manaRefresh
							manaRefreshCurrent = manaRefreshInterval[0]
							fmt.Println("✅ Mana ENABLED 12")
						} else if manaRefresh && manaRefreshCurrent == manaRefreshInterval[0] {
							manaStatusText = "ENABLED 15"
							manaRefreshCurrent = manaRefreshInterval[1]
							fmt.Println("✅ Mana ENABLED 15")
						} else if manaRefresh && manaRefreshCurrent == manaRefreshInterval[1] {
							manaStatusText = "PAUSING"
							manaRefresh = !manaRefresh
							manaRefreshCurrent = 0
							fmt.Println("⏸️  Mana PAUSED")
						}
					} else if slices.Contains(actions, ev.Keycode) {
						lastActionTime = time.Now()
					} else if ev.Rawcode == 13 || (typing && ev.Rawcode == 27) { //enter or esc
						typing = !typing
					} else {
						fmt.Printf("Hook: %v %v %v\n", ev.Rawcode, ev.Mask, ev.Keycode)
					}
				case hook.MouseDown:
					if ev.Button == hook.MouseMap["center"] {
						//PrintScreen
						robotgo.KeyTap("printscreen")

						//Find coords
						/*
							x, y := robotgo.Location()
							fmt.Printf("🖱️ center-click at: x=%d, y=%d\n", x, y)
						*/
					}
				}
			}
		}
	}()
}

func runFishingBotLogic() {
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

	for {
		waitUntilFocused()
		if !fishing {
			time.Sleep(1000 * time.Millisecond)
			statusText = "PAUSED"
			continue
		}

		if currentTier == 0 {
			fmt.Println("⚠️ No tier selected. Please press Ctrl+F7 to choose a tier.")
			time.Sleep(1 * time.Second)
			continue
		}

		screenBounds := screenshot.GetDisplayBounds(0)
		screenWidth := screenBounds.Dx()
		screenHeight := screenBounds.Dy()

		waitUntilFocused()
		key := fmt.Sprintf("%dx%d", screenWidth, screenHeight)
		res, ok := resolutions[key]
		if !ok {
			fmt.Printf("⚠️ Unsupported resolution: %s\n", key)
			fishing = false
			statusText = "PAUSED"
			continue
		}
		resolution = res
		logDebug("YellowCheck Coords x:%v y:%v", resolution.YellowCheck[0], resolution.YellowCheck[1])
		logDebug("GreenCheck Coords x:%v y:%v", resolution.GreenCheck[0], resolution.GreenCheck[1])
		logDebug("MoveCheck Coords x:%v y:%v", resolution.MoveCheck[0], resolution.MoveCheck[1])

		if currentTier != lastTier {
			fmt.Printf("⬆️  Tier changed to %d — dropping fish...\n", currentTier)
			if lastTier != 0 {
				deaths = dropAllFish(4, screenWidth, screenHeight)
			}
			lastTier = currentTier
		}

		fmt.Println("Selecting hero...")
		waitUntilFocused()
		robotgo.KeyTap("f1")
		time.Sleep(500 * time.Millisecond)

		heroCheckx := resolution.HeroCheck[0]
		heroChecky := resolution.HeroCheck[1]
		logDebug("HeroCheck Coords x:%v y:%v", heroCheckx, heroChecky)
		waitUntilFocused()
		if utils.IsBlack(heroCheckx, heroChecky, screenBounds) {
			fmt.Println("⚠️ Hero is dead or not selected. Retrying...")
			if currentTier == 5 && fishing {
				fishing = false
				statusText = "PAUSED"
				fmt.Println("⏸️  Bot PAUSED for Tier 5 Death")
			}
			deaths++
			time.Sleep(15 * time.Second)
			continue
		}

		if !fishing {
			statusText = "PAUSED"
			fmt.Println("Fishing was paused - F8 to re-enable")
			continue
		}

		fmt.Println("Hero alive. Moving to fishing spot...")
		deaths = dropAllFish(deaths, screenWidth, screenHeight)

		// Move to fishing spot
		fishX, fishY := resolution.FishingSpots[currentTier][0], resolution.FishingSpots[currentTier][1]
		waitUntilFocused()
		logDebug("Move Mouse Coords x:%v y:%v", fishX, fishY)
		robotgo.Move(fishX, fishY, 0)
		robotgo.Click("right")

		if !utils.WaitInterruptible(&fishing, time.Duration(tiers[currentTier].WaitSec)*time.Second, 250*time.Millisecond) {
			fmt.Println("⏹️  Wait interrupted due to pause")
			statusText = "PAUSED"
			continue
		}

		fmt.Println("Casting fishing skill...")
		waitUntilFocused()
		waitUntilNotTyping()
		robotgo.KeyTap("num8")

		lastRedDetected := time.Now()
		for {
			waitUntilFocused()
			img, err := screenshot.CaptureRect(screenBounds)
			if err != nil {
				log.Printf("Screen capture failed: %v", err)
				continue
			}

			r1, g1, b1 := utils.GetRGB(img.At(resolution.YellowCheck[0], resolution.YellowCheck[1]))
			r2, g2, b2 := utils.GetRGB(img.At(resolution.GreenCheck[0], resolution.GreenCheck[1]))
			rm, gm, bm := utils.GetRGB(img.At(resolution.MoveCheck[0], resolution.MoveCheck[1]))

			if utils.IsYellowish(r1, g1, b1) {
				fmt.Println("🎯 Yellow detected → pressing UP")
				waitUntilFocused()
				waitUntilNotTyping()
				robotgo.KeyTap("up")
			} else if utils.IsGreenish(r2, g2, b2) {
				fmt.Println("🎯 Green detected → pressing DOWN")
				waitUntilFocused()
				waitUntilNotTyping()
				robotgo.KeyTap("down")
			} else if utils.IsRedish(rm, gm, bm) {
				if !fishing {
					statusText = "PAUSED"
					fmt.Println("Fishing was paused - F8 to re-enable")
					break
				}

				if time.Since(lastRedDetected) >= 5*time.Second {
					lastRedDetected = time.Now()
					fmt.Println("🐟 Red detected → Recasting...")
					waitUntilFocused()
					waitUntilNotTyping()
					robotgo.KeyTap("num8")
				}
			} else if utils.IsBlack(heroCheckx, heroChecky, screenBounds) {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func runDodgeArrowLogic() {
	for {
		if !dodgeArrows {
			time.Sleep(1000 * time.Millisecond)
			continue
		}

		waitUntilFocused()
		screenBounds := screenshot.GetDisplayBounds(0)
		screenWidth := screenBounds.Dx()
		screenHeight := screenBounds.Dy()

		key := fmt.Sprintf("%dx%d", screenWidth, screenHeight)
		res, ok := resolutions[key]
		if !ok {
			fmt.Printf("⚠️ Unsupported resolution: %s\n", key)
			dodgeArrows = false
			dodgeText = "PAUSED"
			continue
		}
		resolution = res

		for {
			waitUntilFocused()
			if !dodgeArrows {
				break
			}

			waitUntilNotTyping()
			img, err := screenshot.CaptureRect(screenBounds)
			if err != nil {
				log.Printf("Screen capture failed: %v", err)
				continue
			}

			if utils.IsRedish(utils.GetRGB(img.At(resolution.DodgeArrows["red"][0][0], resolution.DodgeArrows["red"][0][1]))) {
				if utils.CheckDodge(img, resolution.DodgeArrows["up"]) {
					fmt.Println("🎯 Up detected → pressing UP")
					robotgo.KeyTap("up")
					robotgo.KeyTap("down")
					robotgo.KeyTap("esc")
				} else if utils.CheckDodge(img, resolution.DodgeArrows["down"]) {
					fmt.Println("🎯 Down detected → pressing DOWN")
					robotgo.KeyTap("down")
					robotgo.KeyTap("up")
					robotgo.KeyTap("esc")
				} else if utils.CheckDodge(img, resolution.DodgeArrows["left"]) {
					fmt.Println("🎯 Left detected → pressing LEFT")
					robotgo.KeyTap("left")
					robotgo.KeyTap("right")
					robotgo.KeyTap("esc")
				} else if utils.CheckDodge(img, resolution.DodgeArrows["right"]) {
					fmt.Println("🎯 Right detected → pressing RIGHT")
					robotgo.KeyTap("right")
					robotgo.KeyTap("left")
					robotgo.KeyTap("esc")
				}
			}

			time.Sleep(500 * time.Millisecond)
		}
	}
}

func runRefreshManaLogic() {
	go func() {
		for {
			if !pauseManaRefresh && manaRefresh && time.Since(lastActionTime) >= time.Duration(manaRefreshCurrent)*time.Millisecond {
				fmt.Println("⏹️  ManaRefresh interrupted due to no actions in ", manaRefreshCurrent)
				manaStatusText = "TEMP PAUSED"
				pauseManaRefresh = true
			} else if pauseManaRefresh && manaRefresh && time.Since(lastActionTime) <= time.Duration(manaRefreshCurrent)*time.Millisecond {
				fmt.Println("✅ ManaRefresh restarting on ", manaRefreshCurrent)
				manaStatusText = "ENABLED " + strconv.Itoa(manaRefreshCurrent)[:2]
				pauseManaRefresh = false
			} else if !manaRefresh {
				manaStatusText = "PAUSED"
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	for {
		if !manaRefresh {
			time.Sleep(5000 * time.Millisecond)
			continue
		}

		for {

			if !manaRefresh {
				manaStatusText = "PAUSED"
				break
			} else if manaRefresh && pauseManaRefresh {
				time.Sleep(500 * time.Millisecond)
				continue
			}

			fmt.Println("💙⚗️  Pressing Mana Refresh")
			waitUntilNotTyping()
			robotgo.KeyTap("num8")
			fmt.Printf("💙⚗️  Refreshing in %v\n", manaRefreshCurrent)
			if !utils.WaitInterruptible(&manaRefresh, time.Duration(manaRefreshCurrent)*time.Millisecond, 1*time.Second) {
				fmt.Println("⏹️  Wait interrupted due to pause")
				manaStatusText = "PAUSED"
				continue
			}
		}
	}
}

func main() {
	lastActionTime = time.Now()
	initGL()
	window := createOverlayWindow()
	makeWindowClickThrough(window)
	screenBounds := screenshot.GetDisplayBounds(0)
	screenHeight := screenBounds.Dy()
	setupFont(screenHeight)
	go hookLogic()
	go runDodgeArrowLogic()
	go runFishingBotLogic()
	go runRefreshManaLogic()
	renderOverlayLoop(window)
}

func dropAllFish(deaths, screenWidth, screenHeight int) int {
	waitUntilFocused()
	key := fmt.Sprintf("%dx%d", screenWidth, screenHeight)
	res, ok := resolutions[key]
	if !ok {
		fmt.Printf("Unsupported resolution: %s\n", key)
		return deaths
	}
	resolution = res

	if deaths >= 4 {
		waitUntilFocused()
		fmt.Println("Dropping Fish")
		robotgo.KeyTap("f1")
		robotgo.KeyTap("f1")

		for _, slot := range resolution.InventorySlots {
			waitUntilFocused()
			x, y := slot[0], slot[1]
			logDebug("Inventory Mouse Coords x:%v y:%v", x, y)
			robotgo.Move(x, y)
			robotgo.Click("right")

			waitUntilFocused()
			dx, dy := resolution.DropPoint[0], resolution.DropPoint[1]
			logDebug("InventoryHero Mouse Coords x:%v y:%v", x, y)
			robotgo.Move(dx, dy)
			robotgo.Click()
		}

		time.Sleep(1 * time.Second)
		deaths = 0
	}
	return deaths
}

func waitUntilFocused() {
	for !windowFocused {
		time.Sleep(500 * time.Millisecond)
	}
}

func waitUntilNotTyping() {
	for typing {
		fmt.Println("waitUntilNotTyping")
		time.Sleep(250 * time.Millisecond)
	}
}
