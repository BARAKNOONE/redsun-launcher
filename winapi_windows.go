//go:build windows

package main

import (
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ===== Win32: หน้าต่างไม่มีกรอบ · ลากย้าย · มุมโค้ง Win11 · เช็คโปรเซส RedM · เปิดลิงก์ =====
var (
	user32            = windows.NewLazySystemDLL("user32.dll")
	dwmapi            = windows.NewLazySystemDLL("dwmapi.dll")
	pGetWindowLongPtr = user32.NewProc("GetWindowLongPtrW")
	pSetWindowLongPtr = user32.NewProc("SetWindowLongPtrW")
	pSetWindowPos     = user32.NewProc("SetWindowPos")
	pReleaseCapture   = user32.NewProc("ReleaseCapture")
	pPostMessage      = user32.NewProc("PostMessageW")
	pShowWindow       = user32.NewProc("ShowWindow")
	pMessageBox       = user32.NewProc("MessageBoxW")
	pDwmSetAttr       = dwmapi.NewProc("DwmSetWindowAttribute")
)

const (
	gwlStyle        = ^uintptr(15) // -16
	wsCaption       = 0x00C00000
	wsThickFrame    = 0x00040000
	wsSysMenu       = 0x00080000
	swpFrameChanged = 0x0020
	swpNoMove       = 0x0002
	swpNoSize       = 0x0001
	swpNoZOrder     = 0x0004
	wmNcLButtonDown = 0x00A1
	htCaption       = 2
	swMinimize      = 6
)

func makeFrameless(hwnd uintptr) {
	style, _, _ := pGetWindowLongPtr.Call(hwnd, gwlStyle)
	style &^= wsCaption | wsThickFrame
	style |= wsSysMenu
	pSetWindowLongPtr.Call(hwnd, gwlStyle, style)
	pSetWindowPos.Call(hwnd, 0, 0, 0, 0, 0, swpFrameChanged|swpNoMove|swpNoSize|swpNoZOrder)
	pref := uint32(2) // DWMWCP_ROUND (Windows 11 · Win10 ไม่มีผล)
	pDwmSetAttr.Call(hwnd, 33, uintptr(unsafe.Pointer(&pref)), 4)
}
func dragWindow(hwnd uintptr) {
	pReleaseCapture.Call()
	pPostMessage.Call(hwnd, wmNcLButtonDown, htCaption, 0)
}
func minimizeWindow(hwnd uintptr) { pShowWindow.Call(hwnd, swMinimize) }

func msgBox(title, text string) {
	t, _ := syscall.UTF16PtrFromString(title)
	m, _ := syscall.UTF16PtrFromString(text)
	pMessageBox.Call(0, uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(t)), 0x40)
}

// เปิดลิงก์/โปรโตคอล (https:// · redm://) ด้วยโปรแกรมของ Windows
func shellOpen(target string) error {
	verb, _ := syscall.UTF16PtrFromString("open")
	f, _ := syscall.UTF16PtrFromString(target)
	return windows.ShellExecute(0, verb, f, nil, nil, windows.SW_SHOWNORMAL)
}

// มี RedM เปิดอยู่ไหม (RedM.exe / RedM_b1491_GTAProcess.exe ฯลฯ)
func redmRunning() bool {
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return false
	}
	defer windows.CloseHandle(snap)
	var e windows.ProcessEntry32
	e.Size = uint32(unsafe.Sizeof(e))
	for err = windows.Process32First(snap, &e); err == nil; err = windows.Process32Next(snap, &e) {
		n := strings.ToLower(windows.UTF16ToString(e.ExeFile[:]))
		if strings.HasPrefix(n, "redm") && strings.HasSuffix(n, ".exe") {
			return true
		}
	}
	return false
}

func unsafePtr(p *uint16) unsafe.Pointer { return unsafe.Pointer(p) }
