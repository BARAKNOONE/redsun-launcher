//go:build !windows

package main

import "fmt"

// บน Linux ใช้ทดสอบตรรกะเท่านั้น (go test) · ตัวจริง build ด้วย GOOS=windows
func main() { fmt.Println("REDSUN Launcher ใช้ได้บน Windows เท่านั้น") }
