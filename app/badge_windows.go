//go:build windows

package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	ico "github.com/Kodeworks/golang-image-ico"
	"golang.org/x/sys/windows"
)

var (
	ole32                = syscall.NewLazyDLL("ole32.dll")
	user32               = syscall.NewLazyDLL("user32.dll")
	procCoInitializeEx   = ole32.NewProc("CoInitializeEx")
	procCoUninitialize   = ole32.NewProc("CoUninitialize")
	procCoCreateInstance = ole32.NewProc("CoCreateInstance")
)

const (
	COINIT_APARTMENTTHREADED = 0x2
	S_OK                     = 0
	IMAGE_ICON               = 1
	LR_LOADFROMFILE          = 0x00000010
)

// ITaskbarList3 interface GUIDs
var (
	CLSID_TaskbarList = windows.GUID{
		Data1: 0x56FDF344,
		Data2: 0xFD6D,
		Data3: 0x11D0,
		Data4: [8]byte{0x95, 0x8A, 0x00, 0x60, 0x97, 0xC9, 0xA0, 0x90},
	}
	IID_ITaskbarList3 = windows.GUID{
		Data1: 0xEA1AFB91,
		Data2: 0x9E28,
		Data3: 0x4B86,
		Data4: [8]byte{0x90, 0xE9, 0x9E, 0x9F, 0x8A, 0x5E, 0xEF, 0xAF},
	}
)

// ITaskbarList3 interface
type ITaskbarList3 struct {
	vtbl *ITaskbarList3Vtbl
}

type ITaskbarList3Vtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	HrInit         uintptr
	AddTab         uintptr
	DeleteTab      uintptr
	ActivateTab    uintptr
	SetActiveAlt   uintptr
	// ITaskbarList2
	MarkFullscreenWindow uintptr
	// ITaskbarList3
	SetProgressValue      uintptr
	SetProgressState      uintptr
	RegisterTab           uintptr
	UnregisterTab         uintptr
	SetTabOrder           uintptr
	SetTabActive          uintptr
	ThumbBarAddButtons    uintptr
	ThumbBarUpdateButtons uintptr
	ThumbBarSetImageList  uintptr
	SetOverlayIcon        uintptr
	SetThumbnailTooltip   uintptr
	SetThumbnailClip      uintptr
}

func (tb *ITaskbarList3) Release() {
	syscall.Syscall(tb.vtbl.Release, 1, uintptr(unsafe.Pointer(tb)), 0, 0)
}

func (tb *ITaskbarList3) HrInit() error {
	ret, _, _ := syscall.Syscall(tb.vtbl.HrInit, 1, uintptr(unsafe.Pointer(tb)), 0, 0)
	if ret != S_OK {
		return fmt.Errorf("HrInit failed with code: %x", ret)
	}
	return nil
}

func (tb *ITaskbarList3) SetOverlayIcon(hwnd windows.HWND, hicon windows.Handle, desc string) error {
	var descPtr *uint16
	if desc != "" {
		descPtr, _ = syscall.UTF16PtrFromString(desc)
	}
	
	ret, _, _ := syscall.Syscall6(
		tb.vtbl.SetOverlayIcon,
		4,
		uintptr(unsafe.Pointer(tb)),
		uintptr(hwnd),
		uintptr(hicon),
		uintptr(unsafe.Pointer(descPtr)),
		0,
		0,
	)
	
	if ret != S_OK {
		return fmt.Errorf("SetOverlayIcon failed with code: %x", ret)
	}
	return nil
}

// SetTaskbarBadge creates an icon with the number and sets overlay on the main window.
func SetTaskbarBadge(ctx context.Context, count int) {
	log.Printf("[Badge] SetTaskbarBadge called with count: %d", count)
	
	// Initialize COM
	procCoInitializeEx.Call(0, COINIT_APARTMENTTHREADED)
	defer procCoUninitialize.Call()
	
	// Find the window
	hwnd := findMainWindow()
	if hwnd == 0 {
		log.Println("[Badge] No window found for this process")
		return
	}
	log.Printf("[Badge] Found window (hwnd: %v)", hwnd)
	
	// Create ITaskbarList3
	tb := createTaskbarList3()
	if tb == nil {
		log.Println("[Badge] Failed to create ITaskbarList3")
		return
	}
	defer tb.Release()
	
	// Initialize the taskbar list
	if err := tb.HrInit(); err != nil {
		log.Printf("[Badge] Failed to initialize taskbar list: %v", err)
		return
	}
	
	if count <= 0 {
		// Clear the badge
		log.Printf("[Badge] Clearing badge")
		if err := tb.SetOverlayIcon(hwnd, 0, ""); err != nil {
			log.Printf("[Badge] Failed to clear overlay icon: %v", err)
		} else {
			log.Println("[Badge] Badge cleared successfully")
		}
		return
	}
	
	// Create badge icon
	tmp := filepath.Join(os.TempDir(), fmt.Sprintf("versadumps_badge_%d.ico", count))
	log.Printf("[Badge] Creating badge icon at: %s", tmp)
	if err := createBadgeICO(tmp, count); err != nil {
		log.Printf("[Badge] Error creating badge ICO: %v", err)
		return
	}
	defer os.Remove(tmp)
	
	// Load the icon
	hicon := loadIconFromFile(tmp)
	if hicon == 0 {
		log.Println("[Badge] Failed to load icon from file")
		return
	}
	defer destroyIcon(hicon)
	log.Printf("[Badge] Icon loaded successfully (handle: %v)", hicon)
	
	// Set the overlay icon
	desc := fmt.Sprintf("%d messages", count)
	if err := tb.SetOverlayIcon(hwnd, hicon, desc); err != nil {
		log.Printf("[Badge] Failed to set overlay icon: %v", err)
	} else {
		log.Printf("[Badge] Badge set successfully with count: %d", count)
	}
}

func createTaskbarList3() *ITaskbarList3 {
	var tb *ITaskbarList3
	ret, _, _ := procCoCreateInstance.Call(
		uintptr(unsafe.Pointer(&CLSID_TaskbarList)),
		0,
		1, // CLSCTX_INPROC_SERVER
		uintptr(unsafe.Pointer(&IID_ITaskbarList3)),
		uintptr(unsafe.Pointer(&tb)),
	)
	
	if ret != S_OK {
		log.Printf("[Badge] CoCreateInstance failed with code: %x", ret)
		return nil
	}
	
	return tb
}

func findMainWindow() windows.HWND {
	pid := uint32(os.Getpid())
	var found windows.HWND
	
	cb := syscall.NewCallback(func(hwnd uintptr, lparam uintptr) uintptr {
		var procId uint32
		windows.GetWindowThreadProcessId(windows.HWND(hwnd), &procId)
		if procId == pid {
			// Check if window is visible
			if windows.IsWindowVisible(windows.HWND(hwnd)) {
				// Check if it's a top-level window (no parent)
				getParent := user32.NewProc("GetParent")
				parent, _, _ := getParent.Call(hwnd)
				if parent == 0 {
					// Get window text length to check if it has a title
					getWindowTextLength := user32.NewProc("GetWindowTextLengthW")
					length, _, _ := getWindowTextLength.Call(hwnd)
					if length > 0 {
						found = windows.HWND(hwnd)
						return 0 // Stop enumeration
					}
				}
			}
		}
		return 1 // Continue enumeration
	})
	
	windows.EnumWindows(cb, unsafe.Pointer(nil))
	return found
}

func loadIconFromFile(path string) windows.Handle {
	p, _ := syscall.UTF16PtrFromString(path)
	loadImage := user32.NewProc("LoadImageW")
	ret, _, _ := loadImage.Call(
		0,
		uintptr(unsafe.Pointer(p)),
		IMAGE_ICON,
		16, // width
		16, // height
		LR_LOADFROMFILE,
	)
	return windows.Handle(ret)
}

func destroyIcon(hicon windows.Handle) {
	destroyIcon := user32.NewProc("DestroyIcon")
	destroyIcon.Call(uintptr(hicon))
}

func createBadgeICO(path string, count int) error {
	const size = 16
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	
	// Draw transparent background
	draw.Draw(img, img.Bounds(), &image.Uniform{color.Transparent}, image.Point{}, draw.Src)
	
	// Draw red circle
	cx, cy := size/2, size/2
	r := size/2 - 1
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := x - cx
			dy := y - cy
			if dx*dx+dy*dy <= r*r {
				img.Set(x, y, color.RGBA{R: 220, G: 38, B: 38, A: 255})
			}
		}
	}
	
	// Draw the number
	label := fmt.Sprintf("%d", count)
	if count > 99 {
		label = "99+"
	}
	
	// Simple number drawing for small icon
	if count < 10 {
		// Single digit - center it
		drawDigit(img, label[0], 6, 5, color.White)
	} else if count < 100 {
		// Two digits
		drawDigit(img, label[0], 3, 5, color.White)
		drawDigit(img, label[1], 9, 5, color.White)
	} else {
		// 99+
		drawDigit(img, '9', 2, 5, color.White)
		drawDigit(img, '9', 7, 5, color.White)
		drawPixel(img, 12, 5, color.White)
		drawPixel(img, 13, 5, color.White)
		drawPixel(img, 12, 7, color.White)
		drawPixel(img, 13, 7, color.White)
	}
	
	// Save as ICO
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	
	return ico.Encode(f, img)
}

func drawDigit(img *image.RGBA, digit byte, x, y int, c color.Color) {
	// Simple 3x5 digit patterns
	patterns := map[byte][]string{
		'0': {"111", "101", "101", "101", "111"},
		'1': {"010", "110", "010", "010", "111"},
		'2': {"111", "001", "111", "100", "111"},
		'3': {"111", "001", "111", "001", "111"},
		'4': {"101", "101", "111", "001", "001"},
		'5': {"111", "100", "111", "001", "111"},
		'6': {"111", "100", "111", "101", "111"},
		'7': {"111", "001", "010", "100", "100"},
		'8': {"111", "101", "111", "101", "111"},
		'9': {"111", "101", "111", "001", "111"},
	}
	
	pattern, ok := patterns[digit]
	if !ok {
		return
	}
	
	for dy, row := range pattern {
		for dx, ch := range row {
			if ch == '1' {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
	}
}

func drawPixel(img *image.RGBA, x, y int, c color.Color) {
	if x >= 0 && x < img.Bounds().Max.X && y >= 0 && y < img.Bounds().Max.Y {
		img.Set(x, y, c)
	}
}

