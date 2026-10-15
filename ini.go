package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// ===== CitizenFX.ini ของ RedM =====
// ตำแหน่ง: %LOCALAPPDATA%\RedM\RedM.app\CitizenFX.ini (ถ้าไม่มีลอง %LOCALAPPDATA%\RedM\CitizenFX.ini)
// แก้เฉพาะคีย์ที่กำหนดใน section [Game] · บรรทัดอื่นของผู้เล่น (IVPath ฯลฯ) เก็บไว้เหมือนเดิม · สำรองไฟล์เดิมครั้งแรก

func redmDir() string { return filepath.Join(os.Getenv("LOCALAPPDATA"), "RedM") }

func redmExe() string {
	p := filepath.Join(redmDir(), "RedM.exe")
	if fileExists(p) {
		return p
	}
	return ""
}

func iniPath() (string, bool) {
	c := []string{filepath.Join(redmDir(), "RedM.app", "CitizenFX.ini"), filepath.Join(redmDir(), "CitizenFX.ini")}
	for _, p := range c {
		if fileExists(p) {
			return p, true
		}
	}
	return c[0], false
}

func fileExists(p string) bool { st, err := os.Stat(p); return err == nil && !st.IsDir() }

type iniItem struct {
	Label string `json:"label"`
	Short string `json:"short"`
	Ok    bool   `json:"ok"`
}
type IniState struct {
	Path   string    `json:"path"`
	Exists bool      `json:"exists"`
	Items  []iniItem `json:"items"`
	AllOk  bool      `json:"allOk"`
	Want   []string  `json:"want"`
}

// อ่านค่าคีย์ใน [Game] (ไม่สนตัวพิมพ์ใหญ่เล็กของชื่อ section/คีย์)
func readIni(p string) map[string]string {
	out := map[string]string{}
	b, err := os.ReadFile(p)
	if err != nil {
		return out
	}
	b = bytes.TrimPrefix(b, []byte{0xEF, 0xBB, 0xBF})
	sec := ""
	for _, ln := range strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n") {
		t := strings.TrimSpace(ln)
		if strings.HasPrefix(t, "[") && strings.HasSuffix(t, "]") {
			sec = strings.ToLower(t)
			continue
		}
		if sec == "[game]" {
			if i := strings.Index(t, "="); i > 0 {
				out[strings.ToLower(strings.TrimSpace(t[:i]))] = strings.TrimSpace(t[i+1:])
			}
		}
	}
	return out
}

func checkIni(want []IniKV) IniState {
	p, ex := iniPath()
	cur := readIni(p)
	st := IniState{Path: p, Exists: ex, AllOk: true}
	for _, kv := range want {
		ok := cur[strings.ToLower(kv.Key)] == kv.Value
		if !ok {
			st.AllOk = false
		}
		st.Items = append(st.Items, iniItem{Label: kv.Label, Short: kv.Short, Ok: ok})
		st.Want = append(st.Want, kv.Key+"="+kv.Value)
	}
	return st
}

// เขียนค่า (ต้องปิด RedM ก่อน ไม่งั้น RedM เขียนทับตอนปิดเกม)
func writeIni(want []IniKV) (string, error) {
	if redmRunning() {
		return "", errors.New("ปิด RedM ให้หมดก่อน แล้วกดอีกครั้ง (RedM เขียนไฟล์ทับตอนปิดเกม)")
	}
	p, ex := iniPath()
	var lines []string
	if ex {
		b, err := os.ReadFile(p)
		if err != nil {
			return p, err
		}
		b = bytes.TrimPrefix(b, []byte{0xEF, 0xBB, 0xBF})
		lines = strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n")
		bak := p + ".redsun-backup"
		if !fileExists(bak) {
			os.WriteFile(bak, b, 0o644)
		}
	} else {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return p, err
		}
	}
	// หา section [Game]
	start, end := -1, len(lines)
	for i, ln := range lines {
		t := strings.ToLower(strings.TrimSpace(ln))
		if t == "[game]" {
			start = i
			continue
		}
		if start >= 0 && strings.HasPrefix(t, "[") && strings.HasSuffix(t, "]") {
			end = i
			break
		}
	}
	if start < 0 {
		lines = append([]string{"[Game]"}, lines...)
		start, end = 0, len(lines)
		for i := 1; i < len(lines); i++ {
			t := strings.TrimSpace(lines[i])
			if strings.HasPrefix(t, "[") {
				end = i
				break
			}
		}
	}
	for _, kv := range want {
		found := false
		for i := start + 1; i < end; i++ {
			t := strings.TrimSpace(lines[i])
			if j := strings.Index(t, "="); j > 0 && strings.EqualFold(strings.TrimSpace(t[:j]), kv.Key) {
				lines[i] = kv.Key + "=" + kv.Value
				found = true
			}
		}
		if !found {
			// แทรกท้าย section [Game] (ก่อนบรรทัดว่างท้าย section)
			at := end
			for at-1 > start && strings.TrimSpace(lines[at-1]) == "" {
				at--
			}
			lines = append(lines[:at], append([]string{kv.Key + "=" + kv.Value}, lines[at:]...)...)
			end++
		}
	}
	out := strings.Join(lines, "\r\n")
	if !strings.HasSuffix(out, "\r\n") {
		out += "\r\n"
	}
	return p, os.WriteFile(p, []byte(out), 0o644)
}
