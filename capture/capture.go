package capture

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"unsafe"
)



func CaptureScreen() (*image.RGBA, error) {
	// Screen dimensions (replace with actual screen size or dynamic retrieval)
	screenWidth := 1920
	screenHeight := 1080

	// Get the device context for the screen
	hdcScreen, _, _ := getDC.Call(0)
	if hdcScreen == 0 {
		return nil, fmt.Errorf("failed to get screen device context")
	}
	defer releaseDC.Call(0, hdcScreen)

	// Create a compatible device context and bitmap
	hdcMem, _, _ := createCompatibleDC.Call(hdcScreen)
	if hdcMem == 0 {
		return nil, fmt.Errorf("failed to create memory device context")
	}
	defer deleteDC.Call(hdcMem)

	hBitmap, _, _ := createCompatibleBMP.Call(hdcScreen, uintptr(screenWidth), uintptr(screenHeight))
	if hBitmap == 0 {
		return nil, fmt.Errorf("failed to create compatible bitmap")
	}
	defer deleteObject.Call(hBitmap)

	// Select the bitmap into the memory device context
	selectObject.Call(hdcMem, hBitmap)

	// Copy the screen content to the memory device context
	if ret, _, _ := bitBlt.Call(hdcMem, 0, 0, uintptr(screenWidth), uintptr(screenHeight), hdcScreen, 0, 0, 0x00CC0020); ret == 0 {
		return nil, fmt.Errorf("failed to copy screen content")
	}

	// Create an image buffer
	img := image.NewRGBA(image.Rect(0, 0, screenWidth, screenHeight))

	// Retrieve the pixel data from the bitmap
	bitmapData, err := getBitmapData(hdcMem, hBitmap, screenWidth, screenHeight)
	if err != nil {
		return nil, err
	}

	// Convert the raw pixel data to RGBA format
	for y := 0; y < screenHeight; y++ {
		for x := 0; x < screenWidth; x++ {
			index := (y*screenWidth + x) * 4
			b := bitmapData[index]
			g := bitmapData[index+1]
			r := bitmapData[index+2]
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	return img, nil
}

func CaptureRegion(region Rect) (*image.RGBA, error) {
    // Calculate region dimensions
    screenWidth := int(region.Right - region.Left)
    screenHeight := int(region.Bottom - region.Top)

    if screenWidth <= 0 || screenHeight <= 0 {
        return nil, fmt.Errorf("invalid region dimensions: width=%d, height=%d", screenWidth, screenHeight)
    }

    // Get the device context for the screen
    hdcScreen, _, _ := getDC.Call(0)
    if hdcScreen == 0 {
        return nil, fmt.Errorf("failed to get screen device context")
    }
    defer releaseDC.Call(0, hdcScreen)

    // Create a compatible device context and bitmap
    hdcMem, _, _ := createCompatibleDC.Call(hdcScreen)
    if hdcMem == 0 {
        return nil, fmt.Errorf("failed to create memory device context")
    }
    defer deleteDC.Call(hdcMem)

    hBitmap, _, _ := createCompatibleBMP.Call(hdcScreen, uintptr(screenWidth), uintptr(screenHeight))
    if hBitmap == 0 {
        return nil, fmt.Errorf("failed to create compatible bitmap")
    }
    defer deleteObject.Call(hBitmap)

    // Select the bitmap into the memory device context
    selectObject.Call(hdcMem, hBitmap)

    // Copy the selected region content to the memory device context
    if ret, _, _ := bitBlt.Call(hdcMem, 0, 0, uintptr(screenWidth), uintptr(screenHeight), hdcScreen, uintptr(region.Left), uintptr(region.Top), 0x00CC0020); ret == 0 {
        return nil, fmt.Errorf("failed to copy screen content")
    }

    // Retrieve the pixel data from the bitmap
    bitmapData, err := getBitmapData(hdcMem, hBitmap, screenWidth, screenHeight)
    if err != nil {
        return nil, err
    }

    if len(bitmapData) == 0 {
        return nil, fmt.Errorf("bitmap data is empty")
    }

    // Convert the raw pixel data to an image
    img := image.NewRGBA(image.Rect(0, 0, screenWidth, screenHeight))
    for y := 0; y < screenHeight; y++ {
        for x := 0; x < screenWidth; x++ {
            index := (y*screenWidth + x) * 4
            b := bitmapData[index]
            g := bitmapData[index+1]
            r := bitmapData[index+2]
            img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
        }
    }

    return img, nil
}


func getBitmapData(hdcMem uintptr, hBitmap uintptr, width, height int) ([]byte, error) {
	// Define a BITMAPINFOHEADER structure
	type BITMAPINFOHEADER struct {
		Size          uint32
		Width         int32
		Height        int32
		Planes        uint16
		BitCount      uint16
		Compression   uint32
		SizeImage     uint32
		XPelsPerMeter int32
		YPelsPerMeter int32
		ClrUsed       uint32
		ClrImportant  uint32
	}
	header := BITMAPINFOHEADER{
		Size:     uint32(unsafe.Sizeof(BITMAPINFOHEADER{})),
		Width:    int32(width),
		Height:   -int32(height), // Negative to indicate top-down bitmap
		Planes:   1,
		BitCount: 32,
	}

	// Allocate buffer for pixel data
	bufSize := width * height * 4 // 4 bytes per pixel (BGRA)
	buf := make([]byte, bufSize)

	// Use GetDIBits to retrieve the bitmap data
	getDIBits := gdi32.NewProc("GetDIBits")
	ret, _, _ := getDIBits.Call(
		hdcMem,
		hBitmap,
		0,
		uintptr(height),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&header)),
		0,
	)
	if ret == 0 {
		return nil, fmt.Errorf("failed to get bitmap data")
	}

	return buf, nil
}

func SaveImage(img *image.RGBA, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, img)
}