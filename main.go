package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"unsafe"

	"github.com/getlantern/systray"
)

var (
	user32                     = syscall.NewLazyDLL("user32.dll")
	procSetWindowsHookEx       = user32.NewProc("SetWindowsHookExW")
	procCallNextHookEx         = user32.NewProc("CallNextHookEx")
	procUnhookWindowsHookEx    = user32.NewProc("UnhookWindowsHookEx")
	procGetMessage             = user32.NewProc("GetMessageW")
	procGetForegroundWindow    = user32.NewProc("GetForegroundWindow")
	procPostMessage            = user32.NewProc("PostMessageW")
	procGetKeyboardLayout      = user32.NewProc("GetKeyboardLayout")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
)

const (
	WH_KEYBOARD_LL = 13
	WM_KEYDOWN     = 0x0100
	WM_KEYUP       = 0x0101
	WM_SYSKEYDOWN  = 0x0104
	WM_SYSKEYUP    = 0x0105
	VK_MENU        = 0x12 // Alt key
	VK_LMENU       = 0xA4 // Left Alt
	VK_RMENU       = 0xA5 // Right Alt
	WM_INPUTLANGCHANGEREQUEST = 0x0050
	INPUTLANGCHANGE_FORWARD   = 0x0002
)

type KBDLLHOOKSTRUCT struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

type MSG struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

var (
	keyboardHook uintptr
	altPressed   bool
	otherKeyPressed bool
)

func lowLevelKeyboardProc(nCode int, wParam uintptr, lParam uintptr) uintptr {
	if nCode >= 0 {
		kbStruct := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))
		vkCode := kbStruct.VkCode

		switch wParam {
		case WM_KEYDOWN, WM_SYSKEYDOWN:
			if vkCode == VK_MENU || vkCode == VK_LMENU || vkCode == VK_RMENU {
				if !altPressed {
					altPressed = true
					otherKeyPressed = false
				}
			} else if altPressed {
				otherKeyPressed = true
			}
		case WM_KEYUP, WM_SYSKEYUP:
			if vkCode == VK_MENU || vkCode == VK_LMENU || vkCode == VK_RMENU {
				if altPressed && !otherKeyPressed {
					switchLanguage()
				}
				altPressed = false
				otherKeyPressed = false
			}
		}
	}

	ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
	return ret
}

func switchLanguage() {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd != 0 {
		procPostMessage.Call(
			hwnd,
			WM_INPUTLANGCHANGEREQUEST,
			uintptr(INPUTLANGCHANGE_FORWARD),
			0,
		)
	}
}

func getCurrentLanguage() string {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return "??"
	}

	threadId, _, _ := procGetWindowThreadProcessId.Call(hwnd, 0)
	hkl, _, _ := procGetKeyboardLayout.Call(threadId)

	langId := uint16(hkl & 0xFFFF)

	switch langId {
	case 0x0409:
		return "EN"
	case 0x0422:
		return "UK"
	case 0x0419:
		return "RU"
	case 0x0415:
		return "PL"
	case 0x0407:
		return "DE"
	case 0x040C:
		return "FR"
	default:
		return fmt.Sprintf("%02X", langId)
	}
}

func setKeyboardHook() error {
	callback := syscall.NewCallback(lowLevelKeyboardProc)

	ret, _, err := procSetWindowsHookEx.Call(
		WH_KEYBOARD_LL,
		callback,
		0,
		0,
	)

	if ret == 0 {
		return fmt.Errorf("SetWindowsHookEx failed: %v", err)
	}

	keyboardHook = ret
	return nil
}

func unhookKeyboard() {
	if keyboardHook != 0 {
		procUnhookWindowsHookEx.Call(keyboardHook)
		keyboardHook = 0
	}
}

func messageLoop() {
	var msg MSG
	for {
		ret, _, _ := procGetMessage.Call(
			uintptr(unsafe.Pointer(&msg)),
			0,
			0,
			0,
		)
		if ret == 0 || int32(ret) == -1 {
			break
		}
	}
}

func onReady() {
	systray.SetIcon(getIcon())
	systray.SetTitle("Lang Switcher")
	systray.SetTooltip("Alt для перемикання мови")

	mStatus := systray.AddMenuItem("Статус: Активний", "")
	mStatus.Disable()

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Вийти", "Закрити програму")

	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()

	// Start keyboard hook
	err := setKeyboardHook()
	if err != nil {
		fmt.Printf("Error setting hook: %v\n", err)
		systray.Quit()
		return
	}

	// Handle system signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		systray.Quit()
	}()

	// Run message loop
	go messageLoop()
}

func onExit() {
	unhookKeyboard()
}

func getIcon() []byte {
	// 16x16 ICO file with 32-bit color depth
	// ICO header
	icon := []byte{
		0x00, 0x00, // Reserved
		0x01, 0x00, // ICO type
		0x01, 0x00, // Number of images
		// ICONDIRENTRY
		0x10,       // Width (16)
		0x10,       // Height (16)
		0x00,       // Color palette
		0x00,       // Reserved
		0x01, 0x00, // Color planes
		0x20, 0x00, // Bits per pixel (32)
		0x68, 0x04, 0x00, 0x00, // Size of image data
		0x16, 0x00, 0x00, 0x00, // Offset to image data
	}

	// BITMAPINFOHEADER
	bmpHeader := []byte{
		0x28, 0x00, 0x00, 0x00, // Header size (40)
		0x10, 0x00, 0x00, 0x00, // Width (16)
		0x20, 0x00, 0x00, 0x00, // Height (32, doubled for mask)
		0x01, 0x00, // Planes
		0x20, 0x00, // Bits per pixel (32)
		0x00, 0x00, 0x00, 0x00, // Compression
		0x00, 0x04, 0x00, 0x00, // Image size
		0x00, 0x00, 0x00, 0x00, // X pixels per meter
		0x00, 0x00, 0x00, 0x00, // Y pixels per meter
		0x00, 0x00, 0x00, 0x00, // Colors used
		0x00, 0x00, 0x00, 0x00, // Important colors
	}
	icon = append(icon, bmpHeader...)

	// Pixel data (BGRA, bottom-up)
	// Create a simple "A" icon with blue background
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			// Blue background with white "A" letter pattern
			isLetter := false
			row := 15 - y // Flip because BMP is bottom-up

			// Simple "A" pattern
			if row >= 2 && row <= 13 {
				if row == 2 || row == 3 { // Top of A
					if x >= 6 && x <= 9 {
						isLetter = true
					}
				} else if row == 7 || row == 8 { // Middle bar
					if x >= 4 && x <= 11 {
						isLetter = true
					}
				} else { // Legs of A
					if (x >= 3 && x <= 5) || (x >= 10 && x <= 12) {
						isLetter = true
					}
				}
			}

			if isLetter {
				icon = append(icon, 0xFF, 0xFF, 0xFF, 0xFF) // White (BGRA)
			} else {
				icon = append(icon, 0xCC, 0x66, 0x00, 0xFF) // Blue (BGRA)
			}
		}
	}

	// AND mask (transparency mask) - all zeros for fully opaque
	for i := 0; i < 64; i++ {
		icon = append(icon, 0x00)
	}

	return icon
}

func main() {
	systray.Run(onReady, onExit)
}
