//go:build debug

package main

import (
	"fmt"

	"github.com/go-vgo/robotgo"
	"github.com/lxn/win"
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

	hwnd := win.GetForegroundWindow()
	if hwnd == 0 {
		fmt.Println("No foreground window")
		return
	}
	dpi := win.GetDpiForWindow(hwnd)

	fmt.Println("------ DEBUG INFO ------")
	fmt.Printf("Mouse: (%d, %d)\n", mouseX, mouseY)
	fmt.Printf("Window: origin=(%d, %d), size=(%d x %d)\n", winX, winY, winW, winH)
	fmt.Printf("Mouse relative to window: (%d, %d)\n", mouseX-winX, mouseY-winY)
	fmt.Printf("DPI for window: %d (100%% scaling = 96 DPI)\n", dpi)
	fmt.Println("-------------------------")
}
