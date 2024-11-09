package capture

import (
	"fmt"
	"syscall"
	"unsafe"
)

// Constants
const (
	WM_LBUTTONDOWN = 0x0201
	WM_MOUSEMOVE   = 0x0200
	WM_LBUTTONUP   = 0x0202
	WM_PAINT       = 0x000F
	WM_SETCURSOR   = 0x0020
	SW_SHOW        = 5
	WS_POPUP       = 0x80000000
	WS_EX_TOPMOST  = 0x00000008
	WS_EX_LAYERED  = 0x00080000
	SM_CXSCREEN    = 0 // Screen width
	SM_CYSCREEN    = 1 // Screen height
	LWA_ALPHA      = 0x00000002
	R2_XORPEN = 7
    SRCCOPY   = 0x00CC0020
)

type WNDCLASSEX struct {
	Size        uint32
	Style       uint32
	WndProc     uintptr
	ClsExtra    int32
	WndExtra    int32
	Instance    uintptr
	Icon        uintptr
	Cursor      uintptr
	Background  uintptr
	MenuName    *uint16
	ClassName   *uint16
	IconSm      uintptr
}

type MSG struct {
	HWND    uintptr
	Message uint32
	WPARAM  uintptr
	LPARAM  uintptr
	Time    uint32
	PT      Point
}

type Point struct {
	X, Y int32
}

type Rect struct {
	Left, Top, Right, Bottom int32
}

type RegionSelector struct {
	StartPoint, EndPoint Point
	SelectionMade        bool
	LastRect           Rect    // Store the last drawn rectangle
    IsDrawing      bool  // Added this field
}

var regionSelector = &RegionSelector{}

// GetSystemMetrics retrieves screen dimensions
func GetSystemMetrics(index int32) int32 {
	ret, _, _ := getSystemMetrics.Call(uintptr(index))
	return int32(ret)
}

// SelectRegion creates an overlay window and captures the selected region
func SelectRegion() (Rect, error) {
	hInstance := uintptr(0) // Default instance handle

	// Register a window class
	wndClassName := syscall.StringToUTF16Ptr("RegionSelector")
	wndClass := WNDCLASSEX{
		Size:        uint32(unsafe.Sizeof(WNDCLASSEX{})),
		WndProc:     syscall.NewCallback(regionWndProc),
		Instance:    hInstance,
		ClassName:   wndClassName,
		Cursor:      0,
		Background:  0,
	}
	ret, _, err := registerClassEx.Call(uintptr(unsafe.Pointer(&wndClass)))
	if ret == 0 {
		return Rect{}, fmt.Errorf("failed to register window class: %v", err)
	}

	// Create a layered window
	screenWidth := GetSystemMetrics(SM_CXSCREEN)
	screenHeight := GetSystemMetrics(SM_CYSCREEN)
	hWnd, _, _ := createWindowEx.Call(
		uintptr(WS_EX_TOPMOST|WS_EX_LAYERED),
		uintptr(unsafe.Pointer(wndClassName)),
		0,
		uintptr(WS_POPUP),
		0, 0, uintptr(screenWidth), uintptr(screenHeight),
		0, 0, hInstance, 0)
	if hWnd == 0 {
		return Rect{}, fmt.Errorf("failed to create overlay window")
	}

	// Set semi-transparency for the entire window
	setLayeredWindowAttributes.Call(hWnd, 0, 16, uintptr(LWA_ALPHA)) // Very transparent overlay

	showWindow.Call(hWnd, SW_SHOW)
	updateWindow.Call(hWnd)

	// Message loop
	var msg MSG
	for {
		ret, _, _ := getMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if ret == 0 || regionSelector.SelectionMade {
			break
		}
		dispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}

	destroyWindow.Call(hWnd)

	// Calculate and validate the selected region
	selectedRegion := Rect{
		Left:   min(regionSelector.StartPoint.X, regionSelector.EndPoint.X),
		Top:    min(regionSelector.StartPoint.Y, regionSelector.EndPoint.Y),
		Right:  max(regionSelector.StartPoint.X, regionSelector.EndPoint.X),
		Bottom: max(regionSelector.StartPoint.Y, regionSelector.EndPoint.Y),
	}

	width := selectedRegion.Right - selectedRegion.Left
	height := selectedRegion.Bottom - selectedRegion.Top
	if width <= 0 || height <= 0 {
		fmt.Println("Selected region is invalid. Please try again.")
		return Rect{}, fmt.Errorf("invalid region dimensions: width=%d, height=%d", width, height)
	}

	return selectedRegion, nil
}

func regionWndProc(hWnd, msg, wParam, lParam uintptr) uintptr {
    switch msg {
    case WM_SETCURSOR:
        crosshairCursor, _, _ := LoadCursor.Call(0, uintptr(32515))
        SetCursor.Call(crosshairCursor)
        return 1

    case WM_LBUTTONDOWN:
        var pt Point
        getCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
        regionSelector.StartPoint = pt
        regionSelector.EndPoint = pt
        regionSelector.IsDrawing = true
        regionSelector.LastRect = Rect{}
        return 0

	case WM_MOUSEMOVE:
    if wParam == 1 && regionSelector.IsDrawing {
        var pt Point
        getCursorPos.Call(uintptr(unsafe.Pointer(&pt)))

        // Only redraw if the point has changed significantly
        if abs(pt.X-regionSelector.EndPoint.X) > 1 || abs(pt.Y-regionSelector.EndPoint.Y) > 1 {
            hdc, _, _ := getDC.Call(hWnd)
            if hdc != 0 {
                // Set XOR drawing mode
                setROP2.Call(hdc, R2_XORPEN)
                
                // Create thicker pen (increased from 1 to 3)
                pen, _, _ := createPen.Call(0, 5, 0xFF0000)
                oldPen, _, _ := selectObject.Call(hdc, pen)

                // Erase previous rectangle
                if regionSelector.LastRect != (Rect{}) {
                    moveToEx.Call(
                        hdc,
                        uintptr(regionSelector.LastRect.Left),
                        uintptr(regionSelector.LastRect.Top),
                        0,
                    )
                    // Draw the lines individually
                    lineTo.Call(hdc, uintptr(regionSelector.LastRect.Right), uintptr(regionSelector.LastRect.Top))
                    lineTo.Call(hdc, uintptr(regionSelector.LastRect.Right), uintptr(regionSelector.LastRect.Bottom))
                    lineTo.Call(hdc, uintptr(regionSelector.LastRect.Left), uintptr(regionSelector.LastRect.Bottom))
                    lineTo.Call(hdc, uintptr(regionSelector.LastRect.Left), uintptr(regionSelector.LastRect.Top))
                }

                // Calculate new rectangle
                newRect := Rect{
                    Left:   min(regionSelector.StartPoint.X, pt.X),
                    Top:    min(regionSelector.StartPoint.Y, pt.Y),
                    Right:  max(regionSelector.StartPoint.X, pt.X),
                    Bottom: max(regionSelector.StartPoint.Y, pt.Y),
                }

                // Draw new rectangle using lines
                moveToEx.Call(
                    hdc,
                    uintptr(newRect.Left),
                    uintptr(newRect.Top),
                    0,
                )
                lineTo.Call(hdc, uintptr(newRect.Right), uintptr(newRect.Top))
                lineTo.Call(hdc, uintptr(newRect.Right), uintptr(newRect.Bottom))
                lineTo.Call(hdc, uintptr(newRect.Left), uintptr(newRect.Bottom))
                lineTo.Call(hdc, uintptr(newRect.Left), uintptr(newRect.Top))

                // Store the new rectangle
                regionSelector.LastRect = newRect

                // Cleanup
                selectObject.Call(hdc, oldPen)
                deleteObject.Call(pen)
                releaseDC.Call(hWnd, hdc)
            }
            
            regionSelector.EndPoint = pt
        }
    }
    return 0

	
	case WM_LBUTTONUP:
        if regionSelector.IsDrawing {
            var pt Point
            getCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
            regionSelector.EndPoint = pt
            regionSelector.SelectionMade = true
            regionSelector.IsDrawing = false
        }
        return 0

	case WM_PAINT:
			var ps PAINTSTRUCT
			beginPaint.Call(hWnd, uintptr(unsafe.Pointer(&ps)))
			endPaint.Call(hWnd, uintptr(unsafe.Pointer(&ps)))
			return 0
    }

    r1, _, _ := defWindowProc.Call(hWnd, msg, wParam, lParam)
    return r1
}

func abs(x int32) int32 {
    if x < 0 {
        return -x
    }
    return x
}

// Utility functions
func min(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}
