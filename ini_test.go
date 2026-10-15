//go:build !windows

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ทดสอบการแก้ CitizenFX.ini: เก็บบรรทัดเดิม · แทนค่าเดิม · เพิ่มคีย์ที่ไม่มี · สำรองไฟล์
func TestIni(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("LOCALAPPDATA", dir)
	p := filepath.Join(dir, "RedM", "RedM.app", "CitizenFX.ini")
	os.MkdirAll(filepath.Dir(p), 0o755)
	os.WriteFile(p, []byte("\xEF\xBB\xBF[Game]\r\nIVPath=D:\\RDR2\r\nSavedBuildNumber=1311\r\nupdatechannel=production\r\n\r\n[Other]\r\nX=1\r\n"), 0o644)
	want := defaultRemote().CitizenFX
	if checkIni(want).AllOk {
		t.Fatal("ควรยังไม่ ok")
	}
	if _, err := writeIni(want); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(p)
	s := string(b)
	for _, must := range []string{"IVPath=D:\\RDR2", "SavedBuildNumber=1491", `PoolSizesIncrease={"DrawableStore":50000,"InteriorProxy":450}`, "ReplaceExecutable=0", "UpdateChannel=specific/feature/redm-physical-player-extension", "[Other]\r\nX=1"} {
		if !strings.Contains(s, must) {
			t.Fatalf("ไม่มี %q ใน\n%s", must, s)
		}
	}
	if strings.Count(s, "1311") != 0 || strings.Count(strings.ToLower(s), "updatechannel=") != 1 {
		t.Fatalf("ค่าเก่าไม่หาย\n%s", s)
	}
	gi, oi := strings.Index(s, "UpdateChannel="), strings.Index(s, "[Other]")
	if gi > oi {
		t.Fatal("คีย์ต้องอยู่ใน [Game]")
	}
	if !checkIni(want).AllOk {
		t.Fatal("เขียนแล้วต้อง ok")
	}
	if _, err := os.Stat(p + ".redsun-backup"); err != nil {
		t.Fatal("ไม่มีไฟล์สำรอง")
	}
	// ไม่มีไฟล์เลย → สร้างใหม่
	os.RemoveAll(filepath.Join(dir, "RedM"))
	if _, err := writeIni(want); err != nil || !checkIni(want).AllOk {
		t.Fatal("สร้างไฟล์ใหม่ไม่ได้", err)
	}
	// RedM เปิดอยู่ = ห้ามเขียน
	fakeRunning = true
	if _, err := writeIni(want); err == nil {
		t.Fatal("ต้องห้ามเขียนตอน RedM เปิด")
	}
	fakeRunning = false
	t.Log(s)
}

func TestVersion(t *testing.T) {
	if !newer("1.0.10", "1.0.9") || newer("1.0.0", "1.0.0") || !newer("v2.0", "1.9.9") || newer("1.0.0", "1.0.1") {
		t.Fatal("เทียบเวอร์ชันผิด")
	}
}
