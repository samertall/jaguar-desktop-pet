//go:build windows

// Jaguar Desktop Pet Store Edition 0.1: Windows x64 transparent desktop pet.
// Builds with Go stdlib (CGO_ENABLED=0) and embeds all frames; no installer or JS runtime.
package main

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"fmt"
	"image"
	"image/draw"
	_ "image/png"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

//go:embed assets/walk_*.png
var assetFS embed.FS

var (
	user32                 = syscall.NewLazyDLL("user32.dll")
	gdi32                  = syscall.NewLazyDLL("gdi32.dll")
	shell32                = syscall.NewLazyDLL("shell32.dll")
	kernel32               = syscall.NewLazyDLL("kernel32.dll")
	pRegisterClass         = user32.NewProc("RegisterClassExW")
	pCreateWindow          = user32.NewProc("CreateWindowExW")
	pDefWindowProc         = user32.NewProc("DefWindowProcW")
	pShowWindow            = user32.NewProc("ShowWindow")
	pDestroyWindow         = user32.NewProc("DestroyWindow")
	pPostQuitMessage       = user32.NewProc("PostQuitMessage")
	pGetMessage            = user32.NewProc("GetMessageW")
	pTranslateMessage      = user32.NewProc("TranslateMessage")
	pDispatchMessage       = user32.NewProc("DispatchMessageW")
	pGetDC                 = user32.NewProc("GetDC")
	pReleaseDC             = user32.NewProc("ReleaseDC")
	pSetTimer              = user32.NewProc("SetTimer")
	pKillTimer             = user32.NewProc("KillTimer")
	pUpdateLayer           = user32.NewProc("UpdateLayeredWindow")
	pSPI                   = user32.NewProc("SystemParametersInfoW")
	pSetDPIAware           = user32.NewProc("SetProcessDPIAware")
	pLoadIcon              = user32.NewProc("LoadIconW")
	pMessageBox            = user32.NewProc("MessageBoxW")
	pRegisterHotkey        = user32.NewProc("RegisterHotKey")
	pUnregisterHotkey      = user32.NewProc("UnregisterHotKey")
	pRegisterWindowMessage = user32.NewProc("RegisterWindowMessageW")
	pCreatePopupMenu       = user32.NewProc("CreatePopupMenu")
	pAppendMenu            = user32.NewProc("AppendMenuW")
	pTrackPopupMenu        = user32.NewProc("TrackPopupMenu")
	pDestroyMenu           = user32.NewProc("DestroyMenu")
	pGetCursorPos          = user32.NewProc("GetCursorPos")
	pSetForeground         = user32.NewProc("SetForegroundWindow")
	pIsWindowVisible       = user32.NewProc("IsWindowVisible")
	pGetWindowRect         = user32.NewProc("GetWindowRect")
	pPostMessage           = user32.NewProc("PostMessageW")
	pCreateDC              = gdi32.NewProc("CreateCompatibleDC")
	pDeleteDC              = gdi32.NewProc("DeleteDC")
	pCreateDIB             = gdi32.NewProc("CreateDIBSection")
	pSelectObject          = gdi32.NewProc("SelectObject")
	pDeleteObject          = gdi32.NewProc("DeleteObject")
	pNotifyIcon            = shell32.NewProc("Shell_NotifyIconW")
	pGetModuleHandle       = kernel32.NewProc("GetModuleHandleW")
)

const (
	wsExLayered      = 0x80000
	wsExTransparent  = 0x20
	wsExToolwindow   = 0x80
	wsExTopmost      = 0x8
	wsExNoActivate   = 0x08000000
	wsPopup          = 0x80000000
	swShowNoActivate = 4
	wmDestroy        = 0x0002
	wmTimer          = 0x0113
	wmHotkey         = 0x0312
	wmDisplayChange  = 0x007E
	wmSettingChange  = 0x001A
	wmRButtonUp      = 0x0205
	wmLButtonDblclk  = 0x0203
	wmNull           = 0x0000
	wmTray           = 0x8000 + 9
	nimAdd           = 0x0
	nimDelete        = 0x2
	nifMessage       = 0x1
	nifIcon          = 0x2
	nifTip           = 0x4
	mfString         = 0x0
	mfSeparator      = 0x800
	tpmReturnCmd     = 0x100
	tpmRightButton   = 0x2
	ulwAlpha         = 0x2
	spiGetWorkArea   = 0x30
	modAlt           = 0x0001
	modControl       = 0x0002
	vkJ              = 0x4A
	vkQ              = 0x51
	vkT              = 0x54
)

type point struct{ X, Y int32 }
type rect struct{ Left, Top, Right, Bottom int32 }
type size struct{ CX, CY int32 }
type blendFunction struct{ Op, Flags, SourceAlpha, AlphaFormat byte }
type wndClassEx struct {
	CbSize                             uint32
	Style                              uint32
	WndProc                            uintptr
	ClsExtra                           int32
	WndExtra                           int32
	Instance, Icon, Cursor, Background uintptr
	MenuName, ClassName                *uint16
	IconSmall                          uintptr
}
type msgType struct {
	Hwnd           uintptr
	Message        uint32
	_              uint32
	WParam, LParam uintptr
	Time           uint32
	Pt             point
	LPrivate       uint32
	_2             uint32
}
type bitmapInfoHeader struct {
	Size                   uint32
	Width, Height          int32
	Planes, BitCount       uint16
	Compression, SizeImage uint32
	XPels, YPels           int32
	ClrUsed, ClrImportant  uint32
}
type bitmapInfo struct {
	Header bitmapInfoHeader
	Colors [1]uint32
}
type notifyIconData struct {
	Size                uint32
	Hwnd                uintptr
	ID, Flags, Callback uint32
	Icon                uintptr
	Tip                 [128]uint16
	State, StateMask    uint32
	Info                [256]uint16
	Version             uint32
	InfoTitle           [64]uint16
	InfoFlags           uint32
	Guid                [16]byte
	BalloonIcon         uintptr
}

var (
	app     *application
	logFile *os.File
	logPath string
)

func logMessage(format string, args ...any) {
	if logFile != nil {
		fmt.Fprintf(logFile, "%s  %s\r\n", time.Now().Format("2006-01-02 15:04:05.000"), fmt.Sprintf(format, args...))
		_ = logFile.Sync()
	}
}
func utf16(s string) *uint16 {
	p, e := syscall.UTF16PtrFromString(s)
	if e != nil {
		return nil
	}
	return p
}
func msgbox(title, message string) {
	pMessageBox.Call(0, uintptr(unsafe.Pointer(utf16(message))), uintptr(unsafe.Pointer(utf16(title))), 0x10)
}
func syscallErr(proc *syscall.LazyProc, r uintptr, e error) error {
	if r != 0 {
		return nil
	}
	if e != nil && e != syscall.Errno(0) {
		return fmt.Errorf("%s: %w", proc.Name, e)
	}
	return fmt.Errorf("%s failed with zero return", proc.Name)
}

func startLog() {
	dir := filepath.Join(os.Getenv("LOCALAPPDATA"), "JaguarDesktopPet")
	if os.Getenv("LOCALAPPDATA") == "" {
		dir = filepath.Join(os.TempDir(), "JaguarDesktopPet")
	}
	_ = os.MkdirAll(dir, 0755)
	logPath = filepath.Join(dir, "jaguar.log")
	logFile, _ = os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	logMessage("--- START Store Edition 0.1 --- Windows AMD64, PID=%d, EXE=%s", os.Getpid(), os.Args[0])
}

type application struct {
	hwnd, screenDC, memoryDC, hbitmap, oldBitmap uintptr
	pixelPtr                                     unsafe.Pointer
	dibSize                                      int
	imgW, imgH                                   int
	frames                                       [8]image.Image
	cached                                       [2][8][]byte
	orientation                                  int // 0=left,1=right
	w                                            Walker
	work                                         rect
	rng                                          *rand.Rand
	lastTick                                     time.Time
	prevState                                    Behavior
	ticks                                        int
	tray                                         bool
	trayInfo                                     notifyIconData
	trayAdded                                    bool
	taskbarMsg                                   uint32
	failures                                     int
	guiSmoke                                     bool
	smokeInitialX                                float64
	smokeInitialFrame                            int
	smokeFrameChanged                            bool
	smokePassed                                  bool
}

func (a *application) getWorkArea() {
	r := rect{}
	ok, _, e := pSPI.Call(spiGetWorkArea, 0, uintptr(unsafe.Pointer(&r)), 0)
	if ok == 0 || r.Right-r.Left < 100 || r.Bottom-r.Top < 100 {
		logMessage("SystemParametersInfoW: %v; using 1920x1080 safe fallback", e)
		r = rect{0, 0, 1920, 1040}
	}
	a.work = r
}

func (a *application) loadFrames() error {
	for i := 0; i < 8; i++ {
		file := fmt.Sprintf("assets/walk_%02d.png", i)
		b, e := assetFS.ReadFile(file)
		if e != nil {
			return fmt.Errorf("embedded %s: %w", file, e)
		}
		img, _, e := image.Decode(bytes.NewReader(b))
		if e != nil {
			return fmt.Errorf("decode %s: %w", file, e)
		}
		a.frames[i] = img
	}
	return nil
}

func decodeToRGBA(img image.Image) *image.NRGBA {
	bounds := img.Bounds()
	out := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(out, out.Bounds(), img, bounds.Min, draw.Src)
	return out
}

// Resize once on launch/size change; generate premultiplied BGRA pixels for UpdateLayeredWindow.
func (a *application) createBuffers() error {
	a.cached = [2][8][]byte{}
	for i := 0; i < 8; i++ {
		src := decodeToRGBA(a.frames[i])
		sw := src.Bounds().Dx()
		sh := src.Bounds().Dy()
		if sw < 2 || sh < 2 {
			return fmt.Errorf("invalid source dimensions %d x %d", sw, sh)
		}
		rendered := make([]byte, a.imgW*a.imgH*4)
		for y := 0; y < a.imgH; y++ {
			sy := (float64(y)+.5)*float64(sh)/float64(a.imgH) - .5
			y0 := int(math.Floor(sy))
			ty := sy - float64(y0)
			if y0 < 0 {
				y0 = 0
				ty = 0
			}
			if y0 >= sh-1 {
				y0 = sh - 1
				ty = 0
			}
			y1 := y0 + 1
			if y1 >= sh {
				y1 = sh - 1
			}
			for x := 0; x < a.imgW; x++ {
				sx := (float64(x)+.5)*float64(sw)/float64(a.imgW) - .5
				x0 := int(math.Floor(sx))
				tx := sx - float64(x0)
				if x0 < 0 {
					x0 = 0
					tx = 0
				}
				if x0 >= sw-1 {
					x0 = sw - 1
					tx = 0
				}
				x1 := x0 + 1
				if x1 >= sw {
					x1 = sw - 1
				}
				pos := (y*a.imgW + x) * 4
				ids := [4]int{y0*src.Stride + x0*4, y0*src.Stride + x1*4, y1*src.Stride + x0*4, y1*src.Stride + x1*4}
				weights := [4]float64{(1 - tx) * (1 - ty), tx * (1 - ty), (1 - tx) * ty, tx * ty}
				var premul [4]float64 // R,G,B,Alpha
				for p, offset := range ids {
					aa := float64(src.Pix[offset+3]) / 255.0
					ww := weights[p]
					premul[0] += float64(src.Pix[offset]) * aa * ww
					premul[1] += float64(src.Pix[offset+1]) * aa * ww
					premul[2] += float64(src.Pix[offset+2]) * aa * ww
					premul[3] += float64(src.Pix[offset+3]) * ww
				}
				for c := 0; c < 4; c++ {
					if premul[c] < 0 {
						premul[c] = 0
					}
					if premul[c] > 255 {
						premul[c] = 255
					}
				}
				rendered[pos+0] = byte(premul[2] + .5)
				rendered[pos+1] = byte(premul[1] + .5)
				rendered[pos+2] = byte(premul[0] + .5)
				rendered[pos+3] = byte(premul[3] + .5)
			}
		}
		a.cached[0][i] = rendered
		flipped := make([]byte, len(rendered))
		for y := 0; y < a.imgH; y++ {
			for x := 0; x < a.imgW; x++ {
				dst := (y*a.imgW + x) * 4
				srcOff := (y*a.imgW + (a.imgW - 1 - x)) * 4
				copy(flipped[dst:dst+4], rendered[srcOff:srcOff+4])
			}
		}
		a.cached[1][i] = flipped
	}
	logMessage("Raster cache prepared %dx%d: 8 frames x 2 orientations", a.imgW, a.imgH)
	return nil
}

func (a *application) freeDIB() {
	if a.memoryDC != 0 {
		if a.oldBitmap != 0 {
			pSelectObject.Call(a.memoryDC, a.oldBitmap)
			a.oldBitmap = 0
		}
		if a.hbitmap != 0 {
			pDeleteObject.Call(a.hbitmap)
			a.hbitmap = 0
		}
		pDeleteDC.Call(a.memoryDC)
		a.memoryDC = 0
	}
	a.pixelPtr = nil
	a.dibSize = 0
}
func (a *application) makeDIB() error {
	a.freeDIB()
	if a.screenDC == 0 {
		dc, _, err := pGetDC.Call(0)
		if dc == 0 {
			return fmt.Errorf("GetDC: %w", err)
		}
		a.screenDC = dc
	}
	dc, _, e := pCreateDC.Call(a.screenDC)
	if dc == 0 {
		return fmt.Errorf("CreateCompatibleDC: %w", e)
	}
	a.memoryDC = dc
	info := bitmapInfo{Header: bitmapInfoHeader{Size: uint32(unsafe.Sizeof(bitmapInfoHeader{})), Width: int32(a.imgW), Height: -int32(a.imgH), Planes: 1, BitCount: 32, Compression: 0}}
	var pixels unsafe.Pointer
	bmp, _, e := pCreateDIB.Call(a.screenDC, uintptr(unsafe.Pointer(&info)), 0, uintptr(unsafe.Pointer(&pixels)), 0, 0)
	if bmp == 0 || pixels == nil {
		a.freeDIB()
		return fmt.Errorf("CreateDIBSection: %w", e)
	}
	a.hbitmap = bmp
	a.pixelPtr = pixels
	old, _, _ := pSelectObject.Call(a.memoryDC, bmp)
	a.oldBitmap = old
	a.dibSize = a.imgW * a.imgH * 4
	return nil
}
func (a *application) render(yOffset int) error {
	if a.pixelPtr == nil || a.memoryDC == 0 {
		return fmt.Errorf("DIB not ready")
	}
	orientation := 0
	if a.w.Direction > 0 {
		orientation = 1
	}
	idx := a.w.FrameIndex
	if idx < 0 || idx > 7 {
		idx = 0
	}
	bytesBuf := a.cached[orientation][idx]
	if len(bytesBuf) != a.dibSize {
		return fmt.Errorf("invalid frame cache")
	}
	dest := unsafe.Slice((*byte)(a.pixelPtr), a.dibSize)
	copy(dest, bytesBuf)
	location := point{int32(math.Round(a.w.X)), a.work.Bottom - int32(a.imgH) - 2 + int32(yOffset)}
	layerSize := size{int32(a.imgW), int32(a.imgH)}
	origin := point{0, 0}
	blend := blendFunction{Op: 0, Flags: 0, SourceAlpha: 255, AlphaFormat: 1}
	r, _, err := pUpdateLayer.Call(a.hwnd, a.screenDC, uintptr(unsafe.Pointer(&location)), uintptr(unsafe.Pointer(&layerSize)), a.memoryDC, uintptr(unsafe.Pointer(&origin)), 0, uintptr(unsafe.Pointer(&blend)), ulwAlpha)
	if r == 0 {
		return fmt.Errorf("UpdateLayeredWindow: %w", err)
	}
	return nil
}
func (a *application) addTray() {
	icon, _, _ := pLoadIcon.Call(0, 32512) // standard application icon, no external assets
	a.trayInfo = notifyIconData{Hwnd: a.hwnd, ID: 1, Flags: nifMessage | nifIcon | nifTip, Callback: wmTray, Icon: icon}
	a.trayInfo.Size = uint32(unsafe.Sizeof(a.trayInfo))
	u := syscall.StringToUTF16("Jaguar Desktop Pet - right-click for controls")
	copy(a.trayInfo.Tip[:], u)
	r, _, e := pNotifyIcon.Call(nimAdd, uintptr(unsafe.Pointer(&a.trayInfo)))
	a.trayAdded = r != 0
	logMessage("System tray registration: ok=%v err=%v", a.trayAdded, e)
}
func (a *application) removeTray() {
	if a.trayAdded {
		pNotifyIcon.Call(nimDelete, uintptr(unsafe.Pointer(&a.trayInfo)))
		a.trayAdded = false
	}
}
func (a *application) contextMenu() {
	menu, _, _ := pCreatePopupMenu.Call()
	if menu == 0 {
		return
	}
	defer pDestroyMenu.Call(menu)
	item := func(id uintptr, title string) {
		pAppendMenu.Call(menu, mfString, id, uintptr(unsafe.Pointer(utf16(title))))
	}
	sep := func() { pAppendMenu.Call(menu, mfSeparator, 0, 0) }
	if a.w.Paused {
		item(1, "Resume animation")
	} else {
		item(1, "Pause animation")
	}
	item(2, "Auto roaming")
	sep()
	item(3, "Walk")
	item(4, "Run")
	item(5, "Stalk")
	item(6, "Rest")
	item(7, "Pounce")
	item(8, "Turn around")
	sep()
	item(9, "Slow speed")
	item(10, "Normal speed")
	item(11, "Fast speed")
	sep()
	item(12, "Small size")
	item(13, "Normal size")
	item(14, "Large size")
	sep()
	item(15, "About and keyboard shortcuts")
	item(16, "Exit Jaguar")
	cursor := point{}
	pGetCursorPos.Call(uintptr(unsafe.Pointer(&cursor)))
	pSetForeground.Call(a.hwnd)
	cmd, _, _ := pTrackPopupMenu.Call(menu, tpmReturnCmd|tpmRightButton, uintptr(cursor.X), uintptr(cursor.Y), 0, a.hwnd, 0)
	a.command(cmd)
	pPostMessage.Call(a.hwnd, wmNull, 0, 0)
}
func (a *application) resize(scale float64) {
	ow, oh := a.imgW, a.imgH
	a.imgW = int(426 * scale)
	a.imgH = int(260 * scale)
	if e := a.createBuffers(); e != nil {
		logMessage("RESIZE FAILED cache: %s", e)
		a.imgW = ow
		a.imgH = oh
		_ = a.createBuffers()
		return
	}
	if e := a.makeDIB(); e != nil {
		logMessage("RESIZE FAILED DIB: %s", e)
		return
	}
	a.getWorkArea()
	a.w.X = math.Max(float64(a.work.Left), math.Min(a.w.X, float64(a.work.Right-int32(a.imgW))))
	_ = a.render(0)
	logMessage("RESIZED to %dx%d", a.imgW, a.imgH)
}
func (a *application) command(cmd uintptr) {
	switch cmd {
	case 1:
		a.w.Paused = !a.w.Paused
	case 2:
		a.w.Automatic = true
		a.w.State = Walk
		a.w.Age = 0
		a.w.Duration = 7
		a.w.Paused = false
	case 3:
		a.w.SetState(Walk)
	case 4:
		a.w.SetState(Run)
	case 5:
		a.w.SetState(Stalk)
	case 6:
		a.w.SetState(Rest)
	case 7:
		a.w.SetState(Pounce)
	case 8:
		a.w.Direction *= -1
	case 9:
		a.w.SpeedMultiplier = .65
	case 10:
		a.w.SpeedMultiplier = 1
	case 11:
		a.w.SpeedMultiplier = 1.6
	case 12:
		a.resize(.70)
	case 13:
		a.resize(1)
	case 14:
		a.resize(1.35)
	case 15:
		msgbox("Jaguar Desktop Pet", "Jaguar Desktop Pet 0.1 - Store Edition\n\nRight-click the tray icon for controls.\nCtrl+Alt+J: pause / resume\nCtrl+Alt+T: turn around\nCtrl+Alt+Q: exit\n\nLog: "+logPath+"\n\nUses rendered animation frames; not a rigged 3D character.")
	case 16:
		pDestroyWindow.Call(a.hwnd)
	}
	if cmd != 0 {
		logMessage("Menu command=%d state=%v paused=%v", cmd, a.w.State, a.w.Paused)
	}
}
func (a *application) tick() {
	now := time.Now()
	dt := now.Sub(a.lastTick).Seconds()
	a.lastTick = now
	left := float64(a.work.Left) - 10
	right := float64(a.work.Right-int32(a.imgW)) + 10
	prevState := a.w.State
	yOffset := a.w.Update(dt, left, right, a.rng)
	if a.w.State != prevState {
		logMessage("Behavior changed to %s", a.w.State.String())
	}
	if e := a.render(yOffset); e != nil {
		a.failures++
		logMessage("RENDER ERROR %d: %s", a.failures, e)
		if a.failures == 1 {
			msgbox("Jaguar rendering error", "The jaguar renderer could not start.\n\n"+e.Error()+"\n\nLog: "+logPath)
		}
		if a.failures > 2 {
			pDestroyWindow.Call(a.hwnd)
		}
		return
	}
	a.ticks++
	if a.guiSmoke {
		if a.w.FrameIndex != a.smokeInitialFrame {
			a.smokeFrameChanged = true
		}
		if a.ticks >= 65 {
			visible, _, _ := pIsWindowVisible.Call(a.hwnd)
			bounds := rect{}
			gotBounds, _, _ := pGetWindowRect.Call(a.hwnd, uintptr(unsafe.Pointer(&bounds)))
			moved := math.Abs(a.w.X-a.smokeInitialX) > 8
			validBounds := gotBounds != 0 && bounds.Right > bounds.Left && bounds.Bottom > bounds.Top
			a.smokePassed = a.failures == 0 && moved && a.smokeFrameChanged && visible != 0 && validBounds
			result := fmt.Sprintf("GUI smoke: passed=%v ticks=%d moved=%v frames_changed=%v visible=%v bounds_valid=%v rendered_failures=%d start_x=%.1f end_x=%.1f\\n", a.smokePassed, a.ticks, moved, a.smokeFrameChanged, visible != 0, validBounds, a.failures, a.smokeInitialX, a.w.X)
			report := os.Getenv("JAGUAR_GUI_SMOKE_REPORT")
			if report == "" { report = filepath.Join(os.TempDir(), "jaguar-gui-smoke.txt") }
			_ = os.WriteFile(report, []byte(result), 0644)
			logMessage("%s", strings.TrimSpace(result))
			pDestroyWindow.Call(a.hwnd)
			return
		}
	}
	if a.ticks == 1 || a.ticks == 30 || a.ticks == 150 {
		logMessage("ANIMATING tick=%d x=%.0f frame=%d behavior=%s direction=%v", a.ticks, a.w.X, a.w.FrameIndex, a.w.State.String(), a.w.Direction)
	}
}
func windowProc(hwnd uintptr, message uint32, wparam, lparam uintptr) uintptr {
	if app == nil {
		return defaultProc(hwnd, message, wparam, lparam)
	}
	switch message {
	case wmTimer:
		if wparam == 1 {
			app.tick()
			return 0
		}
	case wmTray:
		switch uint32(lparam & 0xffffffff) {
		case wmRButtonUp:
			app.contextMenu()
			return 0
		case wmLButtonDblclk:
			app.command(1)
			return 0
		}
	case wmHotkey:
		switch wparam {
		case 100:
			app.command(1)
		case 101:
			app.command(16)
		case 102:
			app.command(8)
		}
		return 0
	case wmDisplayChange, wmSettingChange:
		app.getWorkArea()
		return 0
	case wmDestroy:
		logMessage("CLOSING window")
		pKillTimer.Call(hwnd, 1)
		for _, id := range []uintptr{100, 101, 102} {
			pUnregisterHotkey.Call(hwnd, id)
		}
		app.removeTray()
		app.freeDIB()
		if app.screenDC != 0 {
			pReleaseDC.Call(0, app.screenDC)
			app.screenDC = 0
		}
		pPostQuitMessage.Call(0)
		return 0
	}
	if app.taskbarMsg != 0 && message == app.taskbarMsg {
		app.addTray()
		return 0
	}
	return defaultProc(hwnd, message, wparam, lparam)
}
func defaultProc(hwnd uintptr, message uint32, wparam, lparam uintptr) uintptr {
	r, _, _ := pDefWindowProc.Call(hwnd, uintptr(message), wparam, lparam)
	return r
}

func run() error {
	runtime.LockOSThread()
	pSetDPIAware.Call()
	app = &application{imgW: 426, imgH: 260, rng: rand.New(rand.NewSource(time.Now().UnixNano()))}
	app.guiSmoke = len(os.Args) > 1 && strings.EqualFold(os.Args[1], "--ci-gui-smoke")
	if e := app.loadFrames(); e != nil {
		return e
	}
	if e := app.createBuffers(); e != nil {
		return e
	}
	app.getWorkArea()
	startX := math.Max(float64(app.work.Left), float64(app.work.Left)+65)
	app.w = NewWalker(startX)
	app.smokeInitialX = app.w.X
	app.smokeInitialFrame = app.w.FrameIndex
	hInstance, _, _ := pGetModuleHandle.Call(0)
	className := utf16("JaguarDesktopPetOverlayV4")
	wc := wndClassEx{Style: 0, WndProc: syscall.NewCallback(windowProc), Instance: hInstance, ClassName: className}
	wc.CbSize = uint32(unsafe.Sizeof(wc))
	registered, _, e := pRegisterClass.Call(uintptr(unsafe.Pointer(&wc)))
	if registered == 0 {
		return fmt.Errorf("RegisterClassExW: %w", e)
	}
	hwnd, _, e := pCreateWindow.Call(wsExLayered|wsExTransparent|wsExToolwindow|wsExTopmost|wsExNoActivate, uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(utf16("Jaguar Desktop Pet"))), wsPopup, uintptr(int32(app.w.X)), uintptr(app.work.Bottom-int32(app.imgH)), uintptr(app.imgW), uintptr(app.imgH), 0, 0, hInstance, 0)
	if hwnd == 0 {
		return fmt.Errorf("CreateWindowExW: %w", e)
	}
	app.hwnd = hwnd
	logMessage("Window created HWND=0x%x; working area %d,%d..%d,%d", hwnd, app.work.Left, app.work.Top, app.work.Right, app.work.Bottom)
	if e := app.makeDIB(); e != nil {
		return e
	}
	if e := app.render(0); e != nil {
		return e
	}
	app.addTray()
	taskbarMsg, _, _ := pRegisterWindowMessage.Call(uintptr(unsafe.Pointer(utf16("TaskbarCreated"))))
	app.taskbarMsg = uint32(taskbarMsg)
	pRegisterHotkey.Call(hwnd, 100, modAlt|modControl, vkJ)
	pRegisterHotkey.Call(hwnd, 101, modAlt|modControl, vkQ)
	pRegisterHotkey.Call(hwnd, 102, modAlt|modControl, vkT)
	pShowWindow.Call(hwnd, swShowNoActivate)
	timer, _, e := pSetTimer.Call(hwnd, 1, 40, 0)
	if timer == 0 {
		return fmt.Errorf("SetTimer: %w", e)
	}
	app.lastTick = time.Now()
	logMessage("First render OK; 40ms timer started; tray=%v", app.trayAdded)
	var message msgType
	for {
		ret, _, err := pGetMessage.Call(uintptr(unsafe.Pointer(&message)), 0, 0, 0)
		if int32(ret) == -1 {
			return fmt.Errorf("GetMessageW: %w", err)
		}
		if ret == 0 {
			break
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&message)))
		pDispatchMessage.Call(uintptr(unsafe.Pointer(&message)))
	}
	logMessage("Exited message loop normally")
	return nil
}
// ciSelfTest tests embedded assets and animation logic without opening Windows UI.
// It does not constitute a visual rendering, malware or Store certification test.
func ciSelfTest() error {
    a := &application{imgW: 426, imgH: 260}
    if err := a.loadFrames(); err != nil { return err }
    hashes := make(map[[32]byte]bool)
    for i := 0; i < 8; i++ {
        filename := fmt.Sprintf("assets/walk_%02d.png", i)
        b, err := assetFS.ReadFile(filename)
        if err != nil { return err }
        hashes[sha256.Sum256(b)] = true
        if a.frames[i].Bounds().Dx() <= 0 || a.frames[i].Bounds().Dy() <= 0 { return fmt.Errorf("bad frame %d", i) }
    }
    if len(hashes) != 8 { return fmt.Errorf("expected 8 distinct frames, got %d", len(hashes)) }
    if err := a.createBuffers(); err != nil { return err }
    for _, direction := range []int{0,1} {
        for _, frame := range a.cached[direction] {
            if len(frame) != a.imgW*a.imgH*4 { return fmt.Errorf("missing rendered buffer") }
        }
    }
    walker := NewWalker(100)
    walker.Direction = 1
    random := rand.New(rand.NewSource(1))
    initialX := walker.X
    changed := false
    for i := 0; i < 100; i++ {
        walker.Update(0.04, 0, 1920, random)
        if walker.FrameIndex != 0 { changed = true }
    }
    if walker.X <= initialX || !changed { return fmt.Errorf("walker did not advance or animate") }
    fmt.Printf("CI SELF TEST PASSED: 8 unique PNG frames, alpha buffers, motion x=%.2f, frame=%d\n", walker.X, walker.FrameIndex)
    return nil
}

func main() {
	startLog()
	defer func() {
		if p := recover(); p != nil {
			logMessage("PANIC: %v", p)
			msgbox("Jaguar Desktop Pet - crash", fmt.Sprintf("The app encountered an unexpected error: %v\n\nLog: %s", p, logPath))
		}
		if logFile != nil {
			logFile.Close()
		}
	}()
	if len(os.Args) > 1 && strings.EqualFold(os.Args[1], "--ci-selftest") {
        if e := ciSelfTest(); e != nil { fmt.Fprintln(os.Stderr, "CI SELF TEST FAILED:", e); os.Exit(2) }
        return
    }
	if len(os.Args) > 1 && strings.EqualFold(os.Args[1], "--diagnose") {
		logMessage("Diagnostic mode. OS=%s Arch=%s, embedded frames=8", runtime.GOOS, runtime.GOARCH)
		msgbox("Jaguar Desktop Pet - diagnostic", "Jaguar executable started successfully.\nLog: "+logPath+"\n\nNow launch the application without --diagnose.")
		return
	}
	logMessage("Store Edition native desktop build 0.1")
	guiSmoke := len(os.Args) > 1 && strings.EqualFold(os.Args[1], "--ci-gui-smoke")
	if e := run(); e != nil {
		logMessage("STARTUP ERROR: %s", e)
		if guiSmoke {
			report := os.Getenv("JAGUAR_GUI_SMOKE_REPORT")
			if report != "" { _ = os.WriteFile(report, []byte("GUI smoke startup failed: "+e.Error()+"\n"), 0644) }
			os.Exit(2)
		}
		msgbox("Jaguar Desktop Pet - startup error", "Jaguar could not start:\n\n"+e.Error()+"\n\nLog file: "+logPath)
	}
	if guiSmoke && (app == nil || !app.smokePassed) { os.Exit(3) }
}
