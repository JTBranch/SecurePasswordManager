package main

import (
	"fmt"
	"runtime"

	"go-password-manager/internal/transport/bluetooth"
)

func main() {
	fmt.Println("GOOS:", runtime.GOOS)
	a, err := bluetooth.GetSystemAdapter("")
	if err != nil {
		fmt.Println("GetSystemAdapter error:", err)
		return
	}
	if a == nil {
		fmt.Println("GetSystemAdapter returned nil — no adapter registered or initialization failed")
		return
	}
	fmt.Printf("Adapter present: type=%T\n", a)
}
