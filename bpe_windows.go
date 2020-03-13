package main

import (
	"log"
	"syscall"
	"unsafe"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	pGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
)

func getModuleHandle() (syscall.Handle, error) {
	ret, _, err := pGetModuleHandleW.Call(uintptr(0))
	if ret == 0 {
		return 0, err
	}
	return syscall.Handle(ret), nil
}

var (
	user32 = syscall.NewLazyDLL("user32.dll")

	pCreateWindowExW  = user32.NewProc("CreateWindowExW")
	pDefWindowProcW   = user32.NewProc("DefWindowProcW")
	pDestroyWindow    = user32.NewProc("DestroyWindow")
	pDispatchMessageW = user32.NewProc("DispatchMessageW")
	pPeekMessageW     = user32.NewProc("PeekMessageW")
	pPostQuitMessage  = user32.NewProc("PostQuitMessage")
	pRegisterClassExW = user32.NewProc("RegisterClassExW")
	pTranslateMessage = user32.NewProc("TranslateMessage")
)

func createWindow(className, windowName string, style uint32, x, y, width, height uint32, parent, menu, instance syscall.Handle) (syscall.Handle, error) {
	ret, _, err := pCreateWindowExW.Call(
		uintptr(0),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(className))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(windowName))),
		uintptr(style),
		uintptr(x),
		uintptr(y),
		uintptr(width),
		uintptr(height),
		uintptr(parent),
		uintptr(menu),
		uintptr(instance),
		uintptr(0),
	)
	if ret == 0 {
		return 0, err
	}
	return syscall.Handle(ret), nil
}

func dispatchMessage(msg *tMSG) {
	pDispatchMessageW.Call(uintptr(unsafe.Pointer(msg)))
}

func peekMessage(msg *tMSG, hwnd syscall.Handle, msgFilterMin, msgFilterMax, removeMsg uint32) (bool, error) {
	ret, _, err := pPeekMessageW.Call(
		uintptr(unsafe.Pointer(msg)),
		uintptr(hwnd),
		uintptr(msgFilterMin),
		uintptr(msgFilterMax),
		uintptr(removeMsg)
	)
	if int32(ret) == -1 {
		return false, err
	}
	return int32(ret) != 0, nil
}

func defWindowProc(hwnd syscall.Handle, msg uint32, wparam, lparam uintptr) uintptr {
	ret, _, _ := pDefWindowProcW.Call(
		uintptr(hwnd),
		uintptr(msg),
		uintptr(wparam),
		uintptr(lparam),
	)
	return uintptr(ret)
}

type tPOINT struct {
	x, y int32
}

type tMSG struct {
	hwnd    syscall.Handle
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      tPOINT
}

type tWNDCLASSEXW struct {
	size       uint32
	style      uint32
	wndProc    uintptr
	clsExtra   int32
	wndExtra   int32
	instance   syscall.Handle
	icon       syscall.Handle
	cursor     syscall.Handle
	background syscall.Handle
	menuName   *uint16
	className  *uint16
	iconSm     syscall.Handle
}

const (
	cCS_OWNDC   = 0x0020
        cCS_HREDRAW = 0x0002
        cCS_VREDRAW = 0x0001
)

const (
	cPM_REMOVE  = 0x0001
)

func InitWindow()  error {
        Running := true

	className := "bpeClass"

	instance, err := getModuleHandle()
	if err != nil {
		log.Println(err)
		return err
	}

	fn := func(hwnd syscall.Handle, msg uint32, wparam, lparam uintptr) uintptr {
		switch msg {
		case cWM_CLOSE:
			Running = false
		case cWM_DESTROY:
			Running = false
		default:
			ret := defWindowProc(hwnd, msg, wparam, lparam)
			return ret
		}
		return 0
	}

	wcx := tWNDCLASSEXW{
		wndProc:    syscall.NewCallback(fn),
		instance:   instance,
		className:  syscall.StringToUTF16Ptr(className),
                style:      cCS_OWNDC | cCS_HREDRAW | cCS_VREDRAW
	}
	wcx.size = uint32(unsafe.Sizeof(wcx))

	if _, err = registerClassEx(&wcx); err != nil {
		log.Println(err)
		return err
	}

	_, err = createWindow(
		className,
		"BPE Window",
		(cWS_VISIBLE | cWS_OVERLAPPEDWINDOW),
		cSW_USE_DEFAULT,
		cSW_USE_DEFAULT,
		cSW_USE_DEFAULT,
		cSW_USE_DEFAULT,
		0,
		0,
		instance
	)
	if err != nil {
		log.Println(err)
		return err
	}

	for Running {
		msg := tMSG{}
		gotMessage, err := peekMessage(&msg, 0, 0, 0, cPM_REMOVE)
		if err != nil {
			log.Println(err)
			return err
		}

		if gotMessage {
			translateMessage(&msg)
			dispatchMessage(&msg)
		}
	}
}
