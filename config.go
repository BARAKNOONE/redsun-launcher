package main

import (
	_ "embed"
	"encoding/json"
)

//go:embed remote/launcher.json
var embeddedRemote []byte

// ===== ค่าที่ทีมแก้ได้ (หรือใส่ตอน build ด้วย -ldflags "-X main.Version=1.0.1 -X main.Repo=owner/repo") =====
var (
	Version = "1.0.0"                         // เวอร์ชัน Launcher (ตรงกับ tag บน GitHub เช่น v1.0.0)
	Repo    = "REPLACE_OWNER/redsun-launcher" // GitHub repo ที่โฮส Launcher + ไฟล์ remote/launcher.json
	Branch  = "main"
)

const (
	AppName   = "REDSUN Launcher"
	ExeAsset  = "RedsunLauncher.exe"        // ชื่อไฟล์ใน GitHub Release
	ShaAsset  = "RedsunLauncher.exe.sha256" // ไฟล์ SHA-256 คู่กัน (บังคับ ไม่มี = ไม่อัปเดต)
	WinWidth  = 1120
	WinHeight = 680
)

func remoteURL() string {
	return "https://raw.githubusercontent.com/" + Repo + "/" + Branch + "/remote/launcher.json"
}
func releaseAPI() string { return "https://api.github.com/repos/" + Repo + "/releases/latest" }

// ค่าเริ่มต้น = ไฟล์ remote/launcher.json ที่ฝังมากับตัว exe (ใช้ตอนยังไม่ตั้ง GitHub หรือออฟไลน์)
func defaultRemote() Remote {
	var r Remote
	json.Unmarshal(embeddedRemote, &r)
	if r.Server.Max == 0 {
		r.Server.Max = 2000
	}
	return r
}
