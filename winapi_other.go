//go:build !windows

package main

// ตัวแทนสำหรับทดสอบบน Linux (Launcher ใช้งานจริงบน Windows เท่านั้น)
var fakeRunning = false

func redmRunning() bool      { return fakeRunning }
func shellOpen(string) error { return nil }
func msgBox(string, string)  {}
func makeFrameless(uintptr)  {}
func dragWindow(uintptr)     {}
func minimizeWindow(uintptr) {}

func isInstalled() bool         { return true }
func installDir() string        { return "" }
func install(bool, bool) Result { return Result{true, ""} }
func touchInstalled()           {}
func uninstall()                {}
