package capture

import "syscall"

// DLL Imports
var (
	user32 = syscall.NewLazyDLL("user32.dll")
	gdi32  = syscall.NewLazyDLL("gdi32.dll")
)

type PAINTSTRUCT struct {
	HDC         uintptr
	Erase       int32
	RcPaint     Rect
	Restore     int32
	IncUpdate   int32
	RgbReserved [32]byte
}


// Procedure Declarations
var (
	getDC               = user32.NewProc("GetDC")
	releaseDC           = user32.NewProc("ReleaseDC")
	createCompatibleDC  = gdi32.NewProc("CreateCompatibleDC")
	createCompatibleBMP = gdi32.NewProc("CreateCompatibleBitmap")
	selectObject        = gdi32.NewProc("SelectObject")
	bitBlt              = gdi32.NewProc("BitBlt")
	deleteDC            = gdi32.NewProc("DeleteDC")
	deleteObject        = gdi32.NewProc("DeleteObject")
	getCursorPos        = user32.NewProc("GetCursorPos")
	registerClassEx     = user32.NewProc("RegisterClassExW")
	createWindowEx      = user32.NewProc("CreateWindowExW")
	showWindow          = user32.NewProc("ShowWindow")
	updateWindow        = user32.NewProc("UpdateWindow")
	dispatchMessage     = user32.NewProc("DispatchMessageW")
	getMessage          = user32.NewProc("GetMessageW")
	destroyWindow       = user32.NewProc("DestroyWindow")
	defWindowProc       = user32.NewProc("DefWindowProcW")
	getSystemMetrics    = user32.NewProc("GetSystemMetrics")
	LoadCursor      	= user32.NewProc("LoadCursorW")
    SetCursor        	= user32.NewProc("SetCursor")
	invalidateRect 		= user32.NewProc("InvalidateRect")
    beginPaint     		= user32.NewProc("BeginPaint")
    endPaint       		= user32.NewProc("EndPaint")
    createSolidBrush 	= gdi32.NewProc("CreateSolidBrush")
	rectangle			= gdi32.NewProc("Rectangle")
	createPen           = gdi32.NewProc("CreatePen")
	moveToEx            = gdi32.NewProc("MoveToEx")
	lineTo              = gdi32.NewProc("LineTo")
	setROP2             = gdi32.NewProc("SetROP2")
	setLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes") // Declare SetLayeredWindowAttributes

	
)
