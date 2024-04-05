package pkg

import (
	"bufio"
	"fmt"
	"image"

	// "image/gif"
	"image/jpeg"
	"image/png"
)

// ExpectImage tries to read an image from the given reader and returns it.
func ExpectImage(reader *bufio.Reader) (image.Image, error) {
	var image image.Image
	header, err := reader.Peek(4)
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	if header[0] == 0x89 && header[1] == 0x50 {
		image, err = png.Decode(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to decode PNG image: %w", err)
		}
	} else if header[0] == 0xff && header[1] == 0xd8 {
		image, err = jpeg.Decode(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to decode JPEG image: %w", err)
		}
	} else {
		return nil, fmt.Errorf("unknown image format")
	}

	return image, nil
}
