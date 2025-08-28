//go:build debug

package main

import "fmt"

func logDebug(msg string, args ...any) {
	fmt.Printf("[DEBUG] "+msg+"\n", args...)
}
