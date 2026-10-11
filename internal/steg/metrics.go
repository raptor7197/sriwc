package steg

import (
	"fmt"
	"image"
	"math"
)

// gray converts an RGBA image to a float luminance plane (BT.601 weights).
func gray(img *image.RGBA) ([]float64, int, int) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	out := make([]float64, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			off := img.PixOffset(b.Min.X+x, b.Min.Y+y)
			out[y*w+x] = 0.299*float64(img.Pix[off]) + 0.587*float64(img.Pix[off+1]) + 0.114*float64(img.Pix[off+2])
		}
	}
	return out, w, h
}

func sameSize(a, b *image.RGBA) error {
	if a == nil || b == nil {
		return fmt.Errorf("nil image")
	}
	ab, bb := a.Bounds(), b.Bounds()
	if ab.Dx() != bb.Dx() || ab.Dy() != bb.Dy() {
		return fmt.Errorf("size mismatch: %dx%d vs %dx%d", ab.Dx(), ab.Dy(), bb.Dx(), bb.Dy())
	}
	return nil
}

// PSNR returns the peak signal-to-noise ratio in dB over the R, G and B channels.
// Identical images return 100 dB (a finite stand-in for +Inf).
func PSNR(a, b *image.RGBA) (float64, error) {
	if err := sameSize(a, b); err != nil {
		return 0, err
	}
	ab := a.Bounds()
	var sum float64
	n := 0
	for y := 0; y < ab.Dy(); y++ {
		for x := 0; x < ab.Dx(); x++ {
			ai := a.PixOffset(ab.Min.X+x, ab.Min.Y+y)
			bi := b.PixOffset(ab.Min.X+x, ab.Min.Y+y)
			for c := 0; c < 3; c++ {
				d := float64(a.Pix[ai+c]) - float64(b.Pix[bi+c])
				sum += d * d
				n++
			}
		}
	}
	if n == 0 {
		return 0, fmt.Errorf("empty image")
	}
	mse := sum / float64(n)
	if mse == 0 {
		return 100, nil
	}
	return 10 * math.Log10(255*255/mse), nil
}

// SSIM returns the mean structural similarity over non-overlapping 8x8 luminance windows.
func SSIM(a, b *image.RGBA) (float64, error) {
	if err := sameSize(a, b); err != nil {
		return 0, err
	}
	ga, w, h := gray(a)
	gb, _, _ := gray(b)
	const win = 8
	const c1 = 6.5025 // (0.01*255)^2
	const c2 = 58.5225
	var total float64
	var windows int
	for wy := 0; wy+win <= h; wy += win {
		for wx := 0; wx+win <= w; wx += win {
			var ma, mb float64
			for y := 0; y < win; y++ {
				for x := 0; x < win; x++ {
					i := (wy+y)*w + wx + x
					ma += ga[i]
					mb += gb[i]
				}
			}
			n := float64(win * win)
			ma /= n
			mb /= n
			var va, vb, cov float64
			for y := 0; y < win; y++ {
				for x := 0; x < win; x++ {
					i := (wy+y)*w + wx + x
					da := ga[i] - ma
					db := gb[i] - mb
					va += da * da
					vb += db * db
					cov += da * db
				}
			}
			va /= n - 1
			vb /= n - 1
			cov /= n - 1
			total += ((2*ma*mb + c1) * (2*cov + c2)) / ((ma*ma + mb*mb + c1) * (va + vb + c2))
			windows++
		}
	}
	if windows == 0 {
		return 0, fmt.Errorf("image too small for an %dx%d SSIM window", win, win)
	}
	return total / float64(windows), nil
}
