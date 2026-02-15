//go:build windows

package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"unsafe"

	"syscall"

	ico "github.com/Kodeworks/golang-image-ico"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"golang.org/x/sys/windows"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: taskbar-badge <pid> <count>")
		os.Exit(2)
	}
	pid, _ := strconv.Atoi(os.Args[1])
	count, _ := strconv.Atoi(os.Args[2])

	hwnd := findWindowForPID(uint32(pid))
	if hwnd == 0 {
		fmt.Fprintln(os.Stderr, "no window for pid")
		os.Exit(1)
	}

	if count <= 0 {
		// clear overlay
		setOverlayIcon(hwnd, 0, "")
		return
	}

	tmp := filepath.Join(os.TempDir(), fmt.Sprintf("versadumps_badge_%d.ico", count))
	if err := createBadgeICO(tmp, count); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer os.Remove(tmp)

	hicon := loadIconFromFile(tmp)
	if hicon == 0 {
		fmt.Fprintln(os.Stderr, "failed create icon")
		os.Exit(1)
	}

	setOverlayIcon(hwnd, hicon, "")
}

func createBadgeICO(path string, count int) error {
	const size = 64
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.Transparent}, image.Point{}, draw.Src)
	cx, cy := size/2, size/2
	r := size/2 - 2
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := x - cx
			dy := y - cy
			if dx*dx+dy*dy <= r*r {
				img.Set(x, y, color.RGBA{R: 220, G: 38, B: 38, A: 255})
			}
		}
	}
	// render static label 'VD' instead of number
	label := "VD"
	face := basicfont.Face7x13
	d := &fontDrawer{Dst: img, Src: image.NewUniform(color.White), Face: face}
	txtW := len(label) * 7
	x := (size - txtW) / 2
	y := (size + 7) / 2
	d.DrawStringAt(label, x, y)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := ico.Encode(f, img); err != nil {
		f.Seek(0, 0)
		png.Encode(f, img)
		return err
	}
	return nil
}

// minimal text drawer using basicfont
type fontDrawer struct {
	Dst  draw.Image
	Src  image.Image
	Face *basicfont.Face
}

func (fd *fontDrawer) DrawStringAt(s string, x, y int) {
	// simple naive draw
	pt := fixed.P(x, y)
	d := &font.Drawer{Dst: fd.Dst, Src: fd.Src, Face: fd.Face, Dot: pt}
	d.DrawString(s)
}

func loadIconFromFile(path string) windows.Handle {
	p, _ := syscall.UTF16PtrFromString(path)
	user32 := syscall.NewLazyDLL("user32.dll")
	proc := user32.NewProc("LoadImageW")
	const LR_LOADFROMFILE = 0x00000010
	ret, _, _ := proc.Call(0, uintptr(unsafe.Pointer(p)), uintptr(1), 0, 0, uintptr(LR_LOADFROMFILE))
	return windows.Handle(ret)
}

func findWindowForPID(pid uint32) windows.HWND {
	var found windows.HWND
	cb := syscall.NewCallback(func(hwnd uintptr, lparam uintptr) uintptr {
		var procId uint32
		windows.GetWindowThreadProcessId(windows.HWND(hwnd), &procId)
		if procId == pid {
			if windows.IsWindowVisible(windows.HWND(hwnd)) {
				found = windows.HWND(hwnd)
				return 0
			}
		}
		return 1
	})
	windows.EnumWindows(cb, unsafe.Pointer(nil))
	return found
}

func setOverlayIcon(hwnd windows.HWND, hicon windows.Handle, desc string) {
	ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)
	defer ole.CoUninitialize()
	unknown, err := oleutil.CreateObject("Shell.TaskbarList")
	if err != nil {
		return
	}
	defer unknown.Release()
	tb, err := unknown.QueryInterface(ole.NewGUID("{56FDF344-FD6D-11D0-958A-006097C9A090}"))
	if err != nil {
		return
	}
	defer tb.Release()
	oleutil.CallMethod(tb, "HrInit")
	oleutil.CallMethod(tb, "SetOverlayIcon", uintptr(hwnd), uintptr(hicon), desc)
}
