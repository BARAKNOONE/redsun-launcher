# REDSUN Launcher (.exe)
Launcher ของเซิร์ฟ REDSUN (RedM) · Go + WebView2 · ไฟล์เดียว ~6.6 MB · Windows 10/11 · บอร์ดดีไซน์: `Launcher.v2` บนแคนวาส

## ติดตั้ง (สำหรับผู้เล่น)
- ดับเบิลคลิก `RedsunLauncher.exe` ที่ดาวน์โหลดมา → หน้าติดตั้งพร้อมโลโก้ → กด **ติดตั้ง**
- ติดตั้งที่ `%LOCALAPPDATA%\Programs\RedsunLauncher` (ไม่ต้องใช้สิทธิ์แอดมิน) · สร้างไอคอนบนเดสก์ท็อป + เมนู Start · มีรายการใน "แอปและฟีเจอร์" ถอนได้ตามปกติ
- ไม่อยากติดตั้ง: กด "ไม่ติดตั้ง · ใช้แบบพกพา" หรือเปิดด้วย `--portable`

## แก้หน้า Launcher เอง (ไม่ต้องเขียนโค้ด)
- ใช้หน้า **Launcher Studio** (artifact) กรอกกิจกรรม/รูป/ข่าว/ลิงก์/ปิดปรับปรุง → กดคัดลอก → วางทับ `remote/launcher.json` บน GitHub → Commit · ผู้เล่นเห็นใน 1 นาที
- รูปกิจกรรม: อัปโหลดไว้ใน `remote/img/` ของ repo แล้วพิมพ์ชื่อไฟล์ในช่องรูป

## ทำอะไรได้
1. **บังคับค่า CitizenFX.ini** ทุกครั้งก่อนเข้าเซิร์ฟ (หรือกด "ตั้งค่าใหม่")
   ```
   SavedBuildNumber=1491
   PoolSizesIncrease={"DrawableStore":50000,"InteriorProxy":450}
   ReplaceExecutable=0
   UpdateChannel=specific/feature/redm-physical-player-extension
   ```
   - ไฟล์: `%LOCALAPPDATA%\RedM\RedM.app\CitizenFX.ini` (ไม่มี = ลอง `%LOCALAPPDATA%\RedM\CitizenFX.ini` · ไม่มีทั้งคู่ = สร้างใหม่)
   - แก้เฉพาะคีย์ใน `[Game]` · ค่าอื่นของผู้เล่น (IVPath ฯลฯ) อยู่ครบ · สำรองไฟล์เดิมเป็น `CitizenFX.ini.redsun-backup` ครั้งแรก
   - RedM เปิดค้างอยู่ = ไม่เขียน และบอกให้ปิดก่อน (RedM เขียนไฟล์ทับตอนปิดเกม)
   - ค่าเหล่านี้มาจาก `remote/launcher.json` บน GitHub → เปลี่ยนบิลด์/ช่องอัปเดตได้โดยไม่ต้องออก Launcher ใหม่
2. **โปรโมทกิจกรรม/ข่าว** ก่อนเข้าเซิร์ฟ: ภาพพื้นหลัง · นับถอยหลัง · หมุนทุก 7 วิ · ข่าว · ลิงก์ Discord/Facebook/กฎ (แก้ใน `remote/launcher.json`)
3. **สถานะเซิร์ฟสด** จาก `http://IP:30120/dynamic.json` · `maintenance: true` = ล็อกปุ่มเล่น
4. **เข้าเซิร์ฟปุ่มเดียว**: ตั้งค่า ini → เปิด RedM ผ่าน `redm://connect/<connect>` (ไม่ได้ = `RedM.exe +connect`) → ปิด Launcher เอง (`closeOnPlay`)
5. **อัปเดตตัวเองจาก GitHub Releases**: เปิด Launcher = เช็ค release ล่าสุด → มีเวอร์ชันใหม่เด้งหน้าต่าง → ดาวน์โหลด → **ตรวจ SHA-256** → สลับไฟล์ → เปิดใหม่ · `minVersion` ใน launcher.json = บังคับอัปเดตก่อนเล่น
6. ออฟไลน์ = ใช้ข่าว/ค่าตั้งล่าสุดที่เคยโหลด · เครื่องไม่มี WebView2 = พาไปหน้าดาวน์โหลด

## ตั้งค่าครั้งแรก (Bank / ทีม)
1. สร้าง GitHub repo (public) เช่น `redsun-launcher` แล้วอัปโหลดทั้งโฟลเดอร์นี้
2. แก้ `config.go` → `Repo = "ชื่อบัญชี/redsun-launcher"` (หรือปล่อยไว้ เพราะ workflow ใส่ให้ตอน build อัตโนมัติ)
3. แก้ `remote/launcher.json`:
   - ตั้งไว้แล้ว: `server.connect` = `1.2.210.129:30120` · `server.status` = `http://1.2.210.129:30120/dynamic.json` (ฝังใน exe ด้วย ใช้ได้แม้ยังไม่มี GitHub)
   - `events` (`at` = เวลาเริ่มแบบ unix · `image` = ลิงก์รูป เช่น `https://raw.githubusercontent.com/<repo>/main/remote/img/race.jpg`) · `news` · `links`
4. ออกเวอร์ชัน: `git tag v1.0.0 && git push --tags` → GitHub Actions build `RedsunLauncher.exe` + `.sha256` แล้วสร้าง Release ให้เอง
5. แจกลิงก์ Release ให้ผู้เล่นโหลดครั้งแรก · ครั้งต่อไปอัปเดตเองทั้งหมด

## ออกเวอร์ชันใหม่
- แก้โค้ด/หน้า UI (`ui/index.html`) → commit → `git tag v1.0.1 && git push --tags` → ผู้เล่นเปิด Launcher จะเห็น "มีอัปเดต v1.0.1"
- อยากบังคับ: ใส่ `"minVersion": "1.0.1"` ใน launcher.json

## build เอง (ไม่ใช้ GitHub Actions)
```
go mod tidy
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-H windowsgui -s -w -X main.Version=1.0.0 -X main.Repo=owner/redsun-launcher" -o RedsunLauncher.exe .
```
- ไอคอน/manifest อยู่ใน `rsrc_windows_amd64.syso` (สร้างจาก `res/` ด้วย `rsrc -manifest res/app.manifest -ico res/icon.ico -arch amd64`)
- ทดสอบตรรกะ ini/เวอร์ชัน: `go test .`

## ข้อควรรู้
- ไฟล์ .exe ยังไม่ได้เซ็นดิจิทัล → Windows SmartScreen อาจเตือน "Windows protected your PC" ครั้งแรก (กด More info → Run anyway) · อยากให้ไม่เตือน ต้องซื้อ Code Signing Certificate
- ยังไม่ได้ทดสอบบน Windows จริง: หน้าต่างไม่มีกรอบ/ลากย้าย · เปิด RedM ด้วย redm:// · ตำแหน่ง CitizenFX.ini ของ RedM (เดาตาม FiveM = RedM.app) · อัปเดตตัวเอง
- `UpdateChannel=specific/feature/redm-physical-player-extension` เป็นช่องทดลองของ Cfx — ถ้า Cfx ปิดช่องนี้ ให้เปลี่ยนใน launcher.json ทีเดียวทุกคนได้ค่าใหม่
