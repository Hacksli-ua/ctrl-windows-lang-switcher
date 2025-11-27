//go:build ignore

package main

import (
	"os"
)

func main() {
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

	os.WriteFile("app.ico", icon, 0644)
}
