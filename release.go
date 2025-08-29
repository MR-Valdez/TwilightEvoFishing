//go:build !debug

package main

func logDebug(msg string, args ...any) {
	// no-op in release
}

func printDebugInfo() {
	// no-op in release
}
