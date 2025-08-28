//go:build !debug

package main

func logDebug(msg string, args ...any) {
	// no-op in release
}
