// Package imgio loads and stores images for the stegno pipeline. The pipeline is lossless:
// PNG in, PNG out. JPEG input is accepted (decoded to RGBA) but the watermark must then be
// written as PNG, so any data URL produced here is always image/png.
package imgio

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Decode converts any supported input into a non-premultiplied RGBA image with origin (0,0).
func Decode(r io.Reader) (*image.RGBA, string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, "", err
	}
	return DecodeBytes(data)
}

// DecodeBytes is Decode for an in-memory buffer. It returns the detected format
// ("png", "jpeg" or "unknown").
func DecodeBytes(data []byte) (*image.RGBA, string, error) {
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", fmt.Errorf("decode image: %w", err)
	}
	return ToRGBA(img), format, nil
}

// ToRGBA draws any image onto a fresh RGBA canvas with origin (0,0), dropping the source
// canvas offset and any premultiplication.
func ToRGBA(img image.Image) *image.RGBA {
	b := img.Bounds()
	out := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(out, out.Bounds(), img, b.Min, draw.Src)
	return out
}

// Load reads a file (PNG or JPEG) into RGBA.
func Load(path string) (*image.RGBA, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := Decode(f)
	return img, err
}

// EncodePNG writes img as PNG into a byte slice.
func EncodePNG(img *image.RGBA) ([]byte, error) {
	var buf bytes.Buffer
	enc := png.Encoder{CompressionLevel: png.DefaultCompression}
	if err := enc.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DataURL returns a "data:image/png;base64,..." string for img.
func DataURL(img *image.RGBA) (string, error) {
	raw, err := EncodePNG(img)
	if err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(raw), nil
}

// SavePNG writes img to path, creating parent directories as needed.
func SavePNG(path string, img *image.RGBA) error {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// SaveJPEG is provided for comparison experiments only; the watermark does not survive it.
func SaveJPEG(path string, img *image.RGBA, quality int) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, img, &jpeg.Options{Quality: quality})
}

// GetExtension reports the lowercase extension without the dot.
func GetExtension(path string) string {
	return strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
}
