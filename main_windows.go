//go:build windows

package main

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	webview2 "github.com/jchv/go-webview2"
)

//go:embed ui/index.html
var page string

//go:embed ui/img/logo_h.webp
var logoH []byte

//go:embed ui/img/logo_primary.webp
var logoP []byte

func main() {
	for _, a := range os.Args[1:] {
		switch a {
		case "--updated":
			go cleanupOld()
		case "--uninstall":
			uninstall()
			return
		case "--portable":
			portable = true
		}
	}
	touchInstalled()
	app := newApp()
	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug:         false,
		AutoFocus:     true,
		DataPath:      filepath.Join(appDir(), "WebView2"),
		WindowOptions: webview2.WindowOptions{Title: AppName, Width: WinWidth, Height: WinHeight, IconId: 2, Center: true},
	})
	if w == nil {
		msgBox(AppName, "เครื่องนี้ยังไม่มี Microsoft Edge WebView2 Runtime\nดาวน์โหลดฟรีที่ https://go.microsoft.com/fwlink/p/?LinkId=2124703 แล้วเปิด Launcher ใหม่")
		shellOpen("https://go.microsoft.com/fwlink/p/?LinkId=2124703")
		return
	}
	defer w.Destroy()
	hwnd := uintptr(w.Window())
	makeFrameless(hwnd)

	eval := func(js string) { w.Dispatch(func() { w.Eval(js) }) }
	jsonStr := func(v interface{}) string { b, _ := json.Marshal(v); return string(b) }
	app.push = func(s State) { eval("window.rsSet && window.rsSet(" + jsonStr(s) + ")") }

	w.Bind("rsState", func() State { return app.snapshot() })
	w.Bind("rsRefresh", func() State { app.refresh(); return app.snapshot() })
	w.Bind("rsFixIni", func() Result { return app.fixIni() })
	w.Bind("rsPlay", func() Result {
		r := app.play(func(n int, t string) { eval("window.rsStep && window.rsStep(" + jsonStr(n) + "," + jsonStr(t) + ")") })
		if r.Ok && app.closeOnPlay() {
			go func() { time.Sleep(4 * time.Second); w.Dispatch(func() { w.Terminate() }) }()
		}
		return r
	})
	w.Bind("rsUpdate", func() {
		app.startUpdate(
			func(p float64, t string) { eval("window.rsUpProgress(" + jsonStr(p) + "," + jsonStr(t) + ")") },
			func(t string) { eval("window.rsUpFail(" + jsonStr(t) + ")") },
			func() { w.Dispatch(func() { w.Terminate() }) },
		)
	})
	w.Bind("rsOpen", func(u string) {
		if strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "discord://") {
			shellOpen(u)
		}
	})
	w.Bind("rsInstall", func(desk, start bool) Result {
		r := install(desk, start)
		if r.Ok {
			go func() { time.Sleep(800 * time.Millisecond); w.Dispatch(func() { w.Terminate() }) }()
		}
		return r
	})
	w.Bind("rsOpenFolder", func() { openFolder() })
	w.Bind("rsDrag", func() { dragWindow(hwnd) })
	w.Bind("rsMin", func() { minimizeWindow(hwnd) })
	w.Bind("rsClose", func() { w.Dispatch(func() { w.Terminate() }) })

	// ฝังรูปโลโก้ในหน้า (หน้า UI ถูกโหลดจากหน่วยความจำ ไม่มีไฟล์ข้างนอก)
	html := strings.ReplaceAll(page, "img/logo_h.webp", "data:image/webp;base64,"+base64.StdEncoding.EncodeToString(logoH))
	html = strings.ReplaceAll(html, "img/logo_primary.webp", "data:image/webp;base64,"+base64.StdEncoding.EncodeToString(logoP))
	w.SetHtml(html)
	app.refresh()
	w.Run()
}
