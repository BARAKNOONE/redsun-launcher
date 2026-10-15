module redsun-launcher

go 1.24

require github.com/jchv/go-webview2 v0.0.0-20260205173254-56598839c808

require (
	github.com/jchv/go-winloader v0.0.0-20250406163304-c1995be93bd1 // indirect
	golang.org/x/sys v0.30.0
)

// golang.org ถูกบล็อกในเครื่อง build บางที่ · ใช้ mirror บน GitHub แทน (โค้ดเดียวกัน)
replace golang.org/x/sys => github.com/golang/sys v0.30.0
