package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// ===== อัปเดตตัวเองจาก GitHub Releases =====
// release ล่าสุดต้องมีไฟล์ RedsunLauncher.exe + RedsunLauncher.exe.sha256 · tag เช่น v1.0.1
type UpdateInfo struct {
	Available bool   `json:"available"`
	Version   string `json:"version"`
	Notes     string `json:"notes"`
	exeURL    string
	shaURL    string
}

func newer(a, b string) bool { // a ใหม่กว่า b ไหม (1.2.10 > 1.2.9)
	pa, pb := vparts(a), vparts(b)
	for i := 0; i < 3; i++ {
		if pa[i] != pb[i] {
			return pa[i] > pb[i]
		}
	}
	return false
}
func vparts(v string) [3]int {
	var o [3]int
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	for i, p := range strings.SplitN(v, ".", 3) {
		n, _ := strconv.Atoi(strings.TrimFunc(p, func(r rune) bool { return r < '0' || r > '9' }))
		o[i] = n
	}
	return o
}

func checkUpdate() (UpdateInfo, error) {
	u := UpdateInfo{}
	if strings.HasPrefix(Repo, "REPLACE") {
		return u, nil
	}
	req, _ := http.NewRequest("GET", releaseAPI(), nil)
	req.Header.Set("User-Agent", "RedsunLauncher/"+Version)
	req.Header.Set("Accept", "application/vnd.github+json")
	res, err := httpc.Do(req)
	if err != nil {
		return u, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return u, fmt.Errorf("GitHub ตอบ %d", res.StatusCode)
	}
	var rel struct {
		Tag    string `json:"tag_name"`
		Body   string `json:"body"`
		Draft  bool   `json:"draft"`
		Assets []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 4<<20)).Decode(&rel); err != nil {
		return u, err
	}
	for _, a := range rel.Assets {
		switch a.Name {
		case ExeAsset:
			u.exeURL = a.URL
		case ShaAsset:
			u.shaURL = a.URL
		}
	}
	u.Version = strings.TrimPrefix(rel.Tag, "v")
	u.Notes = rel.Body
	u.Available = !rel.Draft && newer(u.Version, Version) && u.exeURL != "" && u.shaURL != ""
	return u, nil
}

// ดาวน์โหลด → ตรวจ SHA-256 → สลับไฟล์ (เปลี่ยนชื่อตัวที่รันอยู่เป็น .old) → เปิดตัวใหม่
func applyUpdate(u UpdateInfo, progress func(float64, string)) error {
	if u.exeURL == "" || u.shaURL == "" {
		return errors.New("release ไม่มีไฟล์ครบ")
	}
	c := &http.Client{Timeout: 10 * time.Minute}
	sr, err := c.Get(u.shaURL)
	if err != nil {
		return err
	}
	sb, _ := io.ReadAll(io.LimitReader(sr.Body, 4096))
	sr.Body.Close()
	want := strings.ToLower(strings.Fields(string(sb) + " x")[0])
	if len(want) != 64 {
		return errors.New("ไฟล์ SHA-256 ไม่ถูกต้อง")
	}
	self, err := os.Executable()
	if err != nil {
		return err
	}
	tmp := self + ".new"
	res, err := c.Get(u.exeURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("ดาวน์โหลดไม่สำเร็จ (%d)", res.StatusCode)
	}
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	h := sha256.New()
	total := res.ContentLength
	var got int64
	buf := make([]byte, 64<<10)
	last := time.Time{}
	for {
		n, rerr := res.Body.Read(buf)
		if n > 0 {
			f.Write(buf[:n])
			h.Write(buf[:n])
			got += int64(n)
			if total > 0 && time.Since(last) > 120*time.Millisecond {
				last = time.Now()
				progress(float64(got)/float64(total), fmt.Sprintf("ดาวน์โหลด %.1f / %.1f MB", float64(got)/1e6, float64(total)/1e6))
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			f.Close()
			os.Remove(tmp)
			return rerr
		}
	}
	f.Close()
	if hex.EncodeToString(h.Sum(nil)) != want {
		os.Remove(tmp)
		return errors.New("ไฟล์ไม่ตรง SHA-256 · ยกเลิกเพื่อความปลอดภัย")
	}
	progress(1, "ติดตั้งเวอร์ชันใหม่…")
	old := self + ".old"
	os.Remove(old)
	if err := os.Rename(self, old); err != nil {
		os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, self); err != nil {
		os.Rename(old, self)
		return err
	}
	cmd := exec.Command(self, "--updated")
	return cmd.Start()
}

// เปิดเวอร์ชันใหม่แล้วลบไฟล์เก่า
func cleanupOld() {
	if self, err := os.Executable(); err == nil {
		for i := 0; i < 10; i++ {
			if err := os.Remove(self + ".old"); err == nil || os.IsNotExist(err) {
				return
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
}
