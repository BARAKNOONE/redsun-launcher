package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ===== ข้อมูลจาก GitHub (remote/launcher.json) · ทีมแก้ไฟล์นี้บน GitHub = ทุกคนเห็นทันที ไม่ต้องออก Launcher ใหม่ =====
type Server struct {
	Name        string `json:"name"`
	Connect     string `json:"connect"` // ip:port หรือ cfx.re/join/xxxx
	Status      string `json:"status"`  // http://ip:30120/dynamic.json
	Max         int    `json:"max"`
	Maintenance bool   `json:"maintenance"` // true = ปิดปุ่มเล่น
}
type IniKV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Label string `json:"label"`
	Short string `json:"short"`
}
type Event struct {
	Kicker   string `json:"kicker"`
	Title    string `json:"title"`
	Desc     string `json:"desc"`
	At       int64  `json:"at"`       // เวลาเริ่ม unix (ใช้นับถอยหลัง)
	WhenText string `json:"whenText"` // ข้อความวันเวลา
	Place    string `json:"place"`
	Image    string `json:"image"` // ลิงก์รูปพื้นหลัง (แนะนำวางใน repo /remote/img)
	URL      string `json:"url"`
}
type News struct {
	Date  string `json:"date"`
	Title string `json:"title"`
	Text  string `json:"text"`
	URL   string `json:"url"`
}
type Link struct {
	T   string `json:"t"`
	URL string `json:"url"`
}
type Remote struct {
	Server      Server  `json:"server"`
	CitizenFX   []IniKV `json:"citizenfx"`
	MinVersion  string  `json:"minVersion"` // ต่ำกว่านี้ = บังคับอัปเดต
	Events      []Event `json:"events"`
	News        []News  `json:"news"`
	Links       []Link  `json:"links"`
	CloseOnPlay bool    `json:"closeOnPlay"`
}

var httpc = &http.Client{Timeout: 8 * time.Second}

func appDir() string {
	d := filepath.Join(os.Getenv("LOCALAPPDATA"), "RedsunLauncher")
	os.MkdirAll(d, 0o755)
	return d
}

// โหลด launcher.json · ไม่ได้ = ใช้ไฟล์แคชล่าสุด · ไม่มีแคช = ค่าเริ่มต้น
func fetchRemote() (Remote, bool) {
	r := defaultRemote()
	if strings.HasPrefix(Repo, "REPLACE") {
		return r, false // ยังไม่ตั้ง GitHub repo = ใช้ค่าที่ฝังมากับ exe
	}
	cache := filepath.Join(appDir(), "launcher.json")
	req, _ := http.NewRequest("GET", remoteURL()+"?t="+strconv.FormatInt(time.Now().Unix()/60, 10), nil)
	req.Header.Set("Cache-Control", "no-cache")
	if res, err := httpc.Do(req); err == nil {
		defer res.Body.Close()
		if res.StatusCode == 200 {
			b, _ := io.ReadAll(io.LimitReader(res.Body, 2<<20))
			var x Remote
			if json.Unmarshal(b, &x) == nil {
				os.WriteFile(cache, b, 0o644)
				return merge(r, x), true
			}
		}
	}
	if b, err := os.ReadFile(cache); err == nil {
		var x Remote
		if json.Unmarshal(b, &x) == nil {
			return merge(r, x), false
		}
	}
	return r, false
}

// ค่าใน launcher.json ทับค่าเริ่มต้น (ที่ไม่ได้ใส่ใช้ค่าเดิม)
func merge(d, x Remote) Remote {
	if x.Server.Name != "" {
		d.Server = x.Server
		if d.Server.Max == 0 {
			d.Server.Max = 2000
		}
	}
	if len(x.CitizenFX) > 0 {
		d.CitizenFX = x.CitizenFX
	}
	d.MinVersion = x.MinVersion
	d.Events, d.News, d.Links = x.Events, x.News, x.Links
	d.CloseOnPlay = x.CloseOnPlay
	return d
}

// ===== สถานะเซิร์ฟ (FXServer: /dynamic.json → clients, sv_maxclients) =====
type Status struct {
	Up      bool `json:"up"`
	Online  int  `json:"online"`
	Max     int  `json:"max"`
	Checked bool `json:"checked"`
}

func fetchStatus(url string) Status {
	st := Status{Checked: true}
	if url == "" {
		return st
	}
	c := &http.Client{Timeout: 4 * time.Second}
	res, err := c.Get(url)
	if err != nil {
		return st
	}
	defer res.Body.Close()
	var x map[string]interface{}
	if json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&x) != nil {
		return st
	}
	st.Up = true
	st.Online = toInt(x["clients"])
	st.Max = toInt(x["sv_maxclients"])
	return st
}
func toInt(v interface{}) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case string:
		n, _ := strconv.Atoi(t)
		return n
	}
	return 0
}
