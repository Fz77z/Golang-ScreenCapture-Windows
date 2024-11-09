package main

import (
	"fmt"
	"screenshotter/capture"
)

func main() {
    fmt.Println("Select a region to capture...")

    region, err := capture.SelectRegion()
    if err != nil {
        fmt.Println("Error selecting region:", err)
        return
    }

    fmt.Printf("Selected region: %+v\n", region)

    img, err := capture.CaptureRegion(region)
    if err != nil {
        fmt.Println("Error capturing region:", err)
        return
    }

    err = capture.SaveImage(img, "region_screenshot.png")
    if err != nil {
        fmt.Println("Error saving screenshot:", err)
        return
    }

    fmt.Println("Screenshot saved as region_screenshot.png")
}
