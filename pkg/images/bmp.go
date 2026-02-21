package images

import (
	"encoding/binary"
	"os"
)

func SaveBMP(filename string, data []byte, width, height int) error {
	// BMP files require rows to be padded to 4-byte boundaries
	rowSize := (width*3 + 3) &^ 3
	pixelDataSize := rowSize * height

	// BMP file header (14 bytes) + DIB header (40 bytes) = 54 bytes
	fileSize := 54 + pixelDataSize

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// BMP File Header (14 bytes)
	file.Write([]byte{'B', 'M'})                              // Signature
	binary.Write(file, binary.LittleEndian, uint32(fileSize)) // File size
	binary.Write(file, binary.LittleEndian, uint16(0))        // Reserved
	binary.Write(file, binary.LittleEndian, uint16(0))        // Reserved
	binary.Write(file, binary.LittleEndian, uint32(54))       // Pixel data offset

	// DIB Header (BITMAPINFOHEADER - 40 bytes)
	binary.Write(file, binary.LittleEndian, uint32(40))            // DIB header size
	binary.Write(file, binary.LittleEndian, int32(width))          // Width
	binary.Write(file, binary.LittleEndian, int32(height))         // Height (positive = bottom-up)
	binary.Write(file, binary.LittleEndian, uint16(1))             // Color planes
	binary.Write(file, binary.LittleEndian, uint16(24))            // Bits per pixel
	binary.Write(file, binary.LittleEndian, uint32(0))             // Compression (none)
	binary.Write(file, binary.LittleEndian, uint32(pixelDataSize)) // Image size
	binary.Write(file, binary.LittleEndian, int32(2835))           // Horizontal resolution (72 DPI)
	binary.Write(file, binary.LittleEndian, int32(2835))           // Vertical resolution (72 DPI)
	binary.Write(file, binary.LittleEndian, uint32(0))             // Colors in palette
	binary.Write(file, binary.LittleEndian, uint32(0))             // Important colors

	// Pixel data (BMP stores rows bottom-to-top and BGR format)
	padding := make([]byte, rowSize-width*3)
	for y := height - 1; y >= 0; y-- {
		for x := 0; x < width; x++ {
			i := (y*width + x) * 3
			// Convert RGB to BGR
			file.Write([]byte{data[i+2], data[i+1], data[i]})
		}
		file.Write(padding)
	}

	return nil
}
