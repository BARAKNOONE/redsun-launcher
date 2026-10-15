package main

import (
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// ===== สถานะรวมที่ส่งให้หน้า UI =====
type Msg struct {
	T    string `json:"t"`
	Warn bool   `json:"warn"`
}
type State struct {
	Version     string     `json:"version"`
	Remote      Remote     `json:"remote"`
	Online      bool       `json:"online"` // โหลด launcher.json จาก GitHub ได้
	Status      Status     `json:"status"`
	Ini         IniState   `json:"ini"`
	Redm        RedmState  `json:"redm"`
	Update      UpdateInfo `json:"update"`
	MustUpdate  bool       `json:"mustUpdate"`
	Setup       bool       `json:"setup"` // เปิดจากไฟล์ดาวน์โหลด (ยังไม่ติดตั้ง) = โชว์หน้าติดตั้ง
	InstallPath string     `json:"installPath"`
	Message     *Msg       `json:"message,omitempty"`
}
type RedmState struct {
	Found   bool   `json:"found"`
	Path    string `json:"path"`
	Running bool   `json:"running"`
}

var portable bool // --portable = ไม่ถามติดตั้ง

type App struct {
	mu      sync.Mutex
	st      State
	loading bool
	push    func(State) // ส่งสถานะใหม่ไป UI (ตั้งใน main)
}

func newApp() *App {
	a := &App{}
	a.st.Version = Version
	a.st.Setup = !isInstalled() && !portable
	a.st.InstallPath = installDir()
	a.st.Remote = defaultRemote()
	a.local()
	return a
}

// ข้อมูลในเครื่อง (เร็ว ไม่ใช้เน็ต)
func (a *App) local() {
	a.st.Ini = checkIni(a.st.Remote.CitizenFX)
	exe := redmExe()
	a.st.Redm = RedmState{Found: exe != "", Path: exe, Running: redmRunning()}
	a.st.MustUpdate = a.st.Remote.MinVersion != "" && newer(a.st.Remote.MinVersion, Version)
}

func (a *App) snapshot() State {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.local()
	return a.st
}

// โหลดข่าว/ค่าตั้ง/สถานะเซิร์ฟ/อัปเดต เบื้องหลัง แล้วส่งให้ UI
func (a *App) refresh() {
	a.mu.Lock()
	if a.loading {
		a.mu.Unlock()
		return
	}
	a.loading = true
	a.mu.Unlock()
	go func() {
		r, ok := fetchRemote()
		st := fetchStatus(r.Server.Status)
		up, uerr := checkUpdate()
		a.mu.Lock()
		a.st.Remote, a.st.Online, a.st.Status = r, ok, st
		if uerr == nil {
			a.st.Update = up
		}
		if !ok && !strings.HasPrefix(Repo, "REPLACE") {
			a.st.Message = &Msg{T: "ออฟไลน์ · ใช้ข่าวและค่าตั้งล่าสุดที่เคยโหลดไว้", Warn: true}
		} else {
			a.st.Message = nil
		}
		a.local()
		s := a.st
		a.loading = false
		a.mu.Unlock()
		if a.push != nil {
			a.push(s)
		}
	}()
}

type Result struct {
	Ok  bool   `json:"ok"`
	Msg string `json:"msg"`
}

func (a *App) fixIni() Result {
	a.mu.Lock()
	want := a.st.Remote.CitizenFX
	a.mu.Unlock()
	p, err := writeIni(want)
	if err != nil {
		return Result{false, err.Error()}
	}
	return Result{true, "บันทึกค่าลง " + p + " แล้ว"}
}

// เข้าเซิร์ฟ: ตั้งค่า ini → เปิด RedM เชื่อมต่อเซิร์ฟทันที
func (a *App) play(step func(int, string)) Result {
	a.mu.Lock()
	r := a.st.Remote
	must := a.st.MustUpdate
	a.mu.Unlock()
	if must {
		return Result{false, "ต้องอัปเดต Launcher ก่อนเข้าเซิร์ฟ"}
	}
	if r.Server.Maintenance {
		return Result{false, "เซิร์ฟปิดปรับปรุงอยู่"}
	}
	exe := redmExe()
	if exe == "" {
		return Result{false, "ไม่พบ RedM ในเครื่อง · ติดตั้งจาก redm.net ก่อน"}
	}
	step(0, "ตรวจไฟล์ตั้งค่า…")
	if !checkIni(r.CitizenFX).AllOk {
		if redmRunning() {
			return Result{false, "ปิด RedM ให้หมดก่อน แล้วกดเข้าเซิร์ฟอีกครั้ง (ต้องแก้ค่าตั้งเกม)"}
		}
		if _, err := writeIni(r.CitizenFX); err != nil {
			return Result{false, "เขียน CitizenFX.ini ไม่ได้: " + err.Error()}
		}
	}
	step(1, "เปิด RedM…")
	addr := strings.TrimSpace(r.Server.Connect)
	if addr == "" {
		if err := exec.Command(exe).Start(); err != nil {
			return Result{false, "เปิด RedM ไม่ได้: " + err.Error()}
		}
		return Result{true, "เปิด RedM แล้ว (ยังไม่ได้ตั้งที่อยู่เซิร์ฟใน launcher.json)"}
	}
	step(2, "เชื่อมต่อ "+r.Server.Name+"…")
	// ใช้โปรโตคอล redm://connect/ ก่อน · ไม่ได้ค่อยเปิด exe พร้อม +connect
	if err := shellOpen("redm://connect/" + addr); err != nil {
		if err2 := exec.Command(exe, "+connect", addr).Start(); err2 != nil {
			return Result{false, "เปิด RedM ไม่ได้: " + err2.Error()}
		}
	}
	return Result{true, "เปิด RedM แล้ว · กำลังเชื่อมต่อ " + r.Server.Name}
}

func (a *App) closeOnPlay() bool { a.mu.Lock(); defer a.mu.Unlock(); return a.st.Remote.CloseOnPlay }

func (a *App) startUpdate(progress func(float64, string), fail func(string), done func()) {
	a.mu.Lock()
	u := a.st.Update
	a.mu.Unlock()
	go func() {
		if err := applyUpdate(u, progress); err != nil {
			fail("อัปเดตไม่สำเร็จ: " + err.Error())
			return
		}
		time.Sleep(300 * time.Millisecond)
		done()
	}()
}

func openFolder() { os.MkdirAll(redmDir(), 0o755); shellOpen(redmDir()) }
