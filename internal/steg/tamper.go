package steg

import (
	"fmt"
	"image"
	"math/rand"
)

// Tamper simulates a modification of the rectangular region (x, y, w, h) of img and returns
// the modified copy. mode is one of:
//
//	"black", "white" - fill the region with a flat value
//	"noise"          - replace the region with uniform random RGB noise (deterministic seed)
//	"copy"           - copy the same region from fill (falls back to black when fill is nil)
func Tamper(img *image.RGBA, x, y, w, h int, mode string, fill *image.RGBA) (*image.RGBA, error) {
	if img == nil {
		return nil, fmt.Errorf("nil image")
	}
	b := img.Bounds()
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("tamper region must have positive size, got %dx%d", w, h)
	}
	x0, y0 := x, y
	x1, y1 := x+w, y+h
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 > b.Dx() {
		x1 = b.Dx()
	}
	if y1 > b.Dy() {
		y1 = b.Dy()
	}
	if x0 >= x1 || y0 >= y1 {
		return nil, fmt.Errorf("tamper region (%d,%d,%dx%d) is outside image %dx%d", x, y, w, h, b.Dx(), b.Dy())
	}

	out := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	copy(out.Pix, img.Pix)

	switch mode {
	case "white":
		fillRegion(out, x0, y0, x1, y1, 255, 255, 255)
	case "copy":
		if fill == nil {
			fillRegion(out, x0, y0, x1, y1, 0, 0, 0)
			break
		}
		fb := fill.Bounds()
		for yy := y0; yy < y1; yy++ {
			for xx := x0; xx < x1; xx++ {
				sx, sy := xx%fb.Dx(), yy%fb.Dy()
				si := fill.PixOffset(fb.Min.X+sx, fb.Min.Y+sy)
				di := out.PixOffset(xx, yy)
				copy(out.Pix[di:di+4], fill.Pix[si:si+4])
			}
		}
	case "noise":
		rng := rand.New(rand.NewSource(int64(x0*7919 + y0*104729 + x1*31 + y1)))
		for yy := y0; yy < y1; yy++ {
			for xx := x0; xx < x1; xx++ {
				di := out.PixOffset(xx, yy)
				out.Pix[di] = uint8(rng.Intn(256))
				out.Pix[di+1] = uint8(rng.Intn(256))
				out.Pix[di+2] = uint8(rng.Intn(256))
				out.Pix[di+3] = 255
			}
		}
	case "black", "":
		fillRegion(out, x0, y0, x1, y1, 0, 0, 0)
	default:
		return nil, fmt.Errorf("unknown tamper mode %q (use black, white, noise or copy)", mode)
	}

	// Clamp-consistent region bounds for reporting.
	return out, nil
}

func fillRegion(img *image.RGBA, x0, y0, x1, y1 int, r, g, bl uint8) {
	for yy := y0; yy < y1; yy++ {
		for xx := x0; xx < x1; xx++ {
			di := img.PixOffset(xx, yy)
			img.Pix[di] = r
			img.Pix[di+1] = g
			img.Pix[di+2] = bl
			img.Pix[di+3] = 255
		}
	}
}

// Overlay tints each flagged block of img red (alpha 0.55) using rep.TamperMap. The block
// size and grid spacing come from the report.
func Overlay(img *image.RGBA, rep *Report) *image.RGBA {
	if img == nil || rep == nil {
		return img
	}
	b := img.Bounds()
	out := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	copy(out.Pix, img.Pix)
	bs := rep.BlockSize
	for iy := 0; iy < rep.BlocksY; iy++ {
		for ix := 0; ix < rep.BlocksX; ix++ {
			if !rep.TamperMap[iy*rep.BlocksX+ix] {
				continue
			}
			for yy := iy * bs; yy < (iy+1)*bs && yy < b.Dy(); yy++ {
				for xx := ix * bs; xx < (ix+1)*bs && xx < b.Dx(); xx++ {
					di := out.PixOffset(xx, yy)
					out.Pix[di] = uint8(0.45*float64(out.Pix[di]) + 0.55*255)
					out.Pix[di+1] = uint8(0.45 * float64(out.Pix[di+1]))
					out.Pix[di+2] = uint8(0.45 * float64(out.Pix[di+2]))
					out.Pix[di+3] = 255
				}
			}
		}
	}
	return out
}
