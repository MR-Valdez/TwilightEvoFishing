package utils

import (
	"image"
	"image/color"
	"log"
	"syscall"
	"time"

	"github.com/kbinani/screenshot"
	"github.com/lxn/win"
)

func GetRGB(c color.Color) (uint32, uint32, uint32) {
	r, g, b, _ := color.RGBAModel.Convert(c).RGBA()
	return r >> 8, g >> 8, b >> 8
}

func IsBlack(x, y int, bounds image.Rectangle) bool {
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		log.Printf("Screen capture failed: %v", err)
		return false
	}
	r, g, b := GetRGB(img.At(x, y))
	return r == 0 && g == 0 && b == 0
}

func IsYellowish(r, g, b uint32) bool {
	return r > 180 && g > 180 && b < 100
}

func IsGreenish(r, g, b uint32) bool {
	return g > 180 && r < 150 && b < 150
}

func IsRedish(r, g, b uint32) bool {
	return r > 180 && g < 120 && b < 120
}

func WaitInterruptible(check *bool, d time.Duration, checkInterval time.Duration) bool {
	elapsed := time.Duration(0)
	for elapsed < d {
		if !*check {
			return false
		}
		time.Sleep(checkInterval)
		elapsed += checkInterval
	}
	return true
}

func CheckDodge(img *image.RGBA, coords [][]int, dpiScale int) bool {
	coordCount := len(coords)
	matchCount := 0
	for _, coordPair := range coords {
		if IsYellowish(GetRGB(img.At(ScaleCoords(coordPair[0], coordPair[1], dpiScale)))) {
			matchCount++
		}
	}
	return coordCount == matchCount
}

func GetWindowDPI() int {
	// Windows 10+ API
	title, err := syscall.UTF16PtrFromString("Warcraft III")
	if err != nil {
		return 96
	}

	hwnd := win.FindWindow(nil, title)
	if hwnd == 0 {
		return 96
	}

	dpi := win.GetDpiForWindow(hwnd)
	if dpi == 0 {
		// fallback for older systems
		dpi = 96
	}
	return int(dpi)
}

func ScaleCoords(x, y int, dpi int) (int, int) {
	scale := float64(dpi) / 96.0
	return int(float64(x) * scale), int(float64(y) * scale)
}
