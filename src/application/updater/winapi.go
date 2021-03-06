package main

import (
	"fmt"
	`log`
	`os`
	`os/exec`
	`path/filepath`
	"syscall"
	"unsafe"
)

var procBox HWND
var procBar HWND
var procBtn HWND
var procText HWND
var hicon HICON

const (
	NullPtr = uintptr(0)
)
const (
	IDB_OK   = 40000
	IDP_PROC = 40001
	IDT_STAT = 40002
)
const (
	WM_INITDIALOG  = 0x0110
	WM_COMMAND     = 0x0111
	WM_SYSCOMMAND  = 0x0112
	WM_CLOSE       = 0x0010
	WM_SETTEXT     = 0x000C
	WM_ENABLE      = 0x000A
	WM_USER        = 1024 //0x0400
	PBM_SETRANGE   = (WM_USER + 1)
	PBM_SETPOS     = (WM_USER + 2)
	PBM_DELTAPOS   = (WM_USER + 3)
	PBM_SETSTEP    = (WM_USER + 4)
	PBM_STEPIT     = (WM_USER + 5)
	PBM_SETRANGE32 = (WM_USER + 6)
	PBS_MARQUEE    = 0x8
	PBM_SETMARQUEE = (WM_USER + 10)
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	hicon = LoadIcon(GetModuleHandle(nil), MakeIntResource(100))
}
func getUint16PtrOfString(str string) *uint16 {
	ptr, _ := syscall.UTF16PtrFromString(str)
	return ptr
}
func ShowDialog(closed chan bool) {
	callback := syscall.NewCallback(func(hwnd HWND, uMsg uint32, wParam uintptr, lParam uintptr) uintptr {
		switch uMsg {
		case WM_INITDIALOG:
			log.Println(SendMessage(procText, WM_SETTEXT, uintptr(0), uintptr(unsafe.Pointer(getUint16PtrOfString("更新程序中...")))))
			log.Println(`procBar marquee`, PostMessage(procBar, PBM_SETMARQUEE, 1, 30))
			//SetWindowText(procText, WM_SETTEXT, uintptr(unsafe.Pointer(getUint16PtrOfString("更新程序中..."))), uintptr(0))
			log.Println(`dialogProc update procBtn`, procBtn, EnableWindow(procBtn, false))
		case WM_COMMAND:
			log.Println(LOWORD(uint32(wParam)))
			switch LOWORD(uint32(wParam)) {
			case IDB_OK:
				DestroyWindow(hwnd)
				cmd := exec.Command(`cmd`,`/c start `+filepath.Join(root(), appname)+` `+appext+``)
				log.Println(`/c start "`+filepath.Join(root(), appname)+`" `+appext+``)
				//cmd := exec.Command(filepath.Join(root(), appname),appext)
				cmd.Run()
				os.Exit(0)
				return uintptr(1)
			default:
				return 0
			}
		case WM_CLOSE:
			closed <- true
			return uintptr(1)
		default:
			return DefWindowProc(hwnd, uMsg, uintptr(wParam), uintptr(lParam))
		}
		return uintptr(0)
	})
	procBox = CreateDialog(GetModuleHandle(nil), MakeIntResource(102), HWND(uintptr(0)), callback)
	procBar = GetDlgItem(procBox, IDP_PROC)
	procBtn = GetDlgItem(procBox, IDB_OK)
	procText = GetDlgItem(procBox, IDT_STAT)
	updateProcess("更新程序中", 0)
	log.Println(`dialog`, procBox, `bar`, procBar, `btn`, procBtn, `text`, procText)
}
func updateProcess(text string, proc int) {
	ret := SendMessage(procText, WM_SETTEXT, uintptr(0), uintptr(unsafe.Pointer(getUint16PtrOfString(text))))
	if ret != 0 {
		log.Println("update text error", procText, GetLastError())
	}
	retb := PostMessage(procBar, PBM_SETPOS, uintptr(proc), uintptr(0))
	if !retb {
		log.Println("update process error", procBar, GetLastError())
	}
}
func messageBox(title, content string, flag uint) int {
	fmt.Println(hicon)
	ptitle, _ := syscall.UTF16PtrFromString(title)
	pcontent, _ := syscall.UTF16PtrFromString(content)
	return MessageBox(HWND(uintptr(0)), pcontent, ptitle, flag)
}

var (
	libkernel32 = syscall.NewLazyDLL("kernel32.dll")

	// Functions

	closeHandle     = libkernel32.NewProc("CloseHandle")
	getLastError    = libkernel32.NewProc("GetLastError")
	getModuleHandle = libkernel32.NewProc("GetModuleHandleW")

	moduser32 = syscall.NewLazyDLL("user32.dll")

	procLoadIcon            = moduser32.NewProc("LoadIconW")
	procLoadCursor          = moduser32.NewProc("LoadCursorW")
	procShowWindow          = moduser32.NewProc("ShowWindow")
	procUpdateWindow        = moduser32.NewProc("UpdateWindow")
	procCreateWindowEx      = moduser32.NewProc("CreateWindowExW")
	procEnableWindow        = moduser32.NewProc("EnableWindow")
	procDestroyWindow       = moduser32.NewProc("DestroyWindow")
	procDefWindowProc       = moduser32.NewProc("DefWindowProcW")
	procDefDlgProc          = moduser32.NewProc("DefDlgProcW")
	procPostQuitMessage     = moduser32.NewProc("PostQuitMessage")
	procMessageBoxIndirectw = moduser32.NewProc("MessageBoxIndirectW")
	pGetMessageW            = moduser32.NewProc("GetMessageW")
	pDispatchMessageW       = moduser32.NewProc("DispatchMessageW")
	pTranslateMessage       = moduser32.NewProc("TranslateMessage")

	procSendMessage         = moduser32.NewProc("SendMessageW")
	procSendMessageTimeout  = moduser32.NewProc("SendMessageTimeout")
	procPostMessage         = moduser32.NewProc("PostMessageW")
	procWaitMessage         = moduser32.NewProc("WaitMessage")
	procSetWindowText       = moduser32.NewProc("SetWindowTextW")
	procGetWindowTextLength = moduser32.NewProc("GetWindowTextLengthW")
	procGetWindowText       = moduser32.NewProc("GetWindowTextW")

	procSetWindowLong    = moduser32.NewProc("SetWindowLongW")
	procSetWindowLongPtr = moduser32.NewProc("SetWindowLongW")
	procGetWindowLong    = moduser32.NewProc("GetWindowLongW")
	procGetWindowLongPtr = moduser32.NewProc("GetWindowLongW")

	procIsWindowEnabled = moduser32.NewProc("IsWindowEnabled")
	procIsWindowVisible = moduser32.NewProc("IsWindowVisible")
	procSetFocus        = moduser32.NewProc("SetFocus")

	procSetCapture               = moduser32.NewProc("SetCapture")
	procReleaseCapture           = moduser32.NewProc("ReleaseCapture")
	procGetWindowThreadProcessId = moduser32.NewProc("GetWindowThreadProcessId")
	procMessageBox               = moduser32.NewProc("MessageBoxW")

	procCreateDialogParam = moduser32.NewProc("CreateDialogParamW")
	procDialogBoxParam    = moduser32.NewProc("DialogBoxParamW")
	procGetDlgItem        = moduser32.NewProc("GetDlgItem")

	procIsWindow  = moduser32.NewProc("IsWindow")
	procEndDialog = moduser32.NewProc("EndDialog")

	procCreateIcon  = moduser32.NewProc("CreateIcon")
	procDestroyIcon = moduser32.NewProc("DestroyIcon")

	procSetForegroundWindow = moduser32.NewProc("SetForegroundWindow")
	procFindWindowW         = moduser32.NewProc("FindWindowW")
	procFindWindowExW       = moduser32.NewProc("FindWindowExW")
	procGetClassName        = moduser32.NewProc("GetClassNameW")
)

const (
	MB_OK                = 0x00000000
	MB_OKCANCEL          = 0x00000001
	MB_ABORTRETRYIGNORE  = 0x00000002
	MB_YESNOCANCEL       = 0x00000003
	MB_YESNO             = 0x00000004
	MB_RETRYCANCEL       = 0x00000005
	MB_CANCELTRYCONTINUE = 0x00000006
	MB_ICONHAND          = 0x00000010
	MB_ICONQUESTION      = 0x00000020
	MB_ICONEXCLAMATION   = 0x00000030
	MB_ICONASTERISK      = 0x00000040
	MB_USERICON          = 0x00000080
	MB_ICONWARNING       = MB_ICONEXCLAMATION
	MB_ICONERROR         = MB_ICONHAND
	MB_ICONINFORMATION   = MB_ICONASTERISK
	MB_ICONSTOP          = MB_ICONHAND
	MB_DEFBUTTON1        = 0x00000000
	MB_DEFBUTTON2        = 0x00000100
	MB_DEFBUTTON3        = 0x00000200
	MB_DEFBUTTON4        = 0x00000300
	MB_TOPMOST           = 0x00040000
	MB_SETFOREGROUND     = 0x00010000
	MB_SYSTEMMODAL       = 0x00001000
)
const (
	IDOK       = 1
	IDCANCEL   = 2
	IDABORT    = 3
	IDRETRY    = 4
	IDIGNORE   = 5
	IDYES      = 6
	IDNO       = 7
	IDCLOSE    = 8
	IDHELP     = 9
	IDTRYAGAIN = 10
	IDCONTINUE = 11
	IDTIMEOUT  = 32000
)

type (
	ATOM uint16
	BOOL int32
	COLORREF uint32
	DWM_FRAME_COUNT uint64
	DWORD uint32
	HACCEL HANDLE
	HANDLE uintptr
	HBITMAP HANDLE
	HBRUSH HANDLE
	HCURSOR HANDLE
	HDC HANDLE
	HDROP HANDLE
	HDWP HANDLE
	HENHMETAFILE HANDLE
	HFONT HANDLE
	HGDIOBJ HANDLE
	HGLOBAL HANDLE
	HGLRC HANDLE
	HHOOK HANDLE
	HICON HANDLE
	HIMAGELIST HANDLE
	HINSTANCE HANDLE
	HKEY HANDLE
	HKL HANDLE
	HMENU HANDLE
	HMODULE HANDLE
	HMONITOR HANDLE
	HPEN HANDLE
	HRESULT int32
	HRGN HANDLE
	HRSRC HANDLE
	HTHUMBNAIL HANDLE
	HWND HANDLE
	LPARAM uintptr
	LPCVOID unsafe.Pointer
	LRESULT uintptr
	PVOID unsafe.Pointer
	QPC_TIME uint64
	ULONG_PTR uintptr
	WPARAM uintptr
	TRACEHANDLE uintptr
)

func CloseHandle(hObject HANDLE) bool {
	ret, _, _ := closeHandle.Call(uintptr(hObject))
	return ret != 0
}
func GetLastError() uint32 {
	ret, _, _ := getLastError.Call()
	return uint32(ret)
}
func GetModuleHandle(lpModuleName *uint16) HINSTANCE {
	ret, _, _ := getModuleHandle.Call(uintptr(unsafe.Pointer(lpModuleName)))
	return HINSTANCE(ret)
}
func EnableWindow(hwnd HWND, enable bool) bool {
	BoolToBOOL := func(value bool) BOOL {
		if value {
			return 1
		}
		return 0
	}
	ret, _, _ := procEnableWindow.Call(uintptr(hwnd), uintptr(BoolToBOOL(enable)))
	return ret != 0
}
func LoadIcon(instance HINSTANCE, iconName *uint16) HICON {
	ret, _, _ := procLoadIcon.Call(
		uintptr(instance),
		uintptr(unsafe.Pointer(iconName)))

	return HICON(ret)

}

type tMSG struct {
	hwnd    syscall.Handle
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      tPOINT
}
type tPOINT struct {
	x, y int32
}

func dispatchMessage(msg *tMSG) {
	pDispatchMessageW.Call(uintptr(unsafe.Pointer(msg)))
}
func translateMessage(msg *tMSG) {
	pTranslateMessage.Call(uintptr(unsafe.Pointer(msg)))
}
func getMessage(msg *tMSG, hwnd syscall.Handle, msgFilterMin, msgFilterMax uint32) (bool, error) {
	ret, _, err := pGetMessageW.Call(
		uintptr(unsafe.Pointer(msg)),
		uintptr(hwnd),
		uintptr(msgFilterMin),
		uintptr(msgFilterMax),
	)
	if int32(ret) == -1 {
		return false, err
	}
	return int32(ret) != 0, nil
}

func LoadCursor(instance HINSTANCE, cursorName *uint16) HCURSOR {
	ret, _, _ := procLoadCursor.Call(
		uintptr(instance),
		uintptr(unsafe.Pointer(cursorName)))

	return HCURSOR(ret)

}

func GetClassNameW(hwnd HWND) string {
	buf := make([]uint16, 255)
	procGetClassName.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(255))

	return syscall.UTF16ToString(buf)
}

func SetForegroundWindow(hwnd HWND) bool {
	ret, _, _ := procSetForegroundWindow.Call(
		uintptr(hwnd))

	return ret != 0
}

func ShowWindow(hwnd HWND, cmdshow int) bool {
	ret, _, _ := procShowWindow.Call(
		uintptr(hwnd),
		uintptr(cmdshow))

	return ret != 0

}

func UpdateWindow(hwnd HWND) bool {
	ret, _, _ := procUpdateWindow.Call(
		uintptr(hwnd))
	return ret != 0
}

func CreateWindowEx(exStyle uint, className, windowName *uint16,
	style uint, x, y, width, height int, parent HWND, menu HMENU,
	instance HINSTANCE, param unsafe.Pointer) HWND {
	ret, _, _ := procCreateWindowEx.Call(
		uintptr(exStyle),
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowName)),
		uintptr(style),
		uintptr(x),
		uintptr(y),
		uintptr(width),
		uintptr(height),
		uintptr(parent),
		uintptr(menu),
		uintptr(instance),
		uintptr(param))

	return HWND(ret)
}

func FindWindowExW(hwndParent, hwndChildAfter HWND, className, windowName *uint16) HWND {
	ret, _, _ := procFindWindowExW.Call(
		uintptr(hwndParent),
		uintptr(hwndChildAfter),
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowName)))

	return HWND(ret)
}

func FindWindowW(className, windowName *uint16) HWND {
	ret, _, _ := procFindWindowW.Call(
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowName)))

	return HWND(ret)
}

func DestroyWindow(hwnd HWND) bool {
	ret, _, _ := procDestroyWindow.Call(
		uintptr(hwnd))

	return ret != 0
}

func DefWindowProc(hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr {
	ret, _, _ := procDefWindowProc.Call(
		uintptr(hwnd),
		uintptr(msg),
		wParam,
		lParam)

	return ret
}

func DefDlgProc(hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr {
	ret, _, _ := procDefDlgProc.Call(
		uintptr(hwnd),
		uintptr(msg),
		wParam,
		lParam)

	return ret
}

func PostQuitMessage(exitCode int) {
	procPostQuitMessage.Call(
		uintptr(exitCode))
}

func SendMessage(hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr {
	ret, _, _ := procSendMessage.Call(
		uintptr(hwnd),
		uintptr(msg),
		wParam,
		lParam)

	return ret
}

func SendMessageTimeout(hwnd HWND, msg uint32, wParam, lParam uintptr, fuFlags, uTimeout uint32) uintptr {
	ret, _, _ := procSendMessageTimeout.Call(
		uintptr(hwnd),
		uintptr(msg),
		wParam,
		lParam,
		uintptr(fuFlags),
		uintptr(uTimeout))

	return ret
}

func PostMessage(hwnd HWND, msg uint32, wParam, lParam uintptr) bool {
	ret, _, _ := procPostMessage.Call(
		uintptr(hwnd),
		uintptr(msg),
		wParam,
		lParam)

	return ret != 0
}

func WaitMessage() bool {
	ret, _, _ := procWaitMessage.Call()
	return ret != 0
}

func SetWindowText(hwnd HWND, text string) {
	procSetWindowText.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))))
}

func GetWindowTextLength(hwnd HWND) int {
	ret, _, _ := procGetWindowTextLength.Call(
		uintptr(hwnd))

	return int(ret)
}

func GetWindowText(hwnd HWND) string {
	textLen := GetWindowTextLength(hwnd) + 1

	buf := make([]uint16, textLen)
	procGetWindowText.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(textLen))

	return syscall.UTF16ToString(buf)
}

func SetWindowLong(hwnd HWND, index int, value uint32) uint32 {
	ret, _, _ := procSetWindowLong.Call(
		uintptr(hwnd),
		uintptr(index),
		uintptr(value))

	return uint32(ret)
}

func SetWindowLongPtr(hwnd HWND, index int, value uintptr) uintptr {
	ret, _, _ := procSetWindowLongPtr.Call(
		uintptr(hwnd),
		uintptr(index),
		value)

	return ret
}

func GetWindowLong(hwnd HWND, index int) int32 {
	ret, _, _ := procGetWindowLong.Call(
		uintptr(hwnd),
		uintptr(index))

	return int32(ret)
}

func GetWindowLongPtr(hwnd HWND, index int) uintptr {
	ret, _, _ := procGetWindowLongPtr.Call(
		uintptr(hwnd),
		uintptr(index))

	return ret
}

func IsWindowEnabled(hwnd HWND) bool {
	ret, _, _ := procIsWindowEnabled.Call(
		uintptr(hwnd))

	return ret != 0
}

func IsWindowVisible(hwnd HWND) bool {
	ret, _, _ := procIsWindowVisible.Call(
		uintptr(hwnd))

	return ret != 0
}

func SetFocus(hwnd HWND) HWND {
	ret, _, _ := procSetFocus.Call(
		uintptr(hwnd))

	return HWND(ret)
}

func SetCapture(hwnd HWND) HWND {
	ret, _, _ := procSetCapture.Call(
		uintptr(hwnd))

	return HWND(ret)
}

func ReleaseCapture() bool {
	ret, _, _ := procReleaseCapture.Call()

	return ret != 0
}

func GetWindowThreadProcessId(hwnd HWND) (HANDLE, int) {
	var processId int
	ret, _, _ := procGetWindowThreadProcessId.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&processId)))

	return HANDLE(ret), processId
}

func MessageBox(hwnd HWND, title, caption *uint16, flags uint) int {
	ret, _, _ := procMessageBox.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(title)),
		uintptr(unsafe.Pointer(caption)),
		uintptr(flags))

	return int(ret)
}

func CreateDialog(hInstance HINSTANCE, lpTemplate *uint16, hWndParent HWND, lpDialogProc uintptr) HWND {
	ret, _, _ := procCreateDialogParam.Call(
		uintptr(hInstance),
		uintptr(unsafe.Pointer(lpTemplate)),
		uintptr(hWndParent),
		lpDialogProc,
		0)

	return HWND(ret)
}

func DialogBox(hInstance HINSTANCE, lpTemplateName *uint16, hWndParent HWND, lpDialogProc uintptr) int {
	ret, _, _ := procDialogBoxParam.Call(
		uintptr(hInstance),
		uintptr(unsafe.Pointer(lpTemplateName)),
		uintptr(hWndParent),
		lpDialogProc,
		0)

	return int(ret)
}

func GetDlgItem(hDlg HWND, nIDDlgItem int) HWND {
	ret, _, _ := procGetDlgItem.Call(
		uintptr(unsafe.Pointer(hDlg)),
		uintptr(nIDDlgItem))

	return HWND(ret)
}

func IsWindow(hwnd HWND) bool {
	ret, _, _ := procIsWindow.Call(
		uintptr(hwnd))

	return ret != 0
}

func EndDialog(hwnd HWND, nResult uintptr) bool {
	ret, _, _ := procEndDialog.Call(
		uintptr(hwnd),
		nResult)

	return ret != 0
}
func CreateIcon(instance HINSTANCE, nWidth, nHeight int, cPlanes, cBitsPerPixel byte, ANDbits, XORbits *byte) HICON {
	ret, _, _ := procCreateIcon.Call(
		uintptr(instance),
		uintptr(nWidth),
		uintptr(nHeight),
		uintptr(cPlanes),
		uintptr(cBitsPerPixel),
		uintptr(unsafe.Pointer(ANDbits)),
		uintptr(unsafe.Pointer(XORbits)),
	)
	return HICON(ret)
}

func DestroyIcon(icon HICON) bool {
	ret, _, _ := procDestroyIcon.Call(
		uintptr(icon),
	)
	return ret != 0
}

func MakeIntResource(id uint16) *uint16 {
	return (*uint16)(unsafe.Pointer(uintptr(id)))
}
func LOWORD(dw uint32) uint16 {
	return uint16(dw)
}

func HIWORD(dw uint32) uint16 {
	return uint16(dw >> 16 & 0xffff)
}

type MSGBOXPARAMSW struct {
	cbSize             uint32
	hwndOwner          HWND
	hInstance          HINSTANCE
	lpszText           *uint16
	lpszCaption        *uint16
	dwStyle            uint
	lpszIcon           *uint16
	dwContextHelpId    int
	lpfnMsgBoxCallback uintptr
	dwLanguageId       DWORD
}

func MessageBoxIndirect(hwnd HWND, title, content *uint16, icon *uint16, uType uint) int {
	param := &MSGBOXPARAMSW{
		hwndOwner:          hwnd,
		hInstance:          GetModuleHandle(nil),
		lpszText:           content,
		lpszCaption:        title,
		dwStyle:            uType,
		lpszIcon:           icon,
		dwContextHelpId:    0,
		lpfnMsgBoxCallback: uintptr(0),
		dwLanguageId:       0,
	}
	param.cbSize = uint32(unsafe.Sizeof(param))
	ret, _, _ := procMessageBoxIndirectw.Call(uintptr(unsafe.Pointer(param)))
	return int(ret)
}
