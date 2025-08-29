//go:build debug

package main

import (
	"fmt"

	"github.com/go-vgo/robotgo"
)

func logDebug(msg string, args ...any) {
	fmt.Printf("[DEBUG] "+msg+"\n", args...)
}

var bounds struct {
	X int
	Y int
	W int
	H int
}

func printDebugInfo() {
	pids, _ := robotgo.FindIds("Warcraft III")
	if len(pids) > 0 {
		bounds.X, bounds.Y, bounds.W, bounds.H = robotgo.GetBounds(pids[0])
	}

	// Get mouse position
	mouseX, mouseY := robotgo.Location()

	// Get active window bounds

	winX, winY, winW, winH := bounds.X, bounds.Y, bounds.W, bounds.H

	// Try to read the pixel color under the mouse
	bitmap := robotgo.CaptureScreen(mouseX, mouseY, 1, 1)
	defer robotgo.FreeBitmap(bitmap)

	hexString := robotgo.GetPixelColor(mouseX, mouseY)

	fmt.Println("------ DEBUG INFO ------")
	fmt.Printf("Mouse: (%d, %d)\n", mouseX, mouseY)
	fmt.Printf("Window: origin=(%d, %d), size=(%d x %d)\n", winX, winY, winW, winH)
	fmt.Printf("Mouse relative to window: (%d, %d)\n", mouseX-winX, mouseY-winY)
	fmt.Printf("Pixel color: #%s \n", hexString) //(R:%d G:%d B:%d)
	fmt.Println("-------------------------")
}
