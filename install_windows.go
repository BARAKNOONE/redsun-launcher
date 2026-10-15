//go:build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// ===== ติดตั้ง / ถอนการติดตั้ง (ไม่ต้องใช้สิทธิ์แอดมิน · ติดตั้งในโฟลเดอร์ผู้ใช้) =====
//
//	ที่ติดตั้ง: %LOCALAPPDATA%\Programs\RedsunLauncher\RedsunLauncher.exe
//	ไอคอน: เดสก์ท็อป + เมนู Start · รายการใน "แอปและฟีเจอร์" (ถอนได้จาก Windows)
const uninstKey = `Software\Microsoft\Windows\CurrentVersion\Uninstall\RedsunLauncher`

func installDir() string {
	return filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "RedsunLauncher")
}
func installExe() string { return filepath.Join(installDir(), ExeAsset) }

// รันจากที่ติดตั้งอยู่แล้วไหม
func isInstalled() bool {
	self, err := os.Executable()
	if err != nil {
		return true
	}
	return strings.EqualFold(filepath.Clean(filepath.Dir(self)), filepath.Clean(installDir()))
}

func desktopLnk() string {
	return filepath.Join(os.Getenv("USERPROFILE"), "Desktop", "REDSUN Launcher.lnk")
}
func startLnk() string {
	return filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "REDSUN Launcher.lnk")
}

func psq(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }

// สร้างไอคอนลัดด้วย PowerShell (WScript.Shell) · ซ่อนหน้าต่าง
func makeShortcut(lnk, target string) error {
	os.MkdirAll(filepath.Dir(lnk), 0o755)
	ps := "$s=(New-Object -ComObject WScript.Shell).CreateShortcut(" + psq(lnk) + ");$s.TargetPath=" + psq(target) +
		";$s.WorkingDirectory=" + psq(filepath.Dir(target)) + ";$s.IconLocation=" + psq(target+",0") + ";$s.Description='REDSUN Roleplay Launcher';$s.Save()"
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", ps)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	return cmd.Run()
}

func copyFile(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	tmp := dst + ".tmp"
	if err := os.WriteFile(tmp, b, 0o755); err != nil {
		return err
	}
	os.Remove(dst)
	return os.Rename(tmp, dst)
}

func writeUninstallEntry() {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, uninstKey, registry.ALL_ACCESS)
	if err != nil {
		return
	}
	defer k.Close()
	exe := installExe()
	k.SetStringValue("DisplayName", "REDSUN Launcher")
	k.SetStringValue("DisplayVersion", Version)
	k.SetStringValue("Publisher", "REDSUN Roleplay")
	k.SetStringValue("DisplayIcon", exe+",0")
	k.SetStringValue("InstallLocation", installDir())
	k.SetStringValue("UninstallString", `"`+exe+`" --uninstall`)
	k.SetDWordValue("NoModify", 1)
	k.SetDWordValue("NoRepair", 1)
	k.SetDWordValue("EstimatedSize", 8000)
}

// ติดตั้ง: คัดลอกตัวเอง → ไอคอนลัด → รายการถอนการติดตั้ง → เปิดตัวที่ติดตั้ง
func install(desktop, start bool) Result {
	self, err := os.Executable()
	if err != nil {
		return Result{false, err.Error()}
	}
	if err := os.MkdirAll(installDir(), 0o755); err != nil {
		return Result{false, "สร้างโฟลเดอร์ไม่ได้: " + err.Error()}
	}
	if err := copyFile(self, installExe()); err != nil {
		return Result{false, "คัดลอกไฟล์ไม่ได้ (ปิด Launcher ตัวที่เปิดอยู่ก่อน): " + err.Error()}
	}
	if desktop {
		makeShortcut(desktopLnk(), installExe())
	}
	if start {
		makeShortcut(startLnk(), installExe())
	}
	writeUninstallEntry()
	if err := exec.Command(installExe(), "--installed").Start(); err != nil {
		return Result{false, "ติดตั้งแล้วแต่เปิดไม่ได้: " + err.Error()}
	}
	return Result{true, "ติดตั้งเสร็จ"}
}

// เวอร์ชันในรายการ "แอปและฟีเจอร์" ให้ตรงหลังอัปเดตตัวเอง
func touchInstalled() {
	if isInstalled() {
		if k, err := registry.OpenKey(registry.CURRENT_USER, uninstKey, registry.SET_VALUE); err == nil {
			k.SetStringValue("DisplayVersion", Version)
			k.Close()
		}
	}
}

// ถอนการติดตั้ง (เรียกจาก Windows: RedsunLauncher.exe --uninstall)
func uninstall() {
	if !askYesNo(AppName, "ถอนการติดตั้ง REDSUN Launcher ออกจากเครื่องนี้?\n(ไม่ลบ RedM และไฟล์เกม)") {
		return
	}
	os.Remove(desktopLnk())
	os.Remove(startLnk())
	registry.DeleteKey(registry.CURRENT_USER, uninstKey)
	os.RemoveAll(appDir())
	// ลบโฟลเดอร์โปรแกรมหลังตัวเองปิด
	cmd := exec.Command("cmd", "/C", "ping 127.0.0.1 -n 3 > nul & rmdir /s /q \""+installDir()+"\"")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000 | windows.DETACHED_PROCESS}
	cmd.Start()
	msgBox(AppName, "ถอนการติดตั้งเรียบร้อย")
}

var pMessageBoxYN = windows.NewLazySystemDLL("user32.dll").NewProc("MessageBoxW")

func askYesNo(title, text string) bool {
	t, _ := syscall.UTF16PtrFromString(title)
	m, _ := syscall.UTF16PtrFromString(text)
	r, _, _ := pMessageBoxYN.Call(0, uintptr(unsafePtr(m)), uintptr(unsafePtr(t)), 0x24) // MB_YESNO | MB_ICONQUESTION
	return r == 6
}
