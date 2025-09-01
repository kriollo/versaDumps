//go:build windows

package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
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
	// log.Printf("[Badge] SetTaskbarBadge called with count: %d", count)

	// Initialize COM
	procCoInitializeEx.Call(0, COINIT_APARTMENTTHREADED)
	defer procCoUninitialize.Call()

	// Find the window
	hwnd := findMainWindow()
	if hwnd == 0 {
		// log.Println("[Badge] No window found for this process")
		return
	}
	// log.Printf("[Badge] Found window (hwnd: %v)", hwnd)

	// Create ITaskbarList3
	tb := createTaskbarList3()
	if tb == nil {
		// log.Println("[Badge] Failed to create ITaskbarList3")
		return
	}
	defer tb.Release()

	// Initialize the taskbar list
	if err := tb.HrInit(); err != nil {
		// log.Printf("[Badge] Failed to initialize taskbar list: %v", err)
		return
	}

	if count <= 0 {
		// Clear the badge
		// log.Printf("[Badge] Clearing badge")
		if err := tb.SetOverlayIcon(hwnd, 0, ""); err != nil {
			// log.Printf("[Badge] Failed to clear overlay icon: %v", err)
		} else {
			// log.Println("[Badge] Badge cleared successfully")
		}
		return
	}

	// Create badge icon
	tmp := filepath.Join(os.TempDir(), fmt.Sprintf("versadumps_badge_%d.ico", count))
	// log.Printf("[Badge] Creating badge icon at: %s", tmp)
	if err := createBadgeICO(tmp, count); err != nil {
		// log.Printf("[Badge] Error creating badge ICO: %v", err)
		return
	}
	defer os.Remove(tmp)

	// Load the icon
	hicon := loadIconFromFile(tmp)
	if hicon == 0 {
		// log.Println("[Badge] Failed to load icon from file")
		return
	}
	defer destroyIcon(hicon)
	// log.Printf("[Badge] Icon loaded successfully (handle: %v)", hicon)

	// Set the overlay icon
	desc := fmt.Sprintf("%d messages", count)
	if err := tb.SetOverlayIcon(hwnd, hicon, desc); err != nil {
		// log.Printf("[Badge] Failed to set overlay icon: %v", err)
	} else {
		// log.Printf("[Badge] Badge set successfully with count: %d", count)
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
		// log.Printf("[Badge] CoCreateInstance failed with code: %x", ret)
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
		32, // width - 32x32 for maximum visibility
		32, // height - 32x32 for maximum visibility
		LR_LOADFROMFILE,
	)
	return windows.Handle(ret)
}

func destroyIcon(hicon windows.Handle) {
	destroyIcon := user32.NewProc("DestroyIcon")
	destroyIcon.Call(uintptr(hicon))
}

func createBadgeICO(path string, count int) error {
	const size = 32 // 32x32 for maximum visibility
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Draw transparent background
	draw.Draw(img, img.Bounds(), &image.Uniform{color.Transparent}, image.Point{}, draw.Src)

	// Use white background (winter white)
	bgColor := color.RGBA{R: 255, G: 255, B: 255, A: 255} // Pure white

	// Draw filled circle
	cx, cy := float64(size)/2, float64(size)/2
	r := float64(size)/2 - 1.0

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) + 0.5 - cx
			dy := float64(y) + 0.5 - cy
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist <= r {
				// Inside the circle
				img.Set(x, y, bgColor)
			} else if dist <= r+1.0 {
				// Antialiasing for smooth edges
				alpha := 1.0 - (dist - r)
				if alpha > 0 && alpha <= 1 {
					img.Set(x, y, color.RGBA{
						R: bgColor.R,
						G: bgColor.G,
						B: bgColor.B,
						A: uint8(255 * alpha),
					})
				}
			}
		}
	}

	// Format the count
	label := fmt.Sprintf("%d", count)
	if count > 99 {
		label = "99"
	}

	// Black text for contrast on white background
	textColor := color.RGBA{R: 0, G: 0, B: 0, A: 255}

	// Draw numbers with HUGE font for 32x32
	if count < 10 {
		// Single digit - centered with huge font
		drawHugeDigit(img, label[0], 11, 8, textColor)
	} else if count < 100 {
		// Two digits with better spacing
		drawBigDigit(img, label[0], 6, 9, textColor)
		drawBigDigit(img, label[1], 17, 9, textColor)
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

// drawSimpleDigit draws simple 3x5 digits optimized for 16x16 badges
func drawSimpleDigit(img *image.RGBA, digit byte, x, y int, c color.Color) {
	// Simple 3x5 patterns that are clear at small sizes
	switch digit {
	case '0':
		// Top
		drawPixel(img, x+1, y, c)
		// Sides
		drawPixel(img, x, y+1, c)
		drawPixel(img, x+2, y+1, c)
		drawPixel(img, x, y+2, c)
		drawPixel(img, x+2, y+2, c)
		drawPixel(img, x, y+3, c)
		drawPixel(img, x+2, y+3, c)
		// Bottom
		drawPixel(img, x+1, y+4, c)
	case '1':
		drawPixel(img, x+1, y, c)
		drawPixel(img, x+1, y+1, c)
		drawPixel(img, x+1, y+2, c)
		drawPixel(img, x+1, y+3, c)
		drawPixel(img, x+1, y+4, c)
	case '2':
		drawPixel(img, x, y, c)
		drawPixel(img, x+1, y, c)
		drawPixel(img, x+2, y, c)
		drawPixel(img, x+2, y+1, c)
		drawPixel(img, x, y+2, c)
		drawPixel(img, x+1, y+2, c)
		drawPixel(img, x+2, y+2, c)
		drawPixel(img, x, y+3, c)
		drawPixel(img, x, y+4, c)
		drawPixel(img, x+1, y+4, c)
		drawPixel(img, x+2, y+4, c)
	case '3':
		drawPixel(img, x, y, c)
		drawPixel(img, x+1, y, c)
		drawPixel(img, x+2, y, c)
		drawPixel(img, x+2, y+1, c)
		drawPixel(img, x+1, y+2, c)
		drawPixel(img, x+2, y+2, c)
		drawPixel(img, x+2, y+3, c)
		drawPixel(img, x, y+4, c)
		drawPixel(img, x+1, y+4, c)
		drawPixel(img, x+2, y+4, c)
	case '4':
		drawPixel(img, x, y, c)
		drawPixel(img, x+2, y, c)
		drawPixel(img, x, y+1, c)
		drawPixel(img, x+2, y+1, c)
		drawPixel(img, x, y+2, c)
		drawPixel(img, x+1, y+2, c)
		drawPixel(img, x+2, y+2, c)
		drawPixel(img, x+2, y+3, c)
		drawPixel(img, x+2, y+4, c)
	case '5':
		drawPixel(img, x, y, c)
		drawPixel(img, x+1, y, c)
		drawPixel(img, x+2, y, c)
		drawPixel(img, x, y+1, c)
		drawPixel(img, x, y+2, c)
		drawPixel(img, x+1, y+2, c)
		drawPixel(img, x+2, y+2, c)
		drawPixel(img, x+2, y+3, c)
		drawPixel(img, x, y+4, c)
		drawPixel(img, x+1, y+4, c)
		drawPixel(img, x+2, y+4, c)
	case '6':
		drawPixel(img, x, y, c)
		drawPixel(img, x+1, y, c)
		drawPixel(img, x+2, y, c)
		drawPixel(img, x, y+1, c)
		drawPixel(img, x, y+2, c)
		drawPixel(img, x+1, y+2, c)
		drawPixel(img, x+2, y+2, c)
		drawPixel(img, x, y+3, c)
		drawPixel(img, x+2, y+3, c)
		drawPixel(img, x, y+4, c)
		drawPixel(img, x+1, y+4, c)
		drawPixel(img, x+2, y+4, c)
	case '7':
		drawPixel(img, x, y, c)
		drawPixel(img, x+1, y, c)
		drawPixel(img, x+2, y, c)
		drawPixel(img, x+2, y+1, c)
		drawPixel(img, x+2, y+2, c)
		drawPixel(img, x+1, y+3, c)
		drawPixel(img, x+1, y+4, c)
	case '8':
		drawPixel(img, x, y, c)
		drawPixel(img, x+1, y, c)
		drawPixel(img, x+2, y, c)
		drawPixel(img, x, y+1, c)
		drawPixel(img, x+2, y+1, c)
		drawPixel(img, x, y+2, c)
		drawPixel(img, x+1, y+2, c)
		drawPixel(img, x+2, y+2, c)
		drawPixel(img, x, y+3, c)
		drawPixel(img, x+2, y+3, c)
		drawPixel(img, x, y+4, c)
		drawPixel(img, x+1, y+4, c)
		drawPixel(img, x+2, y+4, c)
	case '9':
		drawPixel(img, x, y, c)
		drawPixel(img, x+1, y, c)
		drawPixel(img, x+2, y, c)
		drawPixel(img, x, y+1, c)
		drawPixel(img, x+2, y+1, c)
		drawPixel(img, x, y+2, c)
		drawPixel(img, x+1, y+2, c)
		drawPixel(img, x+2, y+2, c)
		drawPixel(img, x+2, y+3, c)
		drawPixel(img, x, y+4, c)
		drawPixel(img, x+1, y+4, c)
		drawPixel(img, x+2, y+4, c)
	}
}

// drawDigitBold draws bolder, more visible digits
func drawDigitBold(img *image.RGBA, digit byte, x, y int, c color.Color) {
	// Bolder 3x5 digit patterns with thicker strokes
	patterns := map[byte][][]byte{
		'0': {
			{1, 1, 1},
			{1, 0, 1},
			{1, 0, 1},
			{1, 0, 1},
			{1, 1, 1},
		},
		'1': {
			{0, 1, 0},
			{1, 1, 0},
			{0, 1, 0},
			{0, 1, 0},
			{1, 1, 1},
		},
		'2': {
			{1, 1, 1},
			{0, 0, 1},
			{1, 1, 1},
			{1, 0, 0},
			{1, 1, 1},
		},
		'3': {
			{1, 1, 1},
			{0, 0, 1},
			{1, 1, 1},
			{0, 0, 1},
			{1, 1, 1},
		},
		'4': {
			{1, 0, 1},
			{1, 0, 1},
			{1, 1, 1},
			{0, 0, 1},
			{0, 0, 1},
		},
		'5': {
			{1, 1, 1},
			{1, 0, 0},
			{1, 1, 1},
			{0, 0, 1},
			{1, 1, 1},
		},
		'6': {
			{1, 1, 1},
			{1, 0, 0},
			{1, 1, 1},
			{1, 0, 1},
			{1, 1, 1},
		},
		'7': {
			{1, 1, 1},
			{0, 0, 1},
			{0, 1, 0},
			{0, 1, 0},
			{0, 1, 0},
		},
		'8': {
			{1, 1, 1},
			{1, 0, 1},
			{1, 1, 1},
			{1, 0, 1},
			{1, 1, 1},
		},
		'9': {
			{1, 1, 1},
			{1, 0, 1},
			{1, 1, 1},
			{0, 0, 1},
			{1, 1, 1},
		},
	}

	pattern, ok := patterns[digit]
	if !ok {
		return
	}

	// Draw with double thickness for better visibility
	for dy, row := range pattern {
		for dx, pixel := range row {
			if pixel == 1 {
				// Draw main pixel
				drawPixel(img, x+dx, y+dy, c)
				// Add outline for better visibility (optional)
				// This makes the digits slightly bolder
				if dy > 0 && pattern[dy-1][dx] == 0 {
					// Top edge
				}
				if dy < len(pattern)-1 && pattern[dy+1][dx] == 0 {
					// Bottom edge
				}
			}
		}
	}
}

// drawDigitLarge draws larger, more prominent digits for single numbers
func drawDigitLarge(img *image.RGBA, digit byte, x, y int, c color.Color) {
	// Larger 5x7 digit patterns for better visibility
	patterns := map[byte][][]byte{
		'0': {
			{0, 1, 1, 1, 0},
			{1, 1, 0, 1, 1},
			{1, 0, 0, 0, 1},
			{1, 0, 0, 0, 1},
			{1, 0, 0, 0, 1},
			{1, 1, 0, 1, 1},
			{0, 1, 1, 1, 0},
		},
		'1': {
			{0, 0, 1, 0, 0},
			{0, 1, 1, 0, 0},
			{1, 0, 1, 0, 0},
			{0, 0, 1, 0, 0},
			{0, 0, 1, 0, 0},
			{0, 0, 1, 0, 0},
			{1, 1, 1, 1, 1},
		},
		'2': {
			{0, 1, 1, 1, 0},
			{1, 1, 0, 1, 1},
			{0, 0, 0, 1, 1},
			{0, 0, 1, 1, 0},
			{0, 1, 1, 0, 0},
			{1, 1, 0, 0, 0},
			{1, 1, 1, 1, 1},
		},
		'3': {
			{1, 1, 1, 1, 0},
			{0, 0, 0, 1, 1},
			{0, 0, 0, 1, 0},
			{0, 1, 1, 1, 0},
			{0, 0, 0, 1, 1},
			{0, 0, 0, 1, 1},
			{1, 1, 1, 1, 0},
		},
		'4': {
			{0, 0, 1, 1, 0},
			{0, 1, 0, 1, 0},
			{1, 0, 0, 1, 0},
			{1, 1, 1, 1, 1},
			{0, 0, 0, 1, 0},
			{0, 0, 0, 1, 0},
			{0, 0, 0, 1, 0},
		},
		'5': {
			{1, 1, 1, 1, 1},
			{1, 0, 0, 0, 0},
			{1, 1, 1, 1, 0},
			{0, 0, 0, 1, 1},
			{0, 0, 0, 0, 1},
			{1, 0, 0, 1, 1},
			{0, 1, 1, 1, 0},
		},
		'6': {
			{0, 1, 1, 1, 0},
			{1, 1, 0, 0, 0},
			{1, 0, 0, 0, 0},
			{1, 1, 1, 1, 0},
			{1, 0, 0, 1, 1},
			{1, 1, 0, 1, 1},
			{0, 1, 1, 1, 0},
		},
		'7': {
			{1, 1, 1, 1, 1},
			{0, 0, 0, 1, 1},
			{0, 0, 0, 1, 0},
			{0, 0, 1, 0, 0},
			{0, 0, 1, 0, 0},
			{0, 1, 0, 0, 0},
			{0, 1, 0, 0, 0},
		},
		'8': {
			{0, 1, 1, 1, 0},
			{1, 1, 0, 1, 1},
			{1, 1, 0, 1, 1},
			{0, 1, 1, 1, 0},
			{1, 1, 0, 1, 1},
			{1, 1, 0, 1, 1},
			{0, 1, 1, 1, 0},
		},
		'9': {
			{0, 1, 1, 1, 0},
			{1, 1, 0, 1, 1},
			{1, 1, 0, 0, 1},
			{0, 1, 1, 1, 1},
			{0, 0, 0, 0, 1},
			{0, 0, 0, 1, 1},
			{0, 1, 1, 1, 0},
		},
	}

	pattern, ok := patterns[digit]
	if !ok {
		return
	}

	for dy, row := range pattern {
		for dx, pixel := range row {
			if pixel == 1 {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
	}
}

// drawDigitMedium draws medium-sized digits for double-digit numbers
func drawDigitMedium(img *image.RGBA, digit byte, x, y int, c color.Color) {
	// Medium 4x6 digit patterns
	patterns := map[byte][][]byte{
		'0': {
			{1, 1, 1, 1},
			{1, 0, 0, 1},
			{1, 0, 0, 1},
			{1, 0, 0, 1},
			{1, 0, 0, 1},
			{1, 1, 1, 1},
		},
		'1': {
			{0, 1, 1, 0},
			{1, 1, 1, 0},
			{0, 1, 1, 0},
			{0, 1, 1, 0},
			{0, 1, 1, 0},
			{1, 1, 1, 1},
		},
		'2': {
			{1, 1, 1, 1},
			{0, 0, 0, 1},
			{0, 0, 1, 1},
			{1, 1, 1, 0},
			{1, 0, 0, 0},
			{1, 1, 1, 1},
		},
		'3': {
			{1, 1, 1, 1},
			{0, 0, 0, 1},
			{1, 1, 1, 1},
			{0, 0, 0, 1},
			{0, 0, 0, 1},
			{1, 1, 1, 1},
		},
		'4': {
			{1, 0, 0, 1},
			{1, 0, 0, 1},
			{1, 0, 0, 1},
			{1, 1, 1, 1},
			{0, 0, 0, 1},
			{0, 0, 0, 1},
		},
		'5': {
			{1, 1, 1, 1},
			{1, 0, 0, 0},
			{1, 1, 1, 1},
			{0, 0, 0, 1},
			{0, 0, 0, 1},
			{1, 1, 1, 1},
		},
		'6': {
			{1, 1, 1, 1},
			{1, 0, 0, 0},
			{1, 1, 1, 1},
			{1, 0, 0, 1},
			{1, 0, 0, 1},
			{1, 1, 1, 1},
		},
		'7': {
			{1, 1, 1, 1},
			{0, 0, 0, 1},
			{0, 0, 1, 0},
			{0, 1, 0, 0},
			{0, 1, 0, 0},
			{0, 1, 0, 0},
		},
		'8': {
			{1, 1, 1, 1},
			{1, 0, 0, 1},
			{1, 1, 1, 1},
			{1, 0, 0, 1},
			{1, 0, 0, 1},
			{1, 1, 1, 1},
		},
		'9': {
			{1, 1, 1, 1},
			{1, 0, 0, 1},
			{1, 1, 1, 1},
			{0, 0, 0, 1},
			{0, 0, 0, 1},
			{1, 1, 1, 1},
		},
	}

	pattern, ok := patterns[digit]
	if !ok {
		return
	}

	for dy, row := range pattern {
		for dx, pixel := range row {
			if pixel == 1 {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
	}
}

// drawGiantDigit draws huge digits for single number display (16x20 pixels)
func drawGiantDigit(img *image.RGBA, digit byte, x, y int, c color.Color) {
	switch digit {
	case '0':
		// Draw a large 0
		for dy := 0; dy < 20; dy++ {
			for dx := 0; dx < 16; dx++ {
				if (dy < 4 || dy > 15) && dx >= 4 && dx < 12 { // Top and bottom
					drawPixel(img, x+dx, y+dy, c)
				} else if (dx < 4 || dx > 11) && dy >= 4 && dy <= 15 { // Left and right sides
					drawPixel(img, x+dx, y+dy, c)
				}
			}
		}
	case '1':
		// Draw a large 1
		for dy := 0; dy < 20; dy++ {
			if dy < 4 {
				drawPixel(img, x+6, y+dy, c)
				drawPixel(img, x+7, y+dy, c)
				drawPixel(img, x+8, y+dy, c)
				if dy == 3 {
					drawPixel(img, x+4, y+dy, c)
					drawPixel(img, x+5, y+dy, c)
				}
			} else if dy > 15 {
				for dx := 2; dx < 14; dx++ {
					drawPixel(img, x+dx, y+dy, c)
				}
			} else {
				drawPixel(img, x+6, y+dy, c)
				drawPixel(img, x+7, y+dy, c)
				drawPixel(img, x+8, y+dy, c)
			}
		}
	case '2':
		// Draw a large 2
		for dy := 0; dy < 20; dy++ {
			for dx := 0; dx < 16; dx++ {
				if dy < 4 && dx >= 2 && dx < 14 { // Top
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 4 && dy < 8 && dx > 11 { // Right side top
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 8 && dy < 12 && dx >= 2 && dx < 14 { // Middle
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 12 && dy < 16 && dx < 4 { // Left side bottom
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 16 && dx >= 2 && dx < 14 { // Bottom
					drawPixel(img, x+dx, y+dy, c)
				}
			}
		}
	case '3':
		// Draw a large 3
		for dy := 0; dy < 20; dy++ {
			for dx := 0; dx < 16; dx++ {
				if dy < 4 && dx >= 2 && dx < 14 { // Top
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 4 && dy < 8 && dx > 11 { // Right side top
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 8 && dy < 12 && dx >= 6 && dx < 14 { // Middle
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 12 && dy < 16 && dx > 11 { // Right side bottom
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 16 && dx >= 2 && dx < 14 { // Bottom
					drawPixel(img, x+dx, y+dy, c)
				}
			}
		}
	case '4':
		// Draw a large 4
		for dy := 0; dy < 20; dy++ {
			for dx := 0; dx < 16; dx++ {
				if dy < 8 && dx < 4 { // Left side top
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 8 && dy < 12 && dx >= 0 && dx < 14 { // Middle bar
					drawPixel(img, x+dx, y+dy, c)
				} else if dx > 11 { // Right side full
					drawPixel(img, x+dx, y+dy, c)
				}
			}
		}
	case '5':
		// Draw a large 5
		for dy := 0; dy < 20; dy++ {
			for dx := 0; dx < 16; dx++ {
				if dy < 4 && dx >= 2 && dx < 14 { // Top
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 4 && dy < 8 && dx < 4 { // Left side top
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 8 && dy < 12 && dx >= 2 && dx < 14 { // Middle
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 12 && dy < 16 && dx > 11 { // Right side bottom
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 16 && dx >= 2 && dx < 14 { // Bottom
					drawPixel(img, x+dx, y+dy, c)
				}
			}
		}
	case '6':
		// Draw a large 6
		for dy := 0; dy < 20; dy++ {
			for dx := 0; dx < 16; dx++ {
				if dy < 4 && dx >= 2 && dx < 14 { // Top
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 4 && dy < 16 && dx < 4 { // Left side
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 8 && dy < 12 && dx >= 2 && dx < 14 { // Middle
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 12 && dy < 16 && dx > 11 { // Right side bottom
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 16 && dx >= 2 && dx < 14 { // Bottom
					drawPixel(img, x+dx, y+dy, c)
				}
			}
		}
	case '7':
		// Draw a large 7
		for dy := 0; dy < 20; dy++ {
			for dx := 0; dx < 16; dx++ {
				if dy < 4 && dx >= 2 && dx < 14 { // Top
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 4 && dx > 11 { // Right side slanted
					if dx-dy/2 > 7 {
						drawPixel(img, x+dx, y+dy, c)
					}
				}
			}
		}
	case '8':
		// Draw a large 8
		for dy := 0; dy < 20; dy++ {
			for dx := 0; dx < 16; dx++ {
				if (dy < 4 || dy > 15) && dx >= 2 && dx < 14 { // Top and bottom
					drawPixel(img, x+dx, y+dy, c)
				} else if (dx < 4 || dx > 11) && ((dy >= 4 && dy < 8) || (dy >= 12 && dy <= 15)) { // Sides
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 8 && dy < 12 && dx >= 2 && dx < 14 { // Middle
					drawPixel(img, x+dx, y+dy, c)
				}
			}
		}
	case '9':
		// Draw a large 9
		for dy := 0; dy < 20; dy++ {
			for dx := 0; dx < 16; dx++ {
				if dy < 4 && dx >= 2 && dx < 14 { // Top
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 4 && dy < 8 && (dx < 4 || dx > 11) { // Sides top
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 8 && dy < 12 && dx >= 2 && dx < 14 { // Middle
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 12 && dx > 11 { // Right side bottom
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 16 && dx >= 2 && dx < 14 { // Bottom
					drawPixel(img, x+dx, y+dy, c)
				}
			}
		}
	}
}

// drawLargeDigit draws large digits for two-digit display (10x14 pixels)
func drawLargeDigit(img *image.RGBA, digit byte, x, y int, c color.Color) {
	switch digit {
	case '0':
		for dy := 0; dy < 14; dy++ {
			for dx := 0; dx < 10; dx++ {
				if (dy < 3 || dy > 10) && dx >= 2 && dx < 8 { // Top and bottom
					drawPixel(img, x+dx, y+dy, c)
				} else if (dx < 3 || dx > 6) && dy >= 3 && dy <= 10 { // Left and right sides
					drawPixel(img, x+dx, y+dy, c)
				}
			}
		}
	case '1':
		for dy := 0; dy < 14; dy++ {
			if dy < 3 {
				drawPixel(img, x+4, y+dy, c)
				drawPixel(img, x+5, y+dy, c)
				if dy == 2 {
					drawPixel(img, x+3, y+dy, c)
				}
			} else if dy > 10 {
				for dx := 1; dx < 9; dx++ {
					drawPixel(img, x+dx, y+dy, c)
				}
			} else {
				drawPixel(img, x+4, y+dy, c)
				drawPixel(img, x+5, y+dy, c)
			}
		}
	case '2':
		for dy := 0; dy < 14; dy++ {
			for dx := 0; dx < 10; dx++ {
				if dy < 3 && dx >= 1 && dx < 9 { // Top
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 3 && dy < 6 && dx > 6 { // Right side top
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 6 && dy < 8 && dx >= 1 && dx < 9 { // Middle
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 8 && dy < 11 && dx < 3 { // Left side bottom
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 11 && dx >= 1 && dx < 9 { // Bottom
					drawPixel(img, x+dx, y+dy, c)
				}
			}
		}
	case '3':
		for dy := 0; dy < 14; dy++ {
			for dx := 0; dx < 10; dx++ {
				if dy < 3 && dx >= 1 && dx < 9 { // Top
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 3 && dy < 6 && dx > 6 { // Right side top
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 6 && dy < 8 && dx >= 4 && dx < 9 { // Middle
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 8 && dy < 11 && dx > 6 { // Right side bottom
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 11 && dx >= 1 && dx < 9 { // Bottom
					drawPixel(img, x+dx, y+dy, c)
				}
			}
		}
	case '4':
		for dy := 0; dy < 14; dy++ {
			for dx := 0; dx < 10; dx++ {
				if dy < 6 && dx < 3 { // Left side top
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 6 && dy < 8 && dx >= 0 && dx < 9 { // Middle bar
					drawPixel(img, x+dx, y+dy, c)
				} else if dx > 6 { // Right side full
					drawPixel(img, x+dx, y+dy, c)
				}
			}
		}
	case '5':
		for dy := 0; dy < 14; dy++ {
			for dx := 0; dx < 10; dx++ {
				if dy < 3 && dx >= 1 && dx < 9 { // Top
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 3 && dy < 6 && dx < 3 { // Left side top
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 6 && dy < 8 && dx >= 1 && dx < 9 { // Middle
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 8 && dy < 11 && dx > 6 { // Right side bottom
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 11 && dx >= 1 && dx < 9 { // Bottom
					drawPixel(img, x+dx, y+dy, c)
				}
			}
		}
	case '6':
		for dy := 0; dy < 14; dy++ {
			for dx := 0; dx < 10; dx++ {
				if dy < 3 && dx >= 1 && dx < 9 { // Top
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 3 && dy < 11 && dx < 3 { // Left side
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 6 && dy < 8 && dx >= 1 && dx < 9 { // Middle
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 8 && dy < 11 && dx > 6 { // Right side bottom
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 11 && dx >= 1 && dx < 9 { // Bottom
					drawPixel(img, x+dx, y+dy, c)
				}
			}
		}
	case '7':
		for dy := 0; dy < 14; dy++ {
			for dx := 0; dx < 10; dx++ {
				if dy < 3 && dx >= 1 && dx < 9 { // Top
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 3 && dx > 6 && dx-dy/2 > 4 { // Right side slanted
					drawPixel(img, x+dx, y+dy, c)
				}
			}
		}
	case '8':
		for dy := 0; dy < 14; dy++ {
			for dx := 0; dx < 10; dx++ {
				if (dy < 3 || dy > 10) && dx >= 1 && dx < 9 { // Top and bottom
					drawPixel(img, x+dx, y+dy, c)
				} else if (dx < 3 || dx > 6) && ((dy >= 3 && dy < 6) || (dy >= 8 && dy <= 10)) { // Sides
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 6 && dy < 8 && dx >= 1 && dx < 9 { // Middle
					drawPixel(img, x+dx, y+dy, c)
				}
			}
		}
	case '9':
		for dy := 0; dy < 14; dy++ {
			for dx := 0; dx < 10; dx++ {
				if dy < 3 && dx >= 1 && dx < 9 { // Top
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 3 && dy < 6 && (dx < 3 || dx > 6) { // Sides top
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 6 && dy < 8 && dx >= 1 && dx < 9 { // Middle
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 8 && dx > 6 { // Right side bottom
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 11 && dx >= 1 && dx < 9 { // Bottom
					drawPixel(img, x+dx, y+dy, c)
				}
			}
		}
	}
}

// drawMediumDigit draws medium digits (8x10 pixels)
func drawMediumDigit(img *image.RGBA, digit byte, x, y int, c color.Color) {
	switch digit {
	case '9':
		for dy := 0; dy < 10; dy++ {
			for dx := 0; dx < 8; dx++ {
				if dy < 2 && dx >= 1 && dx < 7 { // Top
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 2 && dy < 4 && (dx < 2 || dx > 5) { // Sides top
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 4 && dy < 6 && dx >= 1 && dx < 7 { // Middle
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 6 && dx > 5 { // Right side bottom
					drawPixel(img, x+dx, y+dy, c)
				} else if dy >= 8 && dx >= 1 && dx < 7 { // Bottom
					drawPixel(img, x+dx, y+dy, c)
				}
			}
		}
	}
}

// drawPlusSign draws a plus sign (6x6 pixels)
func drawPlusSign(img *image.RGBA, x, y int, c color.Color) {
	// Horizontal line
	for dx := 0; dx < 6; dx++ {
		drawPixel(img, x+dx, y+2, c)
		drawPixel(img, x+dx, y+3, c)
	}
	// Vertical line
	for dy := 0; dy < 6; dy++ {
		drawPixel(img, x+2, y+dy, c)
		drawPixel(img, x+3, y+dy, c)
	}
}

// drawHugeDigit draws huge bold digits for single numbers in 32x32 badges (10x14 pixels)
func drawHugeDigit(img *image.RGBA, digit byte, x, y int, c color.Color) {
	switch digit {
	case '0':
		// Top bar with thickness
		for dy := 0; dy < 3; dy++ {
			for dx := 2; dx < 8; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Left side
		for dy := 1; dy < 13; dy++ {
			for dx := 0; dx < 3; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Right side
		for dy := 1; dy < 13; dy++ {
			for dx := 7; dx < 10; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Bottom bar
		for dy := 11; dy < 14; dy++ {
			for dx := 2; dx < 8; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
	case '1':
		// Thick vertical line
		for dy := 0; dy < 14; dy++ {
			for dx := 3; dx < 7; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Top serif
		for dx := 1; dx < 4; dx++ {
			drawPixel(img, x+dx, y+2, c)
			drawPixel(img, x+dx, y+3, c)
		}
	case '2':
		// Top bar
		for dy := 0; dy < 3; dy++ {
			for dx := 1; dx < 9; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Right side top
		for dy := 3; dy < 6; dy++ {
			for dx := 6; dx < 9; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Middle bar
		for dy := 6; dy < 8; dy++ {
			for dx := 1; dx < 9; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Left side bottom
		for dy := 8; dy < 11; dy++ {
			for dx := 1; dx < 4; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Bottom bar
		for dy := 11; dy < 14; dy++ {
			for dx := 1; dx < 9; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
	case '3':
		// Top bar
		for dy := 0; dy < 3; dy++ {
			for dx := 1; dx < 9; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Right side top
		for dy := 3; dy < 6; dy++ {
			for dx := 6; dx < 9; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Middle bar
		for dy := 6; dy < 8; dy++ {
			for dx := 3; dx < 9; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Right side bottom
		for dy := 8; dy < 11; dy++ {
			for dx := 6; dx < 9; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Bottom bar
		for dy := 11; dy < 14; dy++ {
			for dx := 1; dx < 9; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
	case '4':
		// Left vertical
		for dy := 0; dy < 8; dy++ {
			for dx := 1; dx < 4; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Middle bar
		for dy := 6; dy < 8; dy++ {
			for dx := 1; dx < 9; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Right vertical
		for dy := 0; dy < 14; dy++ {
			for dx := 6; dx < 9; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
	case '5':
		// Top bar
		for dy := 0; dy < 3; dy++ {
			for dx := 1; dx < 9; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Left side
		for dy := 3; dy < 6; dy++ {
			for dx := 1; dx < 4; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Middle bar
		for dy := 6; dy < 8; dy++ {
			for dx := 1; dx < 9; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Right side
		for dy := 8; dy < 11; dy++ {
			for dx := 6; dx < 9; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Bottom bar
		for dy := 11; dy < 14; dy++ {
			for dx := 1; dx < 9; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
	case '6':
		// Top bar
		for dy := 0; dy < 3; dy++ {
			for dx := 2; dx < 8; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Left side
		for dy := 1; dy < 13; dy++ {
			for dx := 0; dx < 3; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Middle bar
		for dy := 6; dy < 8; dy++ {
			for dx := 3; dx < 9; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Right side bottom
		for dy := 8; dy < 11; dy++ {
			for dx := 7; dx < 10; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Bottom bar
		for dy := 11; dy < 14; dy++ {
			for dx := 2; dx < 8; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
	case '7':
		// Top bar
		for dy := 0; dy < 3; dy++ {
			for dx := 0; dx < 9; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Diagonal
		for dy := 3; dy < 14; dy++ {
			dx := 8 - (dy-3)/2
			if dx >= 2 {
				drawPixel(img, x+dx, y+dy, c)
				drawPixel(img, x+dx-1, y+dy, c)
				drawPixel(img, x+dx-2, y+dy, c)
			}
		}
	case '8':
		// Top bar
		for dy := 0; dy < 3; dy++ {
			for dx := 2; dx < 8; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Top sides
		for dy := 1; dy < 6; dy++ {
			for dx := 0; dx < 3; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
			for dx := 7; dx < 10; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Middle bar
		for dy := 6; dy < 8; dy++ {
			for dx := 2; dx < 8; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Bottom sides
		for dy := 8; dy < 13; dy++ {
			for dx := 0; dx < 3; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
			for dx := 7; dx < 10; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Bottom bar
		for dy := 11; dy < 14; dy++ {
			for dx := 2; dx < 8; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
	case '9':
		// Top bar
		for dy := 0; dy < 3; dy++ {
			for dx := 2; dx < 8; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Top sides
		for dy := 1; dy < 6; dy++ {
			for dx := 0; dx < 3; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
			for dx := 7; dx < 10; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Middle bar
		for dy := 6; dy < 8; dy++ {
			for dx := 2; dx < 10; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Right side
		for dy := 8; dy < 13; dy++ {
			for dx := 7; dx < 10; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Bottom bar
		for dy := 11; dy < 14; dy++ {
			for dx := 2; dx < 8; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
	}
}

// drawBigDigit draws big digits for two-digit numbers in 32x32 badges (8x12 pixels)
func drawBigDigit(img *image.RGBA, digit byte, x, y int, c color.Color) {
	switch digit {
	case '0':
		// Top bar
		for dy := 0; dy < 2; dy++ {
			for dx := 2; dx < 6; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Sides
		for dy := 1; dy < 11; dy++ {
			for dx := 0; dx < 2; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
			for dx := 6; dx < 8; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Bottom bar
		for dy := 10; dy < 12; dy++ {
			for dx := 2; dx < 6; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
	case '1':
		// Vertical line
		for dy := 0; dy < 12; dy++ {
			for dx := 3; dx < 5; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Top serif
		drawPixel(img, x+2, y+1, c)
		drawPixel(img, x+2, y+2, c)
	case '2':
		// Top bar
		for dy := 0; dy < 2; dy++ {
			for dx := 1; dx < 7; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Right side
		for dy := 2; dy < 5; dy++ {
			for dx := 5; dx < 7; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Middle
		for dy := 5; dy < 7; dy++ {
			for dx := 1; dx < 7; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Left side
		for dy := 7; dy < 10; dy++ {
			for dx := 1; dx < 3; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Bottom bar
		for dy := 10; dy < 12; dy++ {
			for dx := 1; dx < 7; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
	case '3':
		// Top bar
		for dy := 0; dy < 2; dy++ {
			for dx := 1; dx < 7; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Right side top
		for dy := 2; dy < 5; dy++ {
			for dx := 5; dx < 7; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Middle
		for dy := 5; dy < 7; dy++ {
			for dx := 2; dx < 7; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Right side bottom
		for dy := 7; dy < 10; dy++ {
			for dx := 5; dx < 7; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Bottom bar
		for dy := 10; dy < 12; dy++ {
			for dx := 1; dx < 7; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
	case '4':
		// Left vertical
		for dy := 0; dy < 7; dy++ {
			for dx := 1; dx < 3; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Middle bar
		for dy := 5; dy < 7; dy++ {
			for dx := 1; dx < 7; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Right vertical
		for dy := 0; dy < 12; dy++ {
			for dx := 5; dx < 7; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
	case '5':
		// Top bar
		for dy := 0; dy < 2; dy++ {
			for dx := 1; dx < 7; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Left side
		for dy := 2; dy < 5; dy++ {
			for dx := 1; dx < 3; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Middle
		for dy := 5; dy < 7; dy++ {
			for dx := 1; dx < 7; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Right side
		for dy := 7; dy < 10; dy++ {
			for dx := 5; dx < 7; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Bottom bar
		for dy := 10; dy < 12; dy++ {
			for dx := 1; dx < 7; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
	case '6':
		// Top bar
		for dy := 0; dy < 2; dy++ {
			for dx := 2; dx < 6; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Left side
		for dy := 1; dy < 11; dy++ {
			for dx := 0; dx < 2; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Middle
		for dy := 5; dy < 7; dy++ {
			for dx := 2; dx < 7; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Right side bottom
		for dy := 7; dy < 10; dy++ {
			for dx := 6; dx < 8; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Bottom bar
		for dy := 10; dy < 12; dy++ {
			for dx := 2; dx < 6; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
	case '7':
		// Top bar
		for dy := 0; dy < 2; dy++ {
			for dx := 0; dx < 7; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Diagonal
		for dy := 2; dy < 12; dy++ {
			dx := 6 - (dy-2)/2
			if dx >= 2 {
				drawPixel(img, x+dx, y+dy, c)
				drawPixel(img, x+dx-1, y+dy, c)
			}
		}
	case '8':
		// Top bar
		for dy := 0; dy < 2; dy++ {
			for dx := 2; dx < 6; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Top sides
		for dy := 1; dy < 5; dy++ {
			for dx := 0; dx < 2; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
			for dx := 6; dx < 8; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Middle
		for dy := 5; dy < 7; dy++ {
			for dx := 2; dx < 6; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Bottom sides
		for dy := 7; dy < 11; dy++ {
			for dx := 0; dx < 2; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
			for dx := 6; dx < 8; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Bottom bar
		for dy := 10; dy < 12; dy++ {
			for dx := 2; dx < 6; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
	case '9':
		// Top bar
		for dy := 0; dy < 2; dy++ {
			for dx := 2; dx < 6; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Top sides
		for dy := 1; dy < 5; dy++ {
			for dx := 0; dx < 2; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
			for dx := 6; dx < 8; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Middle
		for dy := 5; dy < 7; dy++ {
			for dx := 2; dx < 8; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Right side
		for dy := 7; dy < 11; dy++ {
			for dx := 6; dx < 8; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
		// Bottom bar
		for dy := 10; dy < 12; dy++ {
			for dx := 2; dx < 6; dx++ {
				drawPixel(img, x+dx, y+dy, c)
			}
		}
	}
}

// drawLargerDigit draws larger digits optimized for 24x24 badges (5x9 pixels)
func drawLargerDigit(img *image.RGBA, digit byte, x, y int, c color.Color) {
	switch digit {
	case '0':
		// Top bar
		for i := 1; i < 4; i++ {
			drawPixel(img, x+i, y, c)
			drawPixel(img, x+i, y+1, c)
		}
		// Left side
		for i := 0; i < 9; i++ {
			drawPixel(img, x, y+i, c)
			drawPixel(img, x+1, y+i, c)
		}
		// Right side
		for i := 0; i < 9; i++ {
			drawPixel(img, x+3, y+i, c)
			drawPixel(img, x+4, y+i, c)
		}
		// Bottom bar
		for i := 1; i < 4; i++ {
			drawPixel(img, x+i, y+7, c)
			drawPixel(img, x+i, y+8, c)
		}
	case '1':
		// Main vertical line
		for i := 0; i < 9; i++ {
			drawPixel(img, x+2, y+i, c)
			drawPixel(img, x+3, y+i, c)
		}
		// Top diagonal
		drawPixel(img, x+1, y+1, c)
		drawPixel(img, x, y+2, c)
	case '2':
		// Top bar
		for i := 0; i < 5; i++ {
			drawPixel(img, x+i, y, c)
			drawPixel(img, x+i, y+1, c)
		}
		// Top right
		drawPixel(img, x+3, y+2, c)
		drawPixel(img, x+4, y+2, c)
		drawPixel(img, x+3, y+3, c)
		drawPixel(img, x+4, y+3, c)
		// Middle bar
		for i := 0; i < 5; i++ {
			drawPixel(img, x+i, y+4, c)
		}
		// Bottom left
		drawPixel(img, x, y+5, c)
		drawPixel(img, x+1, y+5, c)
		drawPixel(img, x, y+6, c)
		drawPixel(img, x+1, y+6, c)
		// Bottom bar
		for i := 0; i < 5; i++ {
			drawPixel(img, x+i, y+7, c)
			drawPixel(img, x+i, y+8, c)
		}
	case '3':
		// Top bar
		for i := 0; i < 5; i++ {
			drawPixel(img, x+i, y, c)
			drawPixel(img, x+i, y+1, c)
		}
		// Right side top
		drawPixel(img, x+3, y+2, c)
		drawPixel(img, x+4, y+2, c)
		drawPixel(img, x+3, y+3, c)
		drawPixel(img, x+4, y+3, c)
		// Middle bar
		for i := 1; i < 5; i++ {
			drawPixel(img, x+i, y+4, c)
		}
		// Right side bottom
		drawPixel(img, x+3, y+5, c)
		drawPixel(img, x+4, y+5, c)
		drawPixel(img, x+3, y+6, c)
		drawPixel(img, x+4, y+6, c)
		// Bottom bar
		for i := 0; i < 5; i++ {
			drawPixel(img, x+i, y+7, c)
			drawPixel(img, x+i, y+8, c)
		}
	case '4':
		// Left vertical
		for i := 0; i < 5; i++ {
			drawPixel(img, x, y+i, c)
			drawPixel(img, x+1, y+i, c)
		}
		// Middle bar
		for i := 0; i < 5; i++ {
			drawPixel(img, x+i, y+4, c)
		}
		// Right vertical
		for i := 0; i < 9; i++ {
			drawPixel(img, x+3, y+i, c)
			drawPixel(img, x+4, y+i, c)
		}
	case '5':
		// Top bar
		for i := 0; i < 5; i++ {
			drawPixel(img, x+i, y, c)
			drawPixel(img, x+i, y+1, c)
		}
		// Left side
		drawPixel(img, x, y+2, c)
		drawPixel(img, x+1, y+2, c)
		drawPixel(img, x, y+3, c)
		drawPixel(img, x+1, y+3, c)
		// Middle bar
		for i := 0; i < 5; i++ {
			drawPixel(img, x+i, y+4, c)
		}
		// Right side
		drawPixel(img, x+3, y+5, c)
		drawPixel(img, x+4, y+5, c)
		drawPixel(img, x+3, y+6, c)
		drawPixel(img, x+4, y+6, c)
		// Bottom bar
		for i := 0; i < 5; i++ {
			drawPixel(img, x+i, y+7, c)
			drawPixel(img, x+i, y+8, c)
		}
	case '6':
		// Top bar
		for i := 0; i < 5; i++ {
			drawPixel(img, x+i, y, c)
			drawPixel(img, x+i, y+1, c)
		}
		// Left side
		for i := 0; i < 9; i++ {
			drawPixel(img, x, y+i, c)
			drawPixel(img, x+1, y+i, c)
		}
		// Middle bar
		for i := 2; i < 5; i++ {
			drawPixel(img, x+i, y+4, c)
		}
		// Right side bottom
		drawPixel(img, x+3, y+5, c)
		drawPixel(img, x+4, y+5, c)
		drawPixel(img, x+3, y+6, c)
		drawPixel(img, x+4, y+6, c)
		// Bottom bar
		for i := 1; i < 4; i++ {
			drawPixel(img, x+i, y+7, c)
			drawPixel(img, x+i, y+8, c)
		}
	case '7':
		// Top bar
		for i := 0; i < 5; i++ {
			drawPixel(img, x+i, y, c)
			drawPixel(img, x+i, y+1, c)
		}
		// Diagonal
		drawPixel(img, x+4, y+2, c)
		drawPixel(img, x+3, y+3, c)
		drawPixel(img, x+3, y+4, c)
		drawPixel(img, x+2, y+5, c)
		drawPixel(img, x+2, y+6, c)
		drawPixel(img, x+1, y+7, c)
		drawPixel(img, x+1, y+8, c)
	case '8':
		// Top bar
		for i := 1; i < 4; i++ {
			drawPixel(img, x+i, y, c)
			drawPixel(img, x+i, y+1, c)
		}
		// Top sides
		drawPixel(img, x, y+1, c)
		drawPixel(img, x+1, y+1, c)
		drawPixel(img, x+3, y+1, c)
		drawPixel(img, x+4, y+1, c)
		drawPixel(img, x, y+2, c)
		drawPixel(img, x+1, y+2, c)
		drawPixel(img, x+3, y+2, c)
		drawPixel(img, x+4, y+2, c)
		drawPixel(img, x, y+3, c)
		drawPixel(img, x+1, y+3, c)
		drawPixel(img, x+3, y+3, c)
		drawPixel(img, x+4, y+3, c)
		// Middle bar
		for i := 1; i < 4; i++ {
			drawPixel(img, x+i, y+4, c)
		}
		// Bottom sides
		drawPixel(img, x, y+5, c)
		drawPixel(img, x+1, y+5, c)
		drawPixel(img, x+3, y+5, c)
		drawPixel(img, x+4, y+5, c)
		drawPixel(img, x, y+6, c)
		drawPixel(img, x+1, y+6, c)
		drawPixel(img, x+3, y+6, c)
		drawPixel(img, x+4, y+6, c)
		// Bottom bar
		for i := 1; i < 4; i++ {
			drawPixel(img, x+i, y+7, c)
			drawPixel(img, x+i, y+8, c)
		}
		drawPixel(img, x, y+7, c)
		drawPixel(img, x+4, y+7, c)
	case '9':
		// Top bar
		for i := 1; i < 4; i++ {
			drawPixel(img, x+i, y, c)
			drawPixel(img, x+i, y+1, c)
		}
		// Top sides
		drawPixel(img, x, y+1, c)
		drawPixel(img, x+1, y+1, c)
		drawPixel(img, x+3, y+1, c)
		drawPixel(img, x+4, y+1, c)
		drawPixel(img, x, y+2, c)
		drawPixel(img, x+1, y+2, c)
		drawPixel(img, x+3, y+2, c)
		drawPixel(img, x+4, y+2, c)
		drawPixel(img, x, y+3, c)
		drawPixel(img, x+1, y+3, c)
		drawPixel(img, x+3, y+3, c)
		drawPixel(img, x+4, y+3, c)
		// Middle bar
		for i := 1; i < 5; i++ {
			drawPixel(img, x+i, y+4, c)
		}
		// Right side
		drawPixel(img, x+3, y+5, c)
		drawPixel(img, x+4, y+5, c)
		drawPixel(img, x+3, y+6, c)
		drawPixel(img, x+4, y+6, c)
		// Bottom bar
		for i := 0; i < 5; i++ {
			drawPixel(img, x+i, y+7, c)
			drawPixel(img, x+i, y+8, c)
		}
	}
}
